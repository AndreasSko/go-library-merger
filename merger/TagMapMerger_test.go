package merger

import (
	"database/sql"
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/assert"
)

func TestMergeTagMaps(t *testing.T) {
	left := []*model.TagMap{
		{
			TagMapID:       1,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          1,
			Position:       0,
		},
		nil,
		{
			TagMapID:       3,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          2,
			Position:       0,
		},
		{
			TagMapID:       4,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 2, Valid: true},
			TagID:          2,
			Position:       1,
		},
		{
			TagMapID:       5,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          3,
			Position:       0,
		},
		nil,
		{
			TagMapID:       7,
			PlaylistItemID: sql.NullInt32{Int32: 1, Valid: true},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{},
			TagID:          5,
			Position:       0,
		},
		{
			TagMapID:       8,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 2, Valid: true},
			TagID:          3,
			Position:       1,
		},
		{
			TagMapID:       9,
			PlaylistItemID: sql.NullInt32{Int32: 1, Valid: true},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{},
			TagID:          3,
			Position:       2,
		},
		nil,
		{
			TagMapID:       11,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          100,
			Position:       0,
		},
		{
			TagMapID:       12,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          101,
			Position:       0,
		},
		{
			TagMapID:       13,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 999, Valid: true},
			TagID:          3,
			Position:       3,
		},
	}

	right := []*model.TagMap{
		// Duplicate
		{
			TagMapID:       1,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          1,
			Position:       0,
		},
		nil,
		// Duplicate
		{
			TagMapID:       3,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          2,
			Position:       0,
		},
		{
			TagMapID:       4,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 2, Valid: true},
			TagID:          8,
			Position:       0,
		},
		{
			TagMapID:       5,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          4,
			Position:       0,
		},
		nil,
		{
			TagMapID:       7,
			PlaylistItemID: sql.NullInt32{Int32: 1, Valid: true},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{},
			TagID:          1,
			Position:       1,
		},
		{
			TagMapID:       8,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 222, Valid: true},
			TagID:          3,
			Position:       0,
		},
		{
			TagMapID:       9,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 999, Valid: true},
			TagID:          3,
			Position:       1,
		},
	}

	expectedResult := []*model.TagMap{
		nil,
		{
			TagMapID:       1,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          1,
			Position:       0,
		},
		{
			TagMapID:       2,
			PlaylistItemID: sql.NullInt32{Int32: 1, Valid: true},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{},
			TagID:          1,
			Position:       1,
		},
		{
			TagMapID:       3,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          2,
			Position:       0,
		},
		{
			TagMapID:       4,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 2, Valid: true},
			TagID:          2,
			Position:       1,
		},
		{
			TagMapID:       5,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          3,
			Position:       0,
		},
		{
			TagMapID:       6,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 222, Valid: true},
			TagID:          3,
			Position:       1,
		},
		{
			TagMapID:       7,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 2, Valid: true},
			TagID:          3,
			Position:       2,
		},
		{
			TagMapID:       8,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 999, Valid: true},
			TagID:          3,
			Position:       3,
		},
		{
			TagMapID:       9,
			PlaylistItemID: sql.NullInt32{Int32: 1, Valid: true},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{},
			TagID:          3,
			Position:       4,
		},
		{
			TagMapID:       10,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          4,
			Position:       0,
		},
		{
			TagMapID:       11,
			PlaylistItemID: sql.NullInt32{Int32: 1, Valid: true},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{},
			TagID:          5,
			Position:       0,
		},
		{
			TagMapID:       12,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 2, Valid: true},
			TagID:          8,
			Position:       0,
		},
		{
			TagMapID:       13,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          100,
			Position:       0,
		},
		{
			TagMapID:       14,
			PlaylistItemID: sql.NullInt32{},
			LocationID:     sql.NullInt32{},
			NoteID:         sql.NullInt32{Int32: 1, Valid: true},
			TagID:          101,
			Position:       0,
		},
	}

	result, _, err := MergeTagMaps(left, right, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	// Check if original has not been tweaked
	assert.Equal(t, 12, left[11].TagMapID)
	assert.Equal(t, 8, right[7].TagMapID)

	assert.NotPanics(t, func() {
		MergeTagMaps(nil, nil, nil)
		MergeTagMaps([]*model.TagMap{}, []*model.TagMap{}, nil)
	})
}
