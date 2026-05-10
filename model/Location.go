package model

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
)

// Location represents the Location table inside the JW Library database
type Location struct {
	LocationID     int
	BookNumber     sql.NullInt32
	ChapterNumber  sql.NullInt32
	DocumentID     sql.NullInt32
	Track          sql.NullInt32
	IssueTagNumber int
	KeySymbol      sql.NullString
	MepsLanguage   sql.NullInt32
	LocationType   int
	Title          sql.NullString
	Specialty      sql.NullString
	Edition        sql.NullString
}

// ID returns the ID of the entry
func (m *Location) ID() int {
	return m.LocationID
}

// SetID sets the ID of the entry
func (m *Location) SetID(id int) {
	m.LocationID = id
}

// UniqueKey returns the key that makes this Location unique,
// so it can be used as a key in a map.
func (m *Location) UniqueKey() string {
	var sb strings.Builder
	sb.Grow(50)
	sb.WriteString(strconv.FormatInt(int64(m.BookNumber.Int32), 10))
	sb.WriteString("_")
	sb.WriteString(strconv.FormatInt(int64(m.ChapterNumber.Int32), 10))
	sb.WriteString("_")
	sb.WriteString(strconv.FormatInt(int64(m.DocumentID.Int32), 10))
	sb.WriteString("_")
	sb.WriteString(strconv.FormatInt(int64(m.Track.Int32), 10))
	sb.WriteString("_")
	sb.WriteString(strconv.FormatInt(int64(m.IssueTagNumber), 10))
	sb.WriteString("_")
	sb.WriteString(m.KeySymbol.String)
	sb.WriteString("_")
	if m.MepsLanguage.Valid {
		sb.WriteString(strconv.FormatInt(int64(m.MepsLanguage.Int32), 10))
	} else {
		sb.WriteString("!")
	}
	sb.WriteString("_")
	sb.WriteString(strconv.FormatInt(int64(m.LocationType), 10))
	sb.WriteString("_")
	if m.Specialty.Valid {
		sb.WriteString(m.Specialty.String)
	} else {
		sb.WriteString("!")
	}
	sb.WriteString("_")
	if m.Edition.Valid {
		sb.WriteString(m.Edition.String)
	} else {
		sb.WriteString("!")
	}
	return sb.String()
}

// Equals checks if the Location is equal to the given one.
func (m *Location) Equals(m2 Model) bool {
	if m2, ok := m2.(*Location); ok {
		return m.BookNumber.Int32 == m2.BookNumber.Int32 &&
			m.ChapterNumber.Int32 == m2.ChapterNumber.Int32 &&
			m.DocumentID.Int32 == m2.DocumentID.Int32 &&
			m.Track.Int32 == m2.Track.Int32 &&
			m.IssueTagNumber == m2.IssueTagNumber &&
			m.KeySymbol.String == m2.KeySymbol.String &&
			m.MepsLanguage == m2.MepsLanguage &&
			m.LocationType == m2.LocationType &&
			m.Specialty == m2.Specialty &&
			m.Edition == m2.Edition
	}
	return false
}

// RelatedEntries returns entries that are related to this one
func (m *Location) RelatedEntries(db *Database) Related {
	// We don't need it for now
	return Related{}
}

// PrettyPrint prints Location in a human readable format and
// adds information about related entries if helpful.
func (m *Location) PrettyPrint(db *Database) string {
	fields := []string{"Title", "BookNumber", "ChapterNumber", "DocumentID", "Track",
		"IssueTagNumber", "KeySymbol", "MepsLanguage"}
	return prettyPrint(m, fields)
}

// MarshalJSON returns the JSON encoding of the entry
func (m Location) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Type           string         `json:"type"`
		LocationID     int            `json:"locationId"`
		BookNumber     sql.NullInt32  `json:"bookNumber"`
		ChapterNumber  sql.NullInt32  `json:"chapterNumber"`
		DocumentID     sql.NullInt32  `json:"documentId"`
		Track          sql.NullInt32  `json:"track"`
		IssueTagNumber int            `json:"issueTagNumber"`
		KeySymbol      sql.NullString `json:"keySymbol"`
		MepsLanguage   sql.NullInt32  `json:"mepsLanguage"`
		LocationType   int            `json:"locationType"`
		Title          sql.NullString `json:"title"`
		Specialty      sql.NullString `json:"specialty"`
		Edition        sql.NullString `json:"edition"`
	}{
		Type:           "Location",
		LocationID:     m.LocationID,
		BookNumber:     m.BookNumber,
		ChapterNumber:  m.ChapterNumber,
		DocumentID:     m.DocumentID,
		Track:          m.Track,
		IssueTagNumber: m.IssueTagNumber,
		KeySymbol:      m.KeySymbol,
		MepsLanguage:   m.MepsLanguage,
		LocationType:   m.LocationType,
		Title:          m.Title,
		Specialty:      m.Specialty,
		Edition:        m.Edition,
	})
}

func (m *Location) tableName() string {
	return "Location"
}

func (m *Location) idName() string {
	return "LocationId"
}

func (m *Location) scanRow(rows *sql.Rows) (Model, error) {
	err := rows.Scan(&m.LocationID, &m.BookNumber, &m.ChapterNumber, &m.DocumentID, &m.Track,
		&m.IssueTagNumber, &m.KeySymbol, &m.MepsLanguage, &m.LocationType, &m.Title,
		&m.Specialty, &m.Edition)
	return m, err
}

// MakeSlice converts a slice of the generice interface model
func (Location) MakeSlice(mdl []Model) []*Location {
	result := make([]*Location, len(mdl))
	for i := range mdl {
		if mdl[i] != nil {
			result[i] = mdl[i].(*Location)
		}
	}
	return result
}
