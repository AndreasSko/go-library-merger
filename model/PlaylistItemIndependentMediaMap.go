package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// PlaylistItemIndependentMediaMap represents the PlaylistItemIndependentMediaMap table inside the JW Library database
type PlaylistItemIndependentMediaMap struct {
	PlaylistItemID     int
	IndependentMediaID int
	DurationTicks      int64
	pseudoID           int `ignore:"true"`
}

// ID returns a synthetic row ID; playlist and media IDs form the database key.
func (m *PlaylistItemIndependentMediaMap) ID() int {
	return m.pseudoID
}

// SetID changes only the synthetic row ID.
func (m *PlaylistItemIndependentMediaMap) SetID(id int) {
	m.pseudoID = id
}

// UniqueKey returns the key that makes this PlaylistItemIndependentMediaMap unique,
// so it can be used as a key in a map.
func (m *PlaylistItemIndependentMediaMap) UniqueKey() string {
	var sb strings.Builder
	sb.Grow(10)
	sb.WriteString(strconv.FormatInt(int64(m.PlaylistItemID), 10))
	sb.WriteString("_")
	sb.WriteString(strconv.FormatInt(int64(m.IndependentMediaID), 10))
	return sb.String()
}

// Equals checks if the PlaylistItemIndependentMediaMap is equal to the given one.
func (m *PlaylistItemIndependentMediaMap) Equals(m2 Model) bool {
	if m2, ok := m2.(*PlaylistItemIndependentMediaMap); ok {
		return m.PlaylistItemID == m2.PlaylistItemID &&
			m.IndependentMediaID == m2.IndependentMediaID &&
			m.DurationTicks == m2.DurationTicks
	}
	return false
}

// RelatedEntries returns entries that are related to this PlaylistItemIndependentMediaMap
func (m *PlaylistItemIndependentMediaMap) RelatedEntries(db *Database) Related {
	if db != nil {
		for _, item := range db.PlaylistItem {
			if item != nil && item.PlaylistItemID == m.PlaylistItemID {
				result := item.RelatedEntries(db)
				result.PlaylistItem = item
				return result
			}
		}
	}
	return Related{}
}

// PrettyPrint returns a string representation mainly for debugging purposes
func (m *PlaylistItemIndependentMediaMap) PrettyPrint(db *Database) string {
	return fmt.Sprintf("PlaylistItemIndependentMediaMap: PlaylistItem %d -> IndependentMedia %d", m.PlaylistItemID, m.IndependentMediaID)
}

// tableName returns the name of the table
func (m *PlaylistItemIndependentMediaMap) tableName() string {
	return "PlaylistItemIndependentMediaMap"
}

// idName returns the name of the ID field
func (m *PlaylistItemIndependentMediaMap) idName() string {
	return ""
}

// scanRow scans a database row into the PlaylistItemIndependentMediaMap struct
func (m *PlaylistItemIndependentMediaMap) scanRow(rows *sql.Rows) (Model, error) {
	result := PlaylistItemIndependentMediaMap{pseudoID: m.pseudoID}
	err := rows.Scan(
		&result.PlaylistItemID,
		&result.IndependentMediaID,
		&result.DurationTicks,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// MarshalJSON returns the JSON encoding
func (m *PlaylistItemIndependentMediaMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Type               string `json:"type"`
		PlaylistItemID     int    `json:"playlistItemId"`
		IndependentMediaID int    `json:"independentMediaId"`
		DurationTicks      int64  `json:"durationTicks"`
	}{
		Type:               "PlaylistItemIndependentMediaMap",
		PlaylistItemID:     m.PlaylistItemID,
		IndependentMediaID: m.IndependentMediaID,
		DurationTicks:      m.DurationTicks,
	})
}

// UnmarshalJSON parses the JSON encoding
func (m *PlaylistItemIndependentMediaMap) UnmarshalJSON(data []byte) error {
	aux := &struct {
		PlaylistItemID     int   `json:"PlaylistItemId"`
		IndependentMediaID int   `json:"IndependentMediaId"`
		DurationTicks      int64 `json:"DurationTicks"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	m.PlaylistItemID = aux.PlaylistItemID
	m.IndependentMediaID = aux.IndependentMediaID
	m.DurationTicks = aux.DurationTicks
	return nil
}

// MakeSlice converts a slice of the generic interface model
func (PlaylistItemIndependentMediaMap) MakeSlice(mdl []Model) []*PlaylistItemIndependentMediaMap {
	result := make([]*PlaylistItemIndependentMediaMap, len(mdl))
	for i := range mdl {
		if mdl[i] != nil {
			result[i] = mdl[i].(*PlaylistItemIndependentMediaMap)
		}
	}
	return result
}
