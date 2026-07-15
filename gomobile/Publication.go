package gomobile

import (
	"encoding/json"

	"github.com/AndreasSko/go-library-merger/publication"
)

// PublicationLookup represents a lookup for a publication.
// It directly maps to publication.Lookup
type PublicationLookup struct {
	DocumentID     int64
	KeySymbol      string
	IssueTagNumber int
	MepsLanguage   int
}

// LookupPublication looks up a publication from catalogDB located at dbPath
// and returns a JSON string representing the Publication
func LookupPublication(dbPath string, query *PublicationLookup) string {
	result, err := publication.LookupPublication(dbPath, publication.Lookup{
		DocumentID:     query.DocumentID,
		KeySymbol:      query.KeySymbol,
		IssueTagNumber: query.IssueTagNumber,
		MepsLanguage:   query.MepsLanguage,
	})
	if err != nil {
		return ""
	}

	jsn, err := json.Marshal(result)
	if err != nil {
		return ""
	}

	return string(jsn)
}
