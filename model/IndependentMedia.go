package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// IndependentMedia represents the IndependentMedia table inside the JW Library database
type IndependentMedia struct {
	IndependentMediaID  int
	OriginalFilename    string
	FilePath            string
	MimeType            string
	Hash                string
	validatedHash       string `ignore:"true"`
	validatedSourceHash string `ignore:"true"`
}

func (m *IndependentMedia) CopyFile(src, dst string) (resultErr error) {
	if m.FilePath == "." || m.FilePath == ".." || !filepath.IsLocal(m.FilePath) || filepath.Base(m.FilePath) != m.FilePath {
		return fmt.Errorf("invalid independent media path %q", m.FilePath)
	}
	f, err := os.Open(filepath.Join(src, m.FilePath))
	if err != nil {
		return fmt.Errorf("opening file %s in %s: %w", m.FilePath, src, err)
	}
	defer func() { resultErr = errors.Join(resultErr, f.Close()) }()

	df, err := os.Create(filepath.Join(dst, m.FilePath))
	if err != nil {
		return fmt.Errorf("creating destination file %s in %s: %w", m.FilePath, dst, err)
	}
	defer func() { resultErr = errors.Join(resultErr, df.Close()) }()

	_, err = io.Copy(df, f)
	if err != nil {
		return fmt.Errorf("copying file %s from %s to %s: %w", m.FilePath, src, dst, err)
	}

	return nil
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
// so it can be used as a key in a map.
func (m *IndependentMedia) UniqueKey() string {
	if m.validatedHash != "" && m.Hash == m.validatedSourceHash {
		return m.validatedHash
	}
	return m.Hash
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
