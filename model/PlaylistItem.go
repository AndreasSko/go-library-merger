package model

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
)

// PlaylistItem represents the PlaylistItem table inside the JW Library database
type PlaylistItem struct {
	PlaylistItemID       int
	Label                string
	StartTrimOffsetTicks sql.NullInt64
	EndTrimOffsetTicks   sql.NullInt64
	Accuracy             int
	EndAction            int
	ThumbnailFilePath    sql.NullString

	occurrence       int               `ignore:"true"`
	tag              *Tag              `ignore:"true"`
	independentMedia *IndependentMedia `ignore:"true"`
	thumbnail        *IndependentMedia `ignore:"true"`
	location         *Location         `ignore:"true"`
}

// ID returns the ID of the entry
func (m *PlaylistItem) ID() int {
	return m.PlaylistItemID
}

// SetID sets the ID of the entry
func (m *PlaylistItem) SetID(id int) {
	m.PlaylistItemID = id
}

// UniqueKey returns the key that makes this PlaylistItem unique,
// so it can be used as a key in a map.
func (m *PlaylistItem) UniqueKey() string {
	key := m.playlistContentKey()
	if key == "INVALID_KEY" {
		return key
	}
	return key + "_" + strconv.Itoa(m.occurrence)
}

func (m *PlaylistItem) playlistContentKey() string {
	var sb strings.Builder
	sb.Grow(50)
	if m.tag != nil {
		sb.WriteString(m.tag.UniqueKey())
		sb.WriteString("_")
	}
	if m.independentMedia != nil {
		sb.WriteString(m.independentMedia.UniqueKey())
	}
	if m.location != nil {
		sb.WriteString(m.location.UniqueKey())
	}

	if sb.Len() == 0 {
		return "INVALID_KEY"
	}
	return sb.String()
}

// Equals checks if the PlaylistItem is equal to the given one.
func (m *PlaylistItem) Equals(m2 Model) bool {
	if m2, ok := m2.(*PlaylistItem); ok {
		res := m.Label == m2.Label &&
			m.StartTrimOffsetTicks == m2.StartTrimOffsetTicks &&
			m.EndTrimOffsetTicks == m2.EndTrimOffsetTicks &&
			m.Accuracy == m2.Accuracy &&
			m.EndAction == m2.EndAction &&
			m.thumbnailContentKey() == m2.thumbnailContentKey() &&
			((m.tag == nil && m2.tag == nil) || (m.tag != nil && m2.tag != nil && m.tag.Equals(m2.tag))) &&
			((m.independentMedia == nil && m2.independentMedia == nil) || (m.independentMedia != nil && m2.independentMedia != nil && m.independentMedia.UniqueKey() == m2.independentMedia.UniqueKey())) &&
			((m.location == nil && m2.location == nil) || (m.location != nil && m2.location != nil && m.location.UniqueKey() == m2.location.UniqueKey()))
		return res
	}
	return false
}

// Thumbnail filenames are local to a backup and may change during media-path
// collision resolution. Compare the linked content instead of its filename.
// Unlinked models still compare their literal nullable path rather than losing
// a thumbnail change when used outside an imported Database.
func (m *PlaylistItem) thumbnailContentKey() string {
	if !m.ThumbnailFilePath.Valid {
		return "null"
	}
	if m.thumbnail != nil {
		return "hash:" + m.thumbnail.UniqueKey()
	}
	return "path:" + m.ThumbnailFilePath.String
}

// RelatedEntries returns entries that are related to this PlaylistItem
func (m *PlaylistItem) RelatedEntries(db *Database) Related {
	return Related{
		Tag:              m.tag,
		IndependentMedia: m.independentMedia,
		Location:         m.location,
	}
}

// PrettyPrint prints PlaylistItem in a human readable format and
// adds information about related entries if helpful.
func (m *PlaylistItem) PrettyPrint(db *Database) string {
	fields := []string{"Label", "StartTrimOffsetTicks", "EndTrimOffsetTicks", "Accuracy", "EndAction"}
	result := prettyPrint(m, fields)

	if m.tag != nil {
		result += "\n\n\nRelated Tag:\n"
		result += m.tag.PrettyPrint(db)
	}

	if m.independentMedia != nil {
		result += "\n\n\nRelated IndependentMedia:\n"
		result += m.independentMedia.PrettyPrint(db)
	}

	if m.location != nil {
		result += "\n\n\nRelated Location:\n"
		result += m.location.PrettyPrint(db)
	}

	return result
}

// tableName returns the name of the table
func (m *PlaylistItem) tableName() string {
	return "PlaylistItem"
}

// idName returns the name of the ID field
func (m *PlaylistItem) idName() string {
	return "PlaylistItemId"
}

// scanRow scans a database row into the PlaylistItem struct
func (m *PlaylistItem) scanRow(rows *sql.Rows) (Model, error) {
	var result PlaylistItem
	err := rows.Scan(
		&result.PlaylistItemID,
		&result.Label,
		&result.StartTrimOffsetTicks,
		&result.EndTrimOffsetTicks,
		&result.Accuracy,
		&result.EndAction,
		&result.ThumbnailFilePath,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// MarshalJSON returns the JSON encoding
func (m *PlaylistItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Type                 string         `json:"type"`
		PlaylistItemID       int            `json:"playlistItemId"`
		Label                string         `json:"label"`
		StartTrimOffsetTicks sql.NullInt64  `json:"startTrimOffsetTicks"`
		EndTrimOffsetTicks   sql.NullInt64  `json:"endTrimOffsetTicks"`
		Accuracy             int            `json:"accuracy"`
		EndAction            int            `json:"endAction"`
		ThumbnailFilePath    sql.NullString `json:"thumbnailFilePath"`
	}{
		Type:                 "PlaylistItem",
		PlaylistItemID:       m.PlaylistItemID,
		Label:                m.Label,
		StartTrimOffsetTicks: m.StartTrimOffsetTicks,
		EndTrimOffsetTicks:   m.EndTrimOffsetTicks,
		Accuracy:             m.Accuracy,
		EndAction:            m.EndAction,
		ThumbnailFilePath:    m.ThumbnailFilePath,
	})
}

// UnmarshalJSON parses the JSON encoding
func (m *PlaylistItem) UnmarshalJSON(data []byte) error {
	aux := &struct {
		PlaylistItemID       int            `json:"PlaylistItemId"`
		Label                string         `json:"Label"`
		StartTrimOffsetTicks sql.NullInt64  `json:"StartTrimOffsetTicks"`
		EndTrimOffsetTicks   sql.NullInt64  `json:"EndTrimOffsetTicks"`
		Accuracy             int            `json:"Accuracy"`
		EndAction            int            `json:"EndAction"`
		ThumbnailFilePath    sql.NullString `json:"ThumbnailFilePath"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	m.PlaylistItemID = aux.PlaylistItemID
	m.Label = aux.Label
	m.StartTrimOffsetTicks = aux.StartTrimOffsetTicks
	m.EndTrimOffsetTicks = aux.EndTrimOffsetTicks
	m.Accuracy = aux.Accuracy
	m.EndAction = aux.EndAction
	m.ThumbnailFilePath = aux.ThumbnailFilePath
	return nil
}

// MakeSlice converts a slice of the generic interface model
func (PlaylistItem) MakeSlice(mdl []Model) []*PlaylistItem {
	result := make([]*PlaylistItem, len(mdl))
	for i := range mdl {
		if mdl[i] != nil {
			result[i] = mdl[i].(*PlaylistItem)
		}
	}
	return result
}
