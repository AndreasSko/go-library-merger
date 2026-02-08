package merger

import (
	"github.com/AndreasSko/go-jwlm/model"
)

// DuplicateGroup represents a group of duplicate notes
type DuplicateGroup struct {
	Notes []*model.Note
}

// DetectDuplicateNotes scans through notes and groups duplicates together.
// Notes are considered duplicates if they have the same content and title.
func DetectDuplicateNotes(notes []*model.Note) []DuplicateGroup {
	// Map to track notes by their content signature
	contentMap := make(map[string][]*model.Note)

	for _, note := range notes {
		if note == nil {
			continue
		}

		// Create a signature based on title and content
		signature := createNoteSignature(note)
		contentMap[signature] = append(contentMap[signature], note)
	}

	// Extract only groups with duplicates (2+ notes)
	duplicates := []DuplicateGroup{}
	for _, group := range contentMap {
		if len(group) > 1 {
			duplicates = append(duplicates, DuplicateGroup{Notes: group})
		}
	}

	return duplicates
}

// createNoteSignature creates a unique signature for a note based on its content and title
func createNoteSignature(note *model.Note) string {
	title := ""
	if note.Title.Valid {
		title = note.Title.String
	}

	content := ""
	if note.Content.Valid {
		content = note.Content.String
	}

	// Combine title and content as the signature
	return title + "|" + content
}
