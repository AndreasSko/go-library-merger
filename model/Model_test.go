package model

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeModelCopy(t *testing.T) {
	br := &BlockRange{
		BlockRangeID: 1,
		BlockType:    1,
		Identifier:   1,
		StartToken:   sql.NullInt32{Int32: 1, Valid: true},
		EndToken:     sql.NullInt32{Int32: 2, Valid: true},
		UserMarkID:   1,
	}
	brCp := MakeModelCopy(br)
	assert.Equal(t, br, brCp)
	assert.NotSame(t, br, brCp)
	br.SetID(5)
	assert.Equal(t, 1, brCp.ID())

	bm := &Bookmark{
		BookmarkID:            1,
		LocationID:            2,
		PublicationLocationID: 3,
		Slot:                  4,
		Title:                 "Test",
		Snippet:               sql.NullString{},
		BlockType:             0,
		BlockIdentifier:       sql.NullInt32{},
	}
	bmCp := MakeModelCopy(bm)
	assert.Equal(t, bm, bmCp)
	assert.NotSame(t, bm, bmCp)
	bm.SetID(5)
	assert.Equal(t, 1, bmCp.ID())

	impf := &InputField{
		LocationID: 1,
		TextTag:    "2",
		Value:      "3",
		pseudoID:   4,
	}
	impfCp := MakeModelCopy(impf)
	assert.Equal(t, impf, impfCp)
	assert.NotSame(t, impf, impfCp)
	impf.SetID(1)
	assert.Equal(t, 4, impfCp.ID())

	loc := &Location{
		LocationID:     1,
		BookNumber:     sql.NullInt32{Int32: 2, Valid: true},
		ChapterNumber:  sql.NullInt32{Int32: 3, Valid: true},
		DocumentID:     sql.NullInt32{Int32: 4, Valid: true},
		Track:          sql.NullInt32{Int32: 5, Valid: true},
		IssueTagNumber: 6,
		KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
		MepsLanguage:   sql.NullInt32{Int32: 7, Valid: true},
		LocationType:   8,
		Title:          sql.NullString{String: "ThisTitleShouldNotBeInUniqueKey", Valid: true},
		Specialty:      sql.NullString{String: "E", Valid: true},
		Edition:        sql.NullString{String: "r2025", Valid: true},
	}
	locCp := MakeModelCopy(loc)
	assert.Equal(t, loc, locCp)
	assert.NotSame(t, loc, locCp)
	loc.SetID(5)
	assert.Equal(t, 1, locCp.ID())

	note := &Note{
		NoteID:          1,
		GUID:            "GUIDFOR1",
		UserMarkID:      sql.NullInt32{Int32: 1, Valid: true},
		LocationID:      sql.NullInt32{Int32: 1, Valid: true},
		Title:           sql.NullString{String: "A Title", Valid: true},
		Content:         sql.NullString{String: "The content", Valid: true},
		LastModified:    "2017-06-01T19:36:28+0200",
		Created:         "2016-06-01T19:36:28+0200",
		BlockType:       0,
		BlockIdentifier: sql.NullInt32{},
	}
	noteCp := MakeModelCopy(note)
	assert.Equal(t, note, noteCp)
	assert.NotSame(t, note, noteCp)
	note.SetID(5)
	assert.Equal(t, 1, noteCp.ID())

	tag := &Tag{
		TagID:   1,
		TagType: 1,
		Name:    "FirstTag",
	}
	tagCp := MakeModelCopy(tag)
	assert.Equal(t, tag, tagCp)
	assert.NotSame(t, tag, tagCp)
	tag.SetID(5)
	assert.Equal(t, 1, tagCp.ID())

	tm := &TagMap{
		TagMapID:       1,
		PlaylistItemID: sql.NullInt32{1, true},
		LocationID:     sql.NullInt32{1, true},
		NoteID:         sql.NullInt32{1, true},
		TagID:          1,
		Position:       1,
	}
	tmCp := MakeModelCopy(tm)
	assert.Equal(t, tm, tmCp)
	assert.NotSame(t, tm, tmCp)
	tm.SetID(5)
	assert.Equal(t, 1, tmCp.ID())

	um := &UserMark{
		UserMarkID:   1,
		ColorIndex:   1,
		LocationID:   1,
		StyleIndex:   1,
		UserMarkGUID: "FIRST",
		Version:      1,
	}
	umCp := MakeModelCopy(um)
	assert.Equal(t, um, umCp)
	assert.NotSame(t, um, umCp)
	um.SetID(5)
	assert.Equal(t, 1, umCp.ID())

	umbr := &UserMarkBlockRange{
		UserMark: &UserMark{
			UserMarkID:   1,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   1,
			UserMarkGUID: "FIRST",
			Version:      1,
		},
		BlockRanges: []*BlockRange{
			nil,
			{
				BlockRangeID: 1,
				BlockType:    1,
				Identifier:   1,
				StartToken:   sql.NullInt32{Int32: 1, Valid: true},
				EndToken:     sql.NullInt32{Int32: 2, Valid: true},
				UserMarkID:   1,
			},
		},
	}
	umbrCP := MakeModelCopy(umbr)
	assert.Equal(t, umbr, umbrCP)
	assert.NotSame(t, umbr, umbrCP)
	umbr.UserMark.SetID(5)
	umbr.BlockRanges[1].SetID(5)
	assert.Equal(t, 1, umbrCP.(*UserMarkBlockRange).UserMark.UserMarkID)
	assert.Equal(t, 1, umbrCP.(*UserMarkBlockRange).BlockRanges[1].BlockRangeID)
}

