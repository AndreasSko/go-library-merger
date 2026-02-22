package merger

import (
	"database/sql"
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/assert"
)

func TestMergeNotes(t *testing.T) {
	// Merge without conflicts
	left := []*model.Note{
		{
			NoteID:          1,
			GUID:            "FirstGUID",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title", Valid: true},
			Content:         sql.NullString{String: "The content", Valid: true},
			LastModified:    "2017-06-01T20:36:28+0200",
			Created:         "2017-06-01T20:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          2,
			GUID:            "OnlyLeft",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title on the left", Valid: true},
			Content:         sql.NullString{String: "The content on the left", Valid: true},
			LastModified:    "2017-06-01T21:36:28+0200",
			Created:         "2017-06-01T21:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		nil,
	}
	right := []*model.Note{
		{
			NoteID:          1,
			GUID:            "-1stGUID",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A early Title", Valid: true},
			Content:         sql.NullString{String: "The early content", Valid: true},
			LastModified:    "2017-06-01T19:36:28+0200",
			Created:         "2017-06-01T19:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          2,
			GUID:            "FirstGUID",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title", Valid: true},
			Content:         sql.NullString{String: "The content", Valid: true},
			LastModified:    "2017-06-01T20:36:28+0200",
			Created:         "2017-06-01T20:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          3,
			GUID:            "OnlyRight",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title on the right", Valid: true},
			Content:         sql.NullString{String: "The content on the right", Valid: true},
			LastModified:    "2017-06-01T21:40:28+0200",
			Created:         "2017-06-01T21:40:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
	}

	expectedResult := []*model.Note{
		nil,
		{
			NoteID:          1,
			GUID:            "FirstGUID",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title", Valid: true},
			Content:         sql.NullString{String: "The content", Valid: true},
			LastModified:    "2017-06-01T20:36:28+0200",
			Created:         "2017-06-01T20:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          2,
			GUID:            "-1stGUID",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A early Title", Valid: true},
			Content:         sql.NullString{String: "The early content", Valid: true},
			LastModified:    "2017-06-01T19:36:28+0200",
			Created:         "2017-06-01T19:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          3,
			GUID:            "OnlyLeft",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title on the left", Valid: true},
			Content:         sql.NullString{String: "The content on the left", Valid: true},
			LastModified:    "2017-06-01T21:36:28+0200",
			Created:         "2017-06-01T21:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          4,
			GUID:            "OnlyRight",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title on the right", Valid: true},
			Content:         sql.NullString{String: "The content on the right", Valid: true},
			LastModified:    "2017-06-01T21:40:28+0200",
			Created:         "2017-06-01T21:40:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
	}

	expectedChanges := IDChanges{
		Left: map[int]int{
			2: 3,
		},
		Right: map[int]int{
			1: 2,
			2: 1,
			3: 4,
		},
	}

	result, changes, err := MergeNotes(left, right, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, expectedChanges, changes)
	// Check if original has not been tweaked
	assert.Equal(t, 2, left[1].NoteID)
	assert.Equal(t, 3, right[2].NoteID)

	// Call Merge while some entries have been updated => conflict
	left = []*model.Note{
		{
			NoteID:          1,
			GUID:            "FirstGUIDUpdating",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title", Valid: true},
			Content:         sql.NullString{String: "The content", Valid: true},
			LastModified:    "2017-06-01T20:36:28+0200",
			Created:         "2017-06-01T20:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          2,
			GUID:            "OnlyLeft",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title on the left", Valid: true},
			Content:         sql.NullString{String: "The content on the left", Valid: true},
			LastModified:    "2017-06-01T21:36:28+0200",
			Created:         "2017-06-01T21:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          3,
			GUID:            "AnotherUpdated",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "An updated Title", Valid: true},
			Content:         sql.NullString{String: "The content on the updated side", Valid: true},
			LastModified:    "2019-06-01T21:40:28+0200",
			Created:         "2018-06-01T21:40:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
	}
	right = []*model.Note{
		{
			NoteID:          1,
			GUID:            "-1stGUID",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A early Title", Valid: true},
			Content:         sql.NullString{String: "The early content", Valid: true},
			LastModified:    "2017-06-01T19:36:28+0200",
			Created:         "2017-06-01T19:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          2,
			GUID:            "FirstGUIDUpdating",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title that has been updated", Valid: true},
			Content:         sql.NullString{String: "The content is also updated", Valid: true},
			LastModified:    "2018-06-01T20:36:28+0200",
			Created:         "2017-06-01T20:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          3,
			GUID:            "OnlyRight",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title on the right", Valid: true},
			Content:         sql.NullString{String: "The content on the right", Valid: true},
			LastModified:    "2017-06-01T21:40:28+0200",
			Created:         "2017-06-01T21:40:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          4,
			GUID:            "AnotherUpdated",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A old title", Valid: true},
			Content:         sql.NullString{String: "The old content", Valid: true},
			LastModified:    "2018-06-01T21:40:28+0200",
			Created:         "2018-06-01T21:40:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
	}

	expectedCollisions := map[string]MergeConflict{
		"FirstGUIDUpdating": {
			Left: &model.Note{
				NoteID:          1,
				GUID:            "FirstGUIDUpdating",
				UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
				LocationID:      sql.NullInt32{Int32: 1, Valid: true},
				Title:           sql.NullString{String: "A Title", Valid: true},
				Content:         sql.NullString{String: "The content", Valid: true},
				LastModified:    "2017-06-01T20:36:28+0200",
				Created:         "2017-06-01T20:36:28+0200",
				BlockType:       0,
				BlockIdentifier: sql.NullInt32{},
			},
			Right: &model.Note{
				NoteID:          2,
				GUID:            "FirstGUIDUpdating",
				UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
				LocationID:      sql.NullInt32{Int32: 1, Valid: true},
				Title:           sql.NullString{String: "A Title that has been updated", Valid: true},
				Content:         sql.NullString{String: "The content is also updated", Valid: true},
				LastModified:    "2018-06-01T20:36:28+0200",
				Created:         "2017-06-01T20:36:28+0200",
				BlockType:       0,
				BlockIdentifier: sql.NullInt32{},
			},
		},
		"AnotherUpdated": {
			Left: &model.Note{
				NoteID:          3,
				GUID:            "AnotherUpdated",
				UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
				LocationID:      sql.NullInt32{Int32: 1, Valid: true},
				Title:           sql.NullString{String: "An updated Title", Valid: true},
				Content:         sql.NullString{String: "The content on the updated side", Valid: true},
				LastModified:    "2019-06-01T21:40:28+0200",
				Created:         "2018-06-01T21:40:28+0200",
				BlockType:       0,
				BlockIdentifier: sql.NullInt32{},
			},
			Right: &model.Note{
				NoteID:          4,
				GUID:            "AnotherUpdated",
				UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
				LocationID:      sql.NullInt32{Int32: 1, Valid: true},
				Title:           sql.NullString{String: "A old title", Valid: true},
				Content:         sql.NullString{String: "The old content", Valid: true},
				LastModified:    "2018-06-01T21:40:28+0200",
				Created:         "2018-06-01T21:40:28+0200",
				BlockType:       0,
				BlockIdentifier: sql.NullInt32{},
			},
		},
	}

	_, _, err = MergeNotes(left, right, nil)
	assert.Error(t, err)
	assert.Equal(t, expectedCollisions, err.(MergeConflictError).Conflicts)

	// Merge successfully with given conflict solution
	conflictSolution := map[string]MergeSolution{
		"FirstGUIDUpdating": {
			Side: RightSide,
			Solution: &model.Note{
				NoteID:          2,
				GUID:            "FirstGUIDUpdating",
				UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
				LocationID:      sql.NullInt32{Int32: 1, Valid: true},
				Title:           sql.NullString{String: "A Title that has been updated", Valid: true},
				Content:         sql.NullString{String: "The content is also updated", Valid: true},
				LastModified:    "2018-06-01T20:36:28+0200",
				Created:         "2017-06-01T20:36:28+0200",
				BlockType:       0,
				BlockIdentifier: sql.NullInt32{},
			},
			Discarded: &model.Note{
				NoteID:          1,
				GUID:            "FirstGUIDUpdating",
				UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
				LocationID:      sql.NullInt32{Int32: 1, Valid: true},
				Title:           sql.NullString{String: "A Title", Valid: true},
				Content:         sql.NullString{String: "The content", Valid: true},
				LastModified:    "2017-06-01T20:36:28+0200",
				Created:         "2017-06-01T20:36:28+0200",
				BlockType:       0,
				BlockIdentifier: sql.NullInt32{},
			},
		},
		"AnotherUpdated": {
			Side: LeftSide,
			Solution: &model.Note{
				NoteID:          3,
				GUID:            "AnotherUpdated",
				UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
				LocationID:      sql.NullInt32{Int32: 1, Valid: true},
				Title:           sql.NullString{String: "An updated Title", Valid: true},
				Content:         sql.NullString{String: "The content on the updated side", Valid: true},
				LastModified:    "2019-06-01T21:40:28+0200",
				Created:         "2018-06-01T21:40:28+0200",
				BlockType:       0,
				BlockIdentifier: sql.NullInt32{},
			},
			Discarded: &model.Note{
				NoteID:          4,
				GUID:            "AnotherUpdated",
				UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
				LocationID:      sql.NullInt32{Int32: 1, Valid: true},
				Title:           sql.NullString{String: "A old title", Valid: true},
				Content:         sql.NullString{String: "The old content", Valid: true},
				LastModified:    "2018-06-01T21:40:28+0200",
				Created:         "2018-06-01T21:40:28+0200",
				BlockType:       0,
				BlockIdentifier: sql.NullInt32{},
			},
		},
	}

	expectedResult = []*model.Note{
		nil,
		{
			NoteID:          1,
			GUID:            "-1stGUID",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A early Title", Valid: true},
			Content:         sql.NullString{String: "The early content", Valid: true},
			LastModified:    "2017-06-01T19:36:28+0200",
			Created:         "2017-06-01T19:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          2,
			GUID:            "OnlyLeft",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title on the left", Valid: true},
			Content:         sql.NullString{String: "The content on the left", Valid: true},
			LastModified:    "2017-06-01T21:36:28+0200",
			Created:         "2017-06-01T21:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          3,
			GUID:            "FirstGUIDUpdating",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title that has been updated", Valid: true},
			Content:         sql.NullString{String: "The content is also updated", Valid: true},
			LastModified:    "2018-06-01T20:36:28+0200",
			Created:         "2017-06-01T20:36:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          4,
			GUID:            "AnotherUpdated",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "An updated Title", Valid: true},
			Content:         sql.NullString{String: "The content on the updated side", Valid: true},
			LastModified:    "2019-06-01T21:40:28+0200",
			Created:         "2018-06-01T21:40:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
		{
			NoteID:          5,
			GUID:            "OnlyRight",
			UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
			LocationID:      sql.NullInt32{Int32: 1, Valid: true},
			Title:           sql.NullString{String: "A Title on the right", Valid: true},
			Content:         sql.NullString{String: "The content on the right", Valid: true},
			LastModified:    "2017-06-01T21:40:28+0200",
			Created:         "2017-06-01T21:40:28+0200",
			BlockType:       0,
			BlockIdentifier: sql.NullInt32{},
		},
	}

	expectedChanges = IDChanges{
		Left: map[int]int{
			1: 3,
			3: 4,
		},
		Right: map[int]int{
			2: 3,
			3: 5,
		},
	}

	result, changes, err = MergeNotes(left, right, conflictSolution)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, expectedChanges, changes)
}
