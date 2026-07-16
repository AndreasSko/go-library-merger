package publication

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"

	// Register SQLite driver
	_ "github.com/mattn/go-sqlite3"
)

// Publication represents a publication with all
// its information from the catalogDB
type Publication struct {
	ID                    int
	PublicationRootKeyID  int
	MepsLanguageID        int
	PublicationTypeID     int
	IssueTagNumber        int
	Title                 string
	IssueTitle            sql.NullString
	ShortTitle            string
	CoverTitle            sql.NullString
	UndatedTitle          sql.NullString
	UndatedReferenceTitle sql.NullString
	Year                  int
	Symbol                string
	KeySymbol             sql.NullString
	Reserved              int
}

// Lookup represents a lookup for a publication.
// This query can contain various fields.
type Lookup struct {
	DocumentID     int64
	KeySymbol      string
	IssueTagNumber int
	MepsLanguage   int
}

// LookupPublication looks up a publication from catalogDB located at dbPath
func LookupPublication(dbPath string, query Lookup) (Publication, error) {
	// Check if file exists
	if _, err := os.Stat(dbPath); err != nil {
		return Publication{}, fmt.Errorf("CatalogDB does not exist at %s", dbPath)
	}

	db, err := sql.Open("sqlite3", dbPath+"?immutable=1")
	if err != nil {
		return Publication{}, errors.Wrap(err, "Error while opening SQLite database")
	}
	defer db.Close()

	return lookupPublication(db, query)
}

func lookupPublication(db *sql.DB, query Lookup) (Publication, error) {
	var row *sql.Row
	if query.DocumentID != 0 {
		stmt, err := db.Prepare("SELECT P.* " +
			"FROM Publication AS P, PublicationDocument AS PD " +
			"WHERE P.Id = PD.PublicationId AND PD.DocumentId = ? AND P.MepsLanguageId = ?")
		if err != nil {
			return Publication{}, errors.Wrap(err, "Error while preparing query")
		}
		row = stmt.QueryRow(query.DocumentID, query.MepsLanguage)
	} else {
		stmt, err := db.Prepare("SELECT * FROM Publication WHERE KeySymbol = ? AND MepsLanguageId = ? AND IssueTagNumber = ?")
		if err != nil {
			return Publication{}, errors.Wrap(err, "Error while preparing query")
		}
		row = stmt.QueryRow(query.KeySymbol, query.MepsLanguage, query.IssueTagNumber)
	}

	publ := Publication{}
	err := row.Scan(&publ.PublicationRootKeyID,
		&publ.MepsLanguageID,
		&publ.PublicationTypeID,
		&publ.IssueTagNumber,
		&publ.Title,
		&publ.IssueTitle,
		&publ.ShortTitle,
		&publ.CoverTitle,
		&publ.UndatedTitle,
		&publ.UndatedReferenceTitle,
		&publ.Year,
		&publ.Symbol,
		&publ.KeySymbol,
		&publ.Reserved,
		&publ.ID)
	if err != nil {
		return Publication{}, errors.Wrap(err, "Error while scanning row for publication")
	}

	return publ, nil
}

// MarshalJSON returns the JSON encoding of the entry
func (m Publication) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID                    int    `json:"id"`
		PublicationRootKeyID  int    `json:"publicationRootKeyId"`
		MepsLanguageID        int    `json:"mepsLanguageId"`
		PublicationTypeID     int    `json:"publicationTypeId"`
		IssueTagNumber        int    `json:"issueTagNumber"`
		Title                 string `json:"title"`
		IssueTitle            string `json:"issueTitle"`
		ShortTitle            string `json:"shortTitle"`
		CoverTitle            string `json:"coverTitle"`
		UndatedTitle          string `json:"undatedTitle"`
		UndatedReferenceTitle string `json:"undatedReferenceTitle"`
		Year                  int    `json:"year"`
		Symbol                string `json:"symbol"`
		KeySymbol             string `json:"keySymbol"`
		Reserved              int    `json:"reserved"`
	}{
		ID:                    m.ID,
		PublicationRootKeyID:  m.PublicationRootKeyID,
		MepsLanguageID:        m.MepsLanguageID,
		PublicationTypeID:     m.PublicationTypeID,
		IssueTagNumber:        m.IssueTagNumber,
		Title:                 m.Title,
		IssueTitle:            m.IssueTitle.String,
		ShortTitle:            m.ShortTitle,
		CoverTitle:            m.CoverTitle.String,
		UndatedTitle:          m.UndatedTitle.String,
		UndatedReferenceTitle: m.UndatedReferenceTitle.String,
		Year:                  m.Year,
		Symbol:                m.Symbol,
		KeySymbol:             m.KeySymbol.String,
		Reserved:              m.Reserved,
	})
}