func TestPrettyPrint(t *testing.T) {
	location := &Location{
		LocationID: 1,
	}
	assert.PanicsWithValue(t, "Given struct does not contain field notexistent", func() {
		prettyPrint(location, []string{"notexistent"})
	})

	umbr := &UserMarkBlockRange{
		UserMark: &UserMark{},
	}

	assert.PanicsWithValue(t, "Unsupported type for field UserMark", func() {
		prettyPrint(umbr, []string{"UserMark"})
	})
}

func Test_sortByUniqueKey(t *testing.T) {
	locations := []*Location{
		nil,
		{
			LocationID: 1,
			BookNumber: sql.NullInt32{4, true},
		},
		{
			LocationID: 2,
			BookNumber: sql.NullInt32{3, true},
		},
		nil,
		{
			LocationID: 4,
			BookNumber: sql.NullInt32{2, true},
		},
		{
			LocationID: 5,
			BookNumber: sql.NullInt32{1, true},
		},
	}
	expectedLocations := []*Location{
		nil,
		{
			LocationID: 1,
			BookNumber: sql.NullInt32{1, true},
		},
		{
			LocationID: 2,
			BookNumber: sql.NullInt32{2, true},
		},
		{
			LocationID: 3,
			BookNumber: sql.NullInt32{3, true},
		},
		{
			LocationID: 4,
			BookNumber: sql.NullInt32{4, true},
		},
	}
	expectedLocIDChanges := map[int]int{
		1: 4,
		2: 3,
		4: 2,
		5: 1,
	}

	notes := []*Note{
		nil,
		{
			NoteID: 1,
			GUID:   "C",
		},
		{
			NoteID: 2,
			GUID:   "B",
		},
		{
			NoteID: 3,
			GUID:   "A",
		},
	}
	expectedNotes := []*Note{
		nil,
		{
			NoteID: 1,
			GUID:   "A",
		},
		{
			NoteID: 2,
			GUID:   "B",
		},
		{
			NoteID: 3,
			GUID:   "C",
		},
	}
	expectedNoteIDChanges := map[int]int{
		1: 3,
		3: 1,
	}

	locIDChanges := SortByUniqueKey(&locations)
	assert.Equal(t, expectedLocations, locations)
	assert.Equal(t, expectedLocIDChanges, locIDChanges)

	noteIDChanges := SortByUniqueKey(&notes)
	assert.Equal(t, expectedNotes, notes)
	assert.Equal(t, expectedNoteIDChanges, noteIDChanges)
}
