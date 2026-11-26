package model

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// IndependentMedia represents the IndependentMedia table inside the JW Library database
type IndependentMedia struct {
	IndependentMediaID int
	OriginalFilename   string
	FilePath           string
	MimeType           string
	Hash               string
	contentHash        string `ignore:"true"`
	fileMissing        bool   `ignore:"true"`
}

// ID returns the ID of the entry
func (m *IndependentMedia) ID() int {
	return m.IndependentMediaID
}

// SetID sets the ID of the entry
func (m *IndependentMedia) SetID(id int) {
	m.IndependentMediaID = id
}

// UniqueKey returns the key that makes this IndependentMedia unique,
// so it can be used as a key in a map. After import, it is the standard
// format of the file hash.
func (m *IndependentMedia) UniqueKey() string {
	if m.contentHash != "" {
		return m.contentHash
	}
	return m.Hash
}

// FileMissing reports whether the backup did not include the file of this
// media. JW Library sometimes writes such backups.
func (m *IndependentMedia) FileMissing() bool {
	return m.fileMissing
}

// SetFileMissing sets whether the file of this media is missing, for
// example after the merge copied it from the other backup.
func (m *IndependentMedia) SetFileMissing(missing bool) {
	m.fileMissing = missing
}

// MatchMissingMedia gives media whose file is missing the key of the same
// media in one of dbs that has the file, so the merge sees them as the same
// media. Without the file, the hash in the database can't be converted to
// the standard format, but it can be compared with the hash of a file.
// A key that another row of the same backup already has is not given, since
// the merge keeps only one row per key and backup.
func MatchMissingMedia(dbs ...*Database) {
	keys := map[string]string{}
	for _, db := range dbs {
		for _, im := range db.IndependentMedia {
			if im == nil || im.fileMissing || im.contentHash == "" {
				continue
			}
			keys[im.contentHash] = im.contentHash
			if digest, err := hex.DecodeString(im.contentHash); err == nil {
				keys[nativeHash(digest)] = im.contentHash
			}
		}
	}
	for _, db := range dbs {
		// Missing rows with the same hash already share a key, so they form a
		// group that gets one key. Keys are compared ignoring case, since the
		// new key is in lower case.
		groups := map[string][]*IndependentMedia{}
		owners := map[string][]string{}
		for _, im := range db.IndependentMedia {
			if im == nil {
				continue
			}
			group := "\x00present"
			if im.fileMissing {
				group = im.Hash
				groups[group] = append(groups[group], im)
			}
			key := strings.ToLower(im.UniqueKey())
			owners[key] = append(owners[key], group)
		}
		// Sorted, so the result doesn't depend on map order when two groups
		// match the same key.
		for _, hash := range slices.Sorted(maps.Keys(groups)) {
			group := groups[hash]
			key, ok := keys[strings.ToLower(hash)]
			if !ok || slices.ContainsFunc(owners[key], func(owner string) bool { return owner != hash }) {
				continue
			}
			owners[key] = append(owners[key], hash)
			for _, im := range group {
				im.contentHash = key
			}
		}
	}
}

// nativeHash formats a digest like JW Library does: each byte as hex without
// a leading zero ("a" instead of "0a").
func nativeHash(digest []byte) string {
	var native strings.Builder
	for _, value := range digest {
		fmt.Fprintf(&native, "%x", value)
	}
	return native.String()
}

// Equals checks if the IndependentMedia is equal to the given one.
func (m *IndependentMedia) Equals(m2 Model) bool {
	if m2, ok := m2.(*IndependentMedia); ok {
		return m.OriginalFilename == m2.OriginalFilename &&
			m.FilePath == m2.FilePath &&
			m.MimeType == m2.MimeType &&
			m.Hash == m2.Hash
	}
	return false
}

// RelatedEntries returns entries that are related to this IndependentMedia
func (m *IndependentMedia) RelatedEntries(db *Database) Related {
	result := Related{}
	return result
}

// PrettyPrint returns a string representation mainly for debugging purposes
func (m *IndependentMedia) PrettyPrint(db *Database) string {
	return prettyPrint(m, []string{"OriginalFilename", "FilePath", "MimeType", "Hash"})
}

// tableName returns the name of the table
func (m *IndependentMedia) tableName() string {
	return "IndependentMedia"
}

// idName returns the name of the ID field
func (m *IndependentMedia) idName() string {
	return "IndependentMediaId"
}

// scanRow scans a database row into the IndependentMedia struct
func (m *IndependentMedia) scanRow(rows *sql.Rows) (Model, error) {
	var result IndependentMedia
	err := rows.Scan(
		&result.IndependentMediaID,
		&result.OriginalFilename,
		&result.FilePath,
		&result.MimeType,
		&result.Hash,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// MarshalJSON returns the JSON encoding
func (m *IndependentMedia) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Type               string `json:"type"`
		IndependentMediaID int    `json:"independentMediaId"`
		OriginalFilename   string `json:"originalFilename"`
		FilePath           string `json:"filePath"`
		MimeType           string `json:"mimeType"`
		Hash               string `json:"hash"`
	}{
		Type:               "IndependentMedia",
		IndependentMediaID: m.IndependentMediaID,
		OriginalFilename:   m.OriginalFilename,
		FilePath:           m.FilePath,
		MimeType:           m.MimeType,
		Hash:               m.Hash,
	})
}

// UnmarshalJSON parses the JSON encoding
func (m *IndependentMedia) UnmarshalJSON(data []byte) error {
	aux := &struct {
		IndependentMediaID int    `json:"IndependentMediaId"`
		OriginalFilename   string `json:"OriginalFilename"`
		FilePath           string `json:"FilePath"`
		MimeType           string `json:"MimeType"`
		Hash               string `json:"Hash"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	m.IndependentMediaID = aux.IndependentMediaID
	m.OriginalFilename = aux.OriginalFilename
	m.FilePath = aux.FilePath
	m.MimeType = aux.MimeType
	m.Hash = aux.Hash
	return nil
}

// MakeSlice converts a slice of the generic interface model
func (IndependentMedia) MakeSlice(mdl []Model) []*IndependentMedia {
	result := make([]*IndependentMedia, len(mdl))
	for i := range mdl {
		if mdl[i] != nil {
			result[i] = mdl[i].(*IndependentMedia)
		}
	}
	return result
}
