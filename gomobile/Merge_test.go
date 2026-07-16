//go:build !windows
// +build !windows

package gomobile

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/assert"
)

func Test_MergeMultiCollisionAllRight(t *testing.T) {
	dbw := DatabaseWrapper{
		left:  model.MakeDatabaseCopy(leftMultiCollision),
		right: model.MakeDatabaseCopy(rightMultiCollision),
	}
	dbw.Init()

	mcw := &MergeConflictsWrapper{}
	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeBookmarks("", mcw))

	assert.Error(t, dbw.MergeInputFields("", mcw))
	conflict, err := mcw.NextConflict()
	assert.NoError(t, err)
	assert.NoError(t, mcw.SolveConflict(conflict.Key, "rightSide"))
	assert.NoError(t, dbw.MergeInputFields("", mcw))

	assert.NoError(t, dbw.MergeTags())

	assert.Error(t, dbw.MergeUserMarkAndBlockRange("", mcw))
	conflict, err = mcw.NextConflict()
	assert.NoError(t, err)
	assert.NoError(t, mcw.SolveConflict(conflict.Key, "rightSide"))

	assert.Error(t, dbw.MergeUserMarkAndBlockRange("", mcw))
	conflict, err = mcw.NextConflict()
	assert.NoError(t, err)
	assert.NoError(t, mcw.SolveConflict(conflict.Key, "rightSide"))

	assert.Error(t, dbw.MergeUserMarkAndBlockRange("", mcw))
	conflict, err = mcw.NextConflict()
	assert.NoError(t, err)
	assert.NoError(t, mcw.SolveConflict(conflict.Key, "rightSide"))

	assert.Error(t, dbw.MergeUserMarkAndBlockRange("", mcw))
	conflict, err = mcw.NextConflict()
	assert.NoError(t, err)
	assert.NoError(t, mcw.SolveConflict(conflict.Key, "rightSide"))

	assert.NoError(t, dbw.MergeUserMarkAndBlockRange("", mcw))
	assert.NoError(t, dbw.MergeNotes("", mcw))
	assert.NoError(t, dbw.MergeTagMaps())

	assert.True(t, dbw.merged.Equals(rightMultiCollision))
}

func Test_MergeMultiCollisionAllExceptOneRight(t *testing.T) {
	// Test multiple times so we are sure that a pass is not only a coincidence
	// (the order of maps change and with this we can check if we depend on it somewhere)
	for i := 0; i < 10; i++ {
		dbw := DatabaseWrapper{
			left:  model.MakeDatabaseCopy(leftMultiCollision),
			right: model.MakeDatabaseCopy(rightMultiCollision),
		}
		dbw.Init()

		mcw := &MergeConflictsWrapper{}
		assert.NoError(t, dbw.MergeLocations())
		assert.NoError(t, dbw.MergeBookmarks("", mcw))

		assert.Error(t, dbw.MergeInputFields("", mcw))
		conflict, err := mcw.NextConflict()
		assert.NoError(t, err)
		assert.NoError(t, mcw.SolveConflict(conflict.Key, "rightSide"))
		assert.NoError(t, dbw.MergeInputFields("", mcw))

		assert.NoError(t, dbw.MergeTags())

		assert.Error(t, dbw.MergeUserMarkAndBlockRange("", mcw))
		conflict, err = mcw.NextConflict()
		assert.NoError(t, err)
		assert.NoError(t, mcw.SolveConflict(conflict.Key, "rightSide"))

		assert.Error(t, dbw.MergeUserMarkAndBlockRange("", mcw))
		conflict, err = mcw.NextConflict()
		assert.NoError(t, err)
		assert.NoError(t, mcw.SolveConflict(conflict.Key, "rightSide"))

		assert.Error(t, dbw.MergeUserMarkAndBlockRange("", mcw))
		conflict, err = mcw.NextConflict()
		assert.NoError(t, err)
		assert.NoError(t, mcw.SolveConflict(conflict.Key, "rightSide"))

		assert.Error(t, dbw.MergeUserMarkAndBlockRange("", mcw))
		conflict, err = mcw.NextConflict()
		assert.NoError(t, err)
		assert.NoError(t, mcw.SolveConflict(conflict.Key, "leftSide"))

		assert.NoError(t, dbw.MergeUserMarkAndBlockRange("", mcw))
		assert.NoError(t, dbw.MergeNotes("", mcw))
		assert.NoError(t, dbw.MergeTagMaps())

		expected := model.MakeDatabaseCopy(rightMultiCollision)
		expected.BlockRange = []*model.BlockRange{
			nil,
			{
				BlockRangeID: 1,
				BlockType:    1,
				Identifier:   1,
				StartToken:   sql.NullInt32{3, true},
				EndToken:     sql.NullInt32{3, true},
				UserMarkID:   1,
			},
		}

		assert.True(t, dbw.merged.Equals(expected))
	}
}

func Test_MergeMultiCollisionAutoSolver(t *testing.T) {
	dbw := DatabaseWrapper{
		left:  model.MakeDatabaseCopy(leftMultiCollision),
		right: model.MakeDatabaseCopy(rightMultiCollision),
	}
	dbw.Init()

	mcw := &MergeConflictsWrapper{}
	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeInputFields("chooseRight", mcw))
	assert.NoError(t, dbw.MergeBookmarks("", mcw))
	assert.NoError(t, dbw.MergeTags())
	assert.NoError(t, dbw.MergeUserMarkAndBlockRange("chooseRight", mcw))
	assert.NoError(t, dbw.MergeNotes("", mcw))
	assert.NoError(t, dbw.MergeTagMaps())

	assert.True(t, dbw.merged.Equals(rightMultiCollision))
}

func Test_MergeNwt(t *testing.T) {
	dbw := DatabaseWrapper{
		left:  model.MakeDatabaseCopy(leftNwtDB),
		right: model.MakeDatabaseCopy(rightNwtDB),
	}
	dbw.Init()

	mcw := &MergeConflictsWrapper{}
	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeInputFields("chooseLeft", mcw))
	assert.NoError(t, dbw.MergeBookmarks("chooseLeft", mcw))
	assert.NoError(t, dbw.MergeTags())
	assert.NoError(t, dbw.MergeUserMarkAndBlockRange("chooseLeft", mcw))
	assert.NoError(t, dbw.MergeNotes("chooseLeft", mcw))
	assert.NoError(t, dbw.MergeTagMaps())

	assert.True(t, dbw.merged.Equals(mergedAllLeftNwtDB))
}

func Test_MergeNwtDifferentDocID(t *testing.T) {
	// As the cleanup happens in the PostMerge hook, which is called on export,
	// we also need to export the merged DB
	tmp := t.TempDir()

	dbw := DatabaseWrapper{
		left:  model.MakeDatabaseCopy(leftDBNwtWithDifferentDocID),
		right: model.MakeDatabaseCopy(rightDBNwtWithDifferentDocID),
	}
	dbw.Init()

	mcw := &MergeConflictsWrapper{}
	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeInputFields("chooseLeft", mcw))
	assert.NoError(t, dbw.MergeBookmarks("chooseLeft", mcw))
	assert.NoError(t, dbw.MergeTags())
	assert.NoError(t, dbw.MergeUserMarkAndBlockRange("chooseLeft", mcw))
	assert.NoError(t, dbw.MergeNotes("chooseLeft", mcw))
	assert.NoError(t, dbw.MergeTagMaps())

	newBackup := filepath.Join(tmp, "test.jwlibrary")
	assert.NoError(t, dbw.ExportMerged(newBackup))

	newDB := &model.Database{}
	assert.NoError(t, newDB.ImportJWLBackup(newBackup))

	assert.True(t, newDB.Equals(mergedDBNwtWithDifferentDocID))
}

var leftMultiCollision = &model.Database{
	BlockRange: []*model.BlockRange{
		nil,
		{
			BlockRangeID: 1,
			BlockType:    1,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{0, true},
			UserMarkID:   1,
		},
		{
			BlockRangeID: 2,
			BlockType:    1,
			Identifier:   1,
			StartToken:   sql.NullInt32{1, true},
			EndToken:     sql.NullInt32{1, true},
			UserMarkID:   2,
		},
		{
			BlockRangeID: 3,
			BlockType:    1,
			Identifier:   1,
			StartToken:   sql.NullInt32{2, true},
			EndToken:     sql.NullInt32{2, true},
			UserMarkID:   3,
		},
		{
			BlockRangeID: 4,
			BlockType:    1,
			Identifier:   1,
			StartToken:   sql.NullInt32{3, true},
			EndToken:     sql.NullInt32{3, true},
			UserMarkID:   4,
		},
	},
	Bookmark: []*model.Bookmark{nil},
	InputField: []*model.InputField{
		nil,
		{
			LocationID: 1,
			TextTag:    "a1",
			Value:      "a1",
		},
		{
			LocationID: 1,
			TextTag:    "a2",
			Value:      "a2",
		},
	},
	Location: []*model.Location{
		nil,
		{
			LocationID:    1,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{1, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 1", true},
		},
	},
	Note:   []*model.Note{nil},
	Tag:    []*model.Tag{nil},
	TagMap: []*model.TagMap{nil},
	UserMark: []*model.UserMark{
		nil,
		{
			UserMarkID:   1,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   1,
			UserMarkGUID: "1",
		},
		{
			UserMarkID:   2,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   1,
			UserMarkGUID: "2",
		},
		{
			UserMarkID:   3,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   1,
			UserMarkGUID: "3",
		},
		{
			UserMarkID:   4,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   1,
			UserMarkGUID: "4",
		},
	},
}

var rightMultiCollision = &model.Database{
	BlockRange: []*model.BlockRange{
		nil,
		{
			BlockRangeID: 1,
			BlockType:    1,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{20, true},
			UserMarkID:   1,
		},
	},
	Bookmark: []*model.Bookmark{nil},
	InputField: []*model.InputField{
		nil,
		{
			LocationID: 1,
			TextTag:    "a1",
			Value:      "different",
		},
		{
			LocationID: 1,
			TextTag:    "a2",
			Value:      "a2",
		},
		{
			LocationID: 1,
			TextTag:    "b1",
			Value:      "b1",
		},
	},
	Location: []*model.Location{
		nil,
		{
			LocationID:    1,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{1, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 1", true},
		},
	},
	Note:   []*model.Note{nil},
	Tag:    []*model.Tag{nil},
	TagMap: []*model.TagMap{nil},
	UserMark: []*model.UserMark{
		nil,
		{
			UserMarkID:   1,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   1,
			UserMarkGUID: "1R",
		},
	},
}

// These tests are more or less copied from cmd/merge_test.go

// Merge against empty DB and see if result is still the same
func Test_MergeWithEmpty(t *testing.T) {
	dbw := DatabaseWrapper{
		left:  model.MakeDatabaseCopy(leftDB),
		right: model.MakeDatabaseCopy(emptyDB),
	}
	dbw.Init()

	mcw := &MergeConflictsWrapper{}

	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeInputFields("", mcw))
	assert.NoError(t, dbw.MergeBookmarks("", mcw))
	assert.NoError(t, dbw.MergeTags())
	assert.NoError(t, dbw.MergeUserMarkAndBlockRange("", mcw))
	assert.NoError(t, dbw.MergeNotes("", mcw))
	assert.NoError(t, dbw.MergeTagMaps())

	assert.True(t, dbw.left.Equals(dbw.merged))
}

// Merge while selecting all right
func Test_MergeAllRight(t *testing.T) {
	dbw := DatabaseWrapper{
		left:  model.MakeDatabaseCopy(leftDB),
		right: model.MakeDatabaseCopy(rightDB),
	}
	dbw.Init()

	mcw := &MergeConflictsWrapper{}

	assert.NoError(t, dbw.MergeLocations())
	assert.Error(t, dbw.MergeBookmarks("", mcw))
	selectSameSide(mcw, "rightSide")

	dbw.Init()
	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeBookmarks("", mcw))
	assert.Error(t, dbw.MergeInputFields("", mcw))
	selectSameSide(mcw, "rightSide")

	dbw.Init()
	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeBookmarks("", mcw))
	assert.NoError(t, dbw.MergeInputFields("", mcw))
	assert.NoError(t, dbw.MergeTags())
	assert.Error(t, dbw.MergeUserMarkAndBlockRange("", mcw))
	selectSameSide(mcw, "rightSide")

	dbw.Init()
	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeBookmarks("", mcw))
	assert.NoError(t, dbw.MergeInputFields("", mcw))
	assert.NoError(t, dbw.MergeTags())
	assert.NoError(t, dbw.MergeUserMarkAndBlockRange("", mcw))
	assert.Error(t, dbw.MergeNotes("", mcw))
	selectSameSide(mcw, "rightSide")

	dbw.Init()
	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeBookmarks("", mcw))
	assert.NoError(t, dbw.MergeInputFields("", mcw))
	assert.NoError(t, dbw.MergeTags())
	assert.NoError(t, dbw.MergeUserMarkAndBlockRange("", mcw))
	assert.NoError(t, dbw.MergeNotes("", mcw))
	assert.NoError(t, dbw.MergeTagMaps())

	assert.True(t, mergedAllRightDB.Equals(dbw.merged))
}

// Merge while selecting all right
func Test_MergeAllLeft(t *testing.T) {
	dbw := DatabaseWrapper{
		left:  model.MakeDatabaseCopy(leftDB),
		right: model.MakeDatabaseCopy(rightDB),
	}
	dbw.Init()

	mcw := &MergeConflictsWrapper{}

	assert.NoError(t, dbw.MergeLocations())
	assert.Error(t, dbw.MergeBookmarks("", mcw))
	selectSameSide(mcw, "leftSide")

	dbw.Init()
	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeBookmarks("", mcw))
	assert.Error(t, dbw.MergeInputFields("", mcw))
	selectSameSide(mcw, "leftSide")

	dbw.Init()
	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeBookmarks("", mcw))
	assert.NoError(t, dbw.MergeInputFields("", mcw))
	assert.NoError(t, dbw.MergeTags())
	assert.Error(t, dbw.MergeUserMarkAndBlockRange("", mcw))
	selectSameSide(mcw, "leftSide")

	dbw.Init()
	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeBookmarks("", mcw))
	assert.NoError(t, dbw.MergeInputFields("", mcw))
	assert.NoError(t, dbw.MergeTags())
	assert.NoError(t, dbw.MergeUserMarkAndBlockRange("", mcw))
	assert.Error(t, dbw.MergeNotes("", mcw))
	selectSameSide(mcw, "leftSide")

	dbw.Init()
	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeBookmarks("", mcw))
	assert.NoError(t, dbw.MergeInputFields("", mcw))
	assert.NoError(t, dbw.MergeTags())
	assert.NoError(t, dbw.MergeUserMarkAndBlockRange("", mcw))
	assert.NoError(t, dbw.MergeNotes("", mcw))
	assert.NoError(t, dbw.MergeTagMaps())

	assert.True(t, mergedAllLeftDB.Equals(dbw.merged))
}

// Merge with auto resolution: chooseRight for Bookmarks & Markings & InputFields,
// chooseNewest for Notes
func Test_MergeWithAutoresolution(t *testing.T) {
	dbw := DatabaseWrapper{
		left:  model.MakeDatabaseCopy(leftDB),
		right: model.MakeDatabaseCopy(rightDB),
	}
	dbw.Init()

	mcw := &MergeConflictsWrapper{}

	assert.NoError(t, dbw.MergeLocations())
	assert.NoError(t, dbw.MergeBookmarks("chooseRight", mcw))
	assert.NoError(t, dbw.MergeInputFields("chooseRight", mcw))
	assert.NoError(t, dbw.MergeTags())
	assert.NoError(t, dbw.MergeUserMarkAndBlockRange("chooseRight", mcw))
	assert.NoError(t, dbw.MergeNotes("chooseNewest", mcw))
	assert.NoError(t, dbw.MergeTagMaps())

	assert.True(t, mergedAllRightDB.Equals(dbw.merged))
}

func selectSameSide(mcw *MergeConflictsWrapper, side string) {
	for {
		conflict, err := mcw.NextConflict()
		if err != nil {
			break
		}
		mcw.SolveConflict(conflict.Key, side)
	}
}

var emptyDB = &model.Database{}

var leftDB = &model.Database{
	BlockRange: []*model.BlockRange{
		nil,
		{
			BlockRangeID: 1,
			BlockType:    2,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{7, true},
			UserMarkID:   1,
		},
	},
	Bookmark: []*model.Bookmark{
		nil,
		{
			BookmarkID:            1,
			LocationID:            1,
			PublicationLocationID: 2,
			Title:                 "1. Mose 1:1",
			Snippet:               sql.NullString{"1 Am Anfang erschuf Gott Himmel und Erde.", true},
			BlockType:             2,
			BlockIdentifier:       sql.NullInt32{1, true},
		},
	},
	InputField: []*model.InputField{
		nil,
		{
			LocationID: 5,
			TextTag:    "a1",
			Value:      "a1",
		},
		{
			LocationID: 5,
			TextTag:    "a2",
			Value:      "a2",
		},
	},
	Location: []*model.Location{
		nil,
		{
			LocationID:    1,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{1, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 1", true},
		},
		{
			LocationID:   2,
			KeySymbol:    sql.NullString{"nwtsty", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 1,
		},
		nil,
		nil,
		{
			LocationID:   5,
			DocumentID:   sql.NullInt64{1102021811, true},
			KeySymbol:    sql.NullString{"lffi", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 0,
		},
	},
	Note: []*model.Note{
		nil,
		{
			NoteID:          1,
			GUID:            "92B192B4-B0B9-49B2-949F-7A8BA6406AF4",
			UserMarkID:      sql.NullInt32{1, true},
			LocationID:      sql.NullInt32{1, true},
			Title:           sql.NullString{"Am Anfang erschuf Gott Himmel und Erde.", true},
			Content:         sql.NullString{"📝 for left version", true},
			LastModified:    "2020-09-15T13:45:38+00:00",
			Created:         "2020-09-15T13:45:38+00:00",
			BlockType:       2,
			BlockIdentifier: sql.NullInt32{1, true},
		},
		{
			NoteID:       2,
			GUID:         "E36B34A0-B70F-4590-9D69-5887AB65A6D5",
			Title:        sql.NullString{"Same Note", true},
			Content:      sql.NullString{"This note is also available on the other side", true},
			LastModified: "2020-09-15T13:52:25+00:00",
			Created:      "2020-09-15T13:52:25+00:00",
			BlockType:    0,
		},
	},
	Tag: []*model.Tag{
		nil,
		{
			TagID:   1,
			TagType: 0,
			Name:    "Favorite",
		},
		{
			TagID:   2,
			TagType: 1,
			Name:    "Left",
		},
		{
			TagID:   3,
			TagType: 1,
			Name:    "Same",
		},
	},
	TagMap: []*model.TagMap{
		nil,
		{
			TagMapID: 1,
			NoteID:   sql.NullInt32{1, true},
			TagID:    2,
			Position: 0,
		},
		{
			TagMapID: 2,
			NoteID:   sql.NullInt32{2, true},
			TagID:    3,
			Position: 0,
		},
	},
	UserMark: []*model.UserMark{
		nil,
		{
			UserMarkID:   1,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "0D5523D9-F784-4B08-A86D-D517F4EB17DE",
			Version:      1,
		},
	},
}

var rightDB = &model.Database{
	BlockRange: []*model.BlockRange{
		nil,
		{
			BlockRangeID: 1,
			BlockType:    2,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{15, true},
			UserMarkID:   1,
		},
		{
			BlockRangeID: 2,
			BlockType:    2,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{7, true},
			UserMarkID:   2,
		},
		{
			BlockRangeID: 3,
			BlockType:    2,
			Identifier:   2,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{12, true},
			UserMarkID:   2,
		},
		{
			BlockRangeID: 4,
			BlockType:    2,
			Identifier:   2,
			StartToken:   sql.NullInt32{13, true},
			EndToken:     sql.NullInt32{26, true},
			UserMarkID:   3,
		},
	},
	Bookmark: []*model.Bookmark{
		nil,
		{
			BookmarkID:            1,
			LocationID:            1,
			PublicationLocationID: 2,
			Title:                 "1. Mose 2:1",
			Snippet:               sql.NullString{"2 So wurde die Erschaffung von Himmel und Erde und allem, was dazugehört, beendet. ", true},
			BlockType:             2,
			BlockIdentifier:       sql.NullInt32{1, true},
		},
	},
	InputField: []*model.InputField{
		nil,
		{
			LocationID: 4,
			TextTag:    "a1",
			Value:      "different",
		},
		{
			LocationID: 4,
			TextTag:    "a2",
			Value:      "a2",
		},
		{
			LocationID: 4,
			TextTag:    "b1",
			Value:      "b1",
		},
	},
	Location: []*model.Location{
		nil,
		{
			LocationID:    1,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{2, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 2", true},
		},
		{
			LocationID:   2,
			KeySymbol:    sql.NullString{"nwtsty", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 1,
		},
		{
			LocationID:    3,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{1, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 1", true},
		},
		{
			LocationID:   4,
			DocumentID:   sql.NullInt64{1102021811, true},
			KeySymbol:    sql.NullString{"lffi", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 0,
		},
	},
	Note: []*model.Note{
		nil,
		{
			NoteID:          1,
			GUID:            "DE4A2CDA-9892-4A94-AF4B-22EBE05A05CA",
			UserMarkID:      sql.NullInt32{1, true},
			LocationID:      sql.NullInt32{1, true},
			Title:           sql.NullString{"So wurde die Erschaffung von Himmel und Erde und allem, was dazugehört, beendet.", true},
			Content:         sql.NullString{"📝 on the right side", true},
			LastModified:    "2020-09-15T13:47:56+00:00",
			Created:         "2020-09-15T13:47:56+00:00",
			BlockType:       2,
			BlockIdentifier: sql.NullInt32{1, true},
		},
		{
			NoteID:       2,
			GUID:         "E36B34A0-B70F-4590-9D69-5887AB65A6D5",
			Title:        sql.NullString{"Same Note", true},
			Content:      sql.NullString{"This note is also available on the other side. Though this one is newer 😏", true},
			LastModified: "2020-09-20T13:52:25+00:00",
			Created:      "2020-09-15T13:52:25+00:00",
			BlockType:    0,
		},
	},
	Tag: []*model.Tag{
		nil,
		{
			TagID:   1,
			TagType: 0,
			Name:    "Favorite",
		},
		{
			TagID:   2,
			TagType: 1,
			Name:    "Right",
		},
		{
			TagID:   3,
			TagType: 1,
			Name:    "Same",
		},
	},
	TagMap: []*model.TagMap{
		nil,
		{
			TagMapID: 1,
			NoteID:   sql.NullInt32{1, true},
			TagID:    2,
			Position: 0,
		},
		{
			TagMapID: 2,
			NoteID:   sql.NullInt32{2, true},
			TagID:    3,
			// Should normally be 0, but changed it to 1 to detect
			// if merger recognizes that its still the same entry,
			// just with a different position
			Position: 1,
		},
	},
	UserMark: []*model.UserMark{
		nil,
		{
			UserMarkID:   1,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "54C23825-AC1E-4890-8041-92B39C5E4B17",
			Version:      1,
		},
		{
			UserMarkID:   2,
			ColorIndex:   1,
			LocationID:   3,
			StyleIndex:   0,
			UserMarkGUID: "A098D2B0-7613-4676-9E96-442755105D9A",
		},
		{
			UserMarkID:   3,
			ColorIndex:   1,
			LocationID:   3,
			StyleIndex:   0,
			UserMarkGUID: "A86DECC8-69B1-4A43-A3A1-F1CA7B8E8388",
			Version:      1,
		},
	},
}

var mergedAllLeftDB = &model.Database{
	BlockRange: []*model.BlockRange{
		nil,
		{
			BlockRangeID: 1,
			BlockType:    2,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{7, true},
			UserMarkID:   1,
		},
		{
			BlockRangeID: 2,
			BlockType:    2,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{15, true},
			UserMarkID:   2,
		},
		{
			BlockRangeID: 3,
			BlockType:    2,
			Identifier:   2,
			StartToken:   sql.NullInt32{13, true},
			EndToken:     sql.NullInt32{26, true},
			UserMarkID:   3,
		},
	},
	Bookmark: []*model.Bookmark{
		nil,
		{
			BookmarkID:            1,
			LocationID:            1,
			PublicationLocationID: 2,
			Title:                 "1. Mose 1:1",
			Snippet:               sql.NullString{"1 Am Anfang erschuf Gott Himmel und Erde.", true},
			BlockType:             2,
			BlockIdentifier:       sql.NullInt32{1, true},
		},
	},
	InputField: []*model.InputField{
		nil,
		{
			LocationID: 4,
			TextTag:    "a1",
			Value:      "a1",
		},
		{
			LocationID: 4,
			TextTag:    "a2",
			Value:      "a2",
		},
		{
			LocationID: 4,
			TextTag:    "b1",
			Value:      "b1",
		},
	},
	Location: []*model.Location{
		nil,
		{
			LocationID:    1,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{1, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 1", true},
		},
		{
			LocationID:   2,
			KeySymbol:    sql.NullString{"nwtsty", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 1,
		},
		{
			LocationID:    3,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{2, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 2", true},
		},
		{
			LocationID:   4,
			DocumentID:   sql.NullInt64{1102021811, true},
			KeySymbol:    sql.NullString{"lffi", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 0,
		},
	},
	Note: []*model.Note{
		nil,
		{
			NoteID:          1,
			GUID:            "92B192B4-B0B9-49B2-949F-7A8BA6406AF4",
			UserMarkID:      sql.NullInt32{1, true},
			LocationID:      sql.NullInt32{1, true},
			Title:           sql.NullString{"Am Anfang erschuf Gott Himmel und Erde.", true},
			Content:         sql.NullString{"📝 for left version", true},
			LastModified:    "2020-09-15T13:45:38+00:00",
			Created:         "2020-09-15T13:45:38+00:00",
			BlockType:       2,
			BlockIdentifier: sql.NullInt32{1, true},
		},
		{
			NoteID:       2,
			GUID:         "E36B34A0-B70F-4590-9D69-5887AB65A6D5",
			Title:        sql.NullString{"Same Note", true},
			Content:      sql.NullString{"This note is also available on the other side", true},
			LastModified: "2020-09-15T13:52:25+00:00",
			Created:      "2020-09-15T13:52:25+00:00",
			BlockType:    0,
		},
		{
			NoteID:          3,
			GUID:            "DE4A2CDA-9892-4A94-AF4B-22EBE05A05CA",
			UserMarkID:      sql.NullInt32{1, true},
			LocationID:      sql.NullInt32{1, true},
			Title:           sql.NullString{"So wurde die Erschaffung von Himmel und Erde und allem, was dazugehört, beendet.", true},
			Content:         sql.NullString{"📝 on the right side", true},
			LastModified:    "2020-09-15T13:47:56+00:00",
			Created:         "2020-09-15T13:47:56+00:00",
			BlockType:       2,
			BlockIdentifier: sql.NullInt32{1, true},
		},
	},
	Tag: []*model.Tag{
		nil,
		{
			TagID:   1,
			TagType: 0,
			Name:    "Favorite",
		},
		{
			TagID:   2,
			TagType: 1,
			Name:    "Left",
		},
		{
			TagID:   3,
			TagType: 1,
			Name:    "Same",
		},
		{
			TagID:   4,
			TagType: 1,
			Name:    "Right",
		},
	},
	TagMap: []*model.TagMap{
		nil,
		{
			TagMapID: 1,
			NoteID:   sql.NullInt32{1, true},
			TagID:    2,
			Position: 0,
		},
		{
			TagMapID: 2,
			NoteID:   sql.NullInt32{2, true},
			TagID:    3,
			Position: 0,
		},
		{
			TagMapID: 2,
			NoteID:   sql.NullInt32{3, true},
			TagID:    4,
			Position: 0,
		},
	},
	UserMark: []*model.UserMark{
		nil,
		{
			UserMarkID:   1,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "0D5523D9-F784-4B08-A86D-D517F4EB17DE",
			Version:      1,
		},
		{
			UserMarkID:   2,
			ColorIndex:   1,
			LocationID:   3,
			StyleIndex:   0,
			UserMarkGUID: "54C23825-AC1E-4890-8041-92B39C5E4B17",
			Version:      1,
		},
		{
			UserMarkID:   3,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "A86DECC-69B1-4A43-A3A1-F1CA7B8E8388",
			Version:      1,
		},
	},
}

var mergedAllRightDB = &model.Database{
	BlockRange: []*model.BlockRange{
		nil,
		{
			BlockRangeID: 1,
			BlockType:    2,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{15, true},
			UserMarkID:   1,
		},
		{
			BlockRangeID: 2,
			BlockType:    2,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{7, true},
			UserMarkID:   2,
		},
		{
			BlockRangeID: 3,
			BlockType:    2,
			Identifier:   2,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{12, true},
			UserMarkID:   2,
		},
		{
			BlockRangeID: 4,
			BlockType:    2,
			Identifier:   2,
			StartToken:   sql.NullInt32{13, true},
			EndToken:     sql.NullInt32{26, true},
			UserMarkID:   3,
		},
	},
	Bookmark: []*model.Bookmark{
		nil,
		{
			BookmarkID:            1,
			LocationID:            3,
			PublicationLocationID: 2,
			Title:                 "1. Mose 2:1",
			Snippet:               sql.NullString{"2 So wurde die Erschaffung von Himmel und Erde und allem, was dazugehört, beendet. ", true},
			BlockType:             2,
			BlockIdentifier:       sql.NullInt32{1, true},
		},
	},
	InputField: []*model.InputField{
		nil,
		{
			LocationID: 4,
			TextTag:    "a1",
			Value:      "different",
		},
		{
			LocationID: 4,
			TextTag:    "a2",
			Value:      "a2",
		},
		{
			LocationID: 4,
			TextTag:    "b1",
			Value:      "b1",
		},
	},
	Location: []*model.Location{
		nil,
		{
			LocationID:    1,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{1, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 1", true},
		},
		{
			LocationID:   2,
			KeySymbol:    sql.NullString{"nwtsty", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 1,
		},
		{
			LocationID:    3,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{2, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 2", true},
		},
		{
			LocationID:   4,
			DocumentID:   sql.NullInt64{1102021811, true},
			KeySymbol:    sql.NullString{"lffi", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 0,
		},
	},
	Note: []*model.Note{
		nil,
		{
			NoteID:          1,
			GUID:            "92B192B4-B0B9-49B2-949F-7A8BA6406AF4",
			UserMarkID:      sql.NullInt32{1, true},
			LocationID:      sql.NullInt32{1, true},
			Title:           sql.NullString{"Am Anfang erschuf Gott Himmel und Erde.", true},
			Content:         sql.NullString{"📝 for left version", true},
			LastModified:    "2020-09-15T13:45:38+00:00",
			Created:         "2020-09-15T13:45:38+00:00",
			BlockType:       2,
			BlockIdentifier: sql.NullInt32{1, true},
		},
		{
			NoteID:       2,
			GUID:         "E36B34A0-B70F-4590-9D69-5887AB65A6D5",
			Title:        sql.NullString{"Same Note", true},
			Content:      sql.NullString{"This note is also available on the other side. Though this one is newer 😏", true},
			LastModified: "2020-09-20T13:52:25+00:00",
			Created:      "2020-09-20T13:52:25+00:00",
			BlockType:    0,
		},
		{
			NoteID:          3,
			GUID:            "DE4A2CDA-9892-4A94-AF4B-22EBE05A05CA",
			UserMarkID:      sql.NullInt32{1, true},
			LocationID:      sql.NullInt32{1, true},
			Title:           sql.NullString{"So wurde die Erschaffung von Himmel und Erde und allem, was dazugehört, beendet.", true},
			Content:         sql.NullString{"📝 on the right side", true},
			LastModified:    "2020-09-15T13:47:56+00:00",
			Created:         "2020-09-15T13:47:56+00:00",
			BlockType:       2,
			BlockIdentifier: sql.NullInt32{1, true},
		},
	},
	Tag: []*model.Tag{
		nil,
		{
			TagID:   1,
			TagType: 0,
			Name:    "Favorite",
		},
		{
			TagID:   2,
			TagType: 1,
			Name:    "Left",
		},
		{
			TagID:   3,
			TagType: 1,
			Name:    "Same",
		},
		{
			TagID:   4,
			TagType: 1,
			Name:    "Right",
		},
	},
	TagMap: []*model.TagMap{
		nil,
		{
			TagMapID: 1,
			NoteID:   sql.NullInt32{1, true},
			TagID:    2,
			Position: 0,
		},
		{
			TagMapID: 2,
			NoteID:   sql.NullInt32{2, true},
			TagID:    3,
			Position: 0,
		},
		{
			TagMapID: 1,
			NoteID:   sql.NullInt32{3, true},
			TagID:    4,
			Position: 0,
		},
	},
	UserMark: []*model.UserMark{
		nil,
		{
			UserMarkID:   1,
			ColorIndex:   1,
			LocationID:   3, // 1. Mose 2
			StyleIndex:   0,
			UserMarkGUID: "54C23825-AC1E-4890-8041-92B39C5E4B17",
			Version:      1,
		},
		{
			UserMarkID:   2,
			ColorIndex:   1,
			LocationID:   1, // 1. Mose 1
			StyleIndex:   0,
			UserMarkGUID: "A098D2B0-7613-4676-9E96-442755105D9A",
		},
		{
			UserMarkID:   3,
			ColorIndex:   1,
			LocationID:   1, // 1. Mose 1
			StyleIndex:   0,
			UserMarkGUID: "A86DECC8-69B1-4A43-A3A1-F1CA7B8E8388",
			Version:      1,
		},
	},
}

var leftNwtDB = &model.Database{
	BlockRange: []*model.BlockRange{
		nil,
		{
			BlockRangeID: 1,
			BlockType:    1,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{7, true},
			UserMarkID:   1,
		},
		{
			BlockRangeID: 2,
			BlockType:    1,
			Identifier:   2,
			StartToken:   sql.NullInt32{20, true},
			EndToken:     sql.NullInt32{30, true},
			UserMarkID:   3,
		},
	},
	Location: []*model.Location{
		nil,
		{
			LocationID:    1,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{1, true},
			KeySymbol:     sql.NullString{"nwt", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 1", true},
		},
		{
			LocationID:   2,
			KeySymbol:    sql.NullString{"nwt", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 1,
		},
		nil,
		nil,
		{
			LocationID:   5,
			DocumentID:   sql.NullInt64{1102021811, true},
			KeySymbol:    sql.NullString{"lffi", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 0,
		},
	},
	UserMark: []*model.UserMark{
		nil,
		{
			UserMarkID:   1,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "1",
			Version:      1,
		},
		nil,
		{
			UserMarkID:   3,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "3",
			Version:      1,
		},
	},
}

var rightNwtDB = &model.Database{
	BlockRange: []*model.BlockRange{
		nil,
		{
			BlockRangeID: 1,
			BlockType:    1,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{7, true},
			UserMarkID:   1,
		},
		{
			BlockRangeID: 2,
			BlockType:    1,
			Identifier:   20,
			StartToken:   sql.NullInt32{5, true},
			EndToken:     sql.NullInt32{10, true},
			UserMarkID:   2,
		},
	},
	Location: []*model.Location{
		nil,
		{
			LocationID:    1,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{2, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 2", true},
		},
		{
			LocationID:   2,
			KeySymbol:    sql.NullString{"nwtsty", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 1,
		},
		{
			LocationID:    3,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{1, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 1", true},
		},
		{
			LocationID:   4,
			DocumentID:   sql.NullInt64{1102021811, true},
			KeySymbol:    sql.NullString{"lffi", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 0,
		},
	},
	UserMark: []*model.UserMark{
		nil,
		{
			UserMarkID:   1,
			ColorIndex:   1,
			LocationID:   3,
			StyleIndex:   0,
			UserMarkGUID: "1",
			Version:      1,
		},
		{
			UserMarkID:   2,
			ColorIndex:   1,
			LocationID:   3,
			StyleIndex:   0,
			UserMarkGUID: "2",
			Version:      1,
		},
	},
}

var mergedAllLeftNwtDB = &model.Database{
	BlockRange: []*model.BlockRange{
		nil,
		{
			BlockRangeID: 1,
			BlockType:    1,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{7, true},
			UserMarkID:   1,
		},
		{
			BlockRangeID: 2,
			BlockType:    1,
			Identifier:   2,
			StartToken:   sql.NullInt32{20, true},
			EndToken:     sql.NullInt32{30, true},
			UserMarkID:   2,
		},
		{
			BlockRangeID: 3,
			BlockType:    1,
			Identifier:   20,
			StartToken:   sql.NullInt32{5, true},
			EndToken:     sql.NullInt32{10, true},
			UserMarkID:   3,
		},
	},
	Bookmark:   []*model.Bookmark{nil},
	InputField: []*model.InputField{nil},
	Location: []*model.Location{
		nil,
		{
			LocationID:    1,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{1, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 1", true},
		},
		{
			LocationID:   2,
			KeySymbol:    sql.NullString{"nwtsty", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 1,
		},
		{
			LocationID:   3,
			DocumentID:   sql.NullInt64{1102021811, true},
			KeySymbol:    sql.NullString{"lffi", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
			LocationType: 0,
		},
		{
			LocationID:    4,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{2, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
			LocationType:  0,
			Title:         sql.NullString{"1. Mose 2", true},
		},
	},
	Note:   []*model.Note{nil},
	Tag:    []*model.Tag{nil},
	TagMap: []*model.TagMap{nil},
	UserMark: []*model.UserMark{
		nil,
		{
			UserMarkID:   1,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "1",
			Version:      1,
		},
		{
			UserMarkID:   2,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "3",
			Version:      1,
		},
		{
			UserMarkID:   3,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "2",
			Version:      1,
		},
	},
}

var leftDBNwtWithDifferentDocID = &model.Database{
	BlockRange: []*model.BlockRange{
		nil,
		{
			BlockRangeID: 1,
			BlockType:    2,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{7, true},
			UserMarkID:   2,
		},
		{
			BlockRangeID: 2,
			BlockType:    2,
			Identifier:   5,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{20, true},
			UserMarkID:   3,
		},
		{
			BlockRangeID: 3,
			BlockType:    2,
			Identifier:   2,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{5, true},
			UserMarkID:   2,
		},
	},
	Bookmark:   []*model.Bookmark{nil},
	InputField: []*model.InputField{nil},
	Location: []*model.Location{
		nil,
		{
			LocationID:   1,
			DocumentID:   sql.NullInt64{1102021811, true},
			KeySymbol:    sql.NullString{"lffi", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
		},
		{
			LocationID:   2,
			DocumentID:   sql.NullInt64{123456789, true},
			KeySymbol:    sql.NullString{"nwtsty", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
		},
		// Make sure that when migrating to nwtsty we don't have duplicate locations
		// in case the nwtsty-location already exists on the other side
		{
			LocationID:    3,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{1, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
		},
	},
	Note:   []*model.Note{nil},
	Tag:    []*model.Tag{nil},
	TagMap: []*model.TagMap{nil},
	UserMark: []*model.UserMark{
		nil,
		nil,
		{
			UserMarkID:   2,
			ColorIndex:   1,
			LocationID:   2,
			StyleIndex:   0,
			UserMarkGUID: "This marking has a different location depending on nwt/nwtsty",
			Version:      1,
		},
		{
			UserMarkID:   3,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "SomeGUID",
			Version:      1,
		},
	},
}

var rightDBNwtWithDifferentDocID = &model.Database{
	BlockRange: []*model.BlockRange{
		nil,
		{
			BlockRangeID: 1,
			BlockType:    2,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{7, true},
			UserMarkID:   1,
		},
		{
			BlockRangeID: 2,
			BlockType:    2,
			Identifier:   2,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{5, true},
			UserMarkID:   1,
		},
		{
			BlockRangeID: 3,
			BlockType:    2,
			Identifier:   3,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{5, true},
			UserMarkID:   2,
		},
	},
	Bookmark:   []*model.Bookmark{nil},
	InputField: []*model.InputField{nil},
	Location: []*model.Location{
		nil,
		{
			LocationID:   1,
			DocumentID:   sql.NullInt64{987654332, true},
			KeySymbol:    sql.NullString{"nwt", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
		},
		// Make sure that when migrating to nwtsty we don't have duplicate locations
		// in case the nwtsty-location already exists on the other side
		{
			LocationID:    2,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{1, true},
			KeySymbol:     sql.NullString{"nwt", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
		},
	},
	Note:   []*model.Note{nil},
	Tag:    []*model.Tag{nil},
	TagMap: []*model.TagMap{nil},
	UserMark: []*model.UserMark{
		nil,
		{
			UserMarkID:   1,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "This marking has a different location depending on nwt/nwtsty",
			Version:      1,
		},
		{
			UserMarkID:   2,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "This should stay in nwt because of different DocumentIDs",
			Version:      1,
		},
	},
}

var mergedDBNwtWithDifferentDocID = &model.Database{
	BlockRange: []*model.BlockRange{
		nil,
		{
			BlockRangeID: 1,
			BlockType:    2,
			Identifier:   1,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{7, true},
			UserMarkID:   1,
		},
		{
			BlockRangeID: 2,
			BlockType:    2,
			Identifier:   5,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{20, true},
			UserMarkID:   2,
		},
		{
			BlockRangeID: 3,
			BlockType:    2,
			Identifier:   2,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{5, true},
			UserMarkID:   1,
		},
		{
			BlockRangeID: 4,
			BlockType:    2,
			Identifier:   3,
			StartToken:   sql.NullInt32{0, true},
			EndToken:     sql.NullInt32{5, true},
			UserMarkID:   3,
		},
	},
	Bookmark:   []*model.Bookmark{nil},
	InputField: []*model.InputField{nil},
	Location: []*model.Location{
		nil,
		{
			LocationID:   1,
			DocumentID:   sql.NullInt64{1102021811, true},
			KeySymbol:    sql.NullString{"lffi", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
		},
		{
			LocationID:   2,
			DocumentID:   sql.NullInt64{123456789, true},
			KeySymbol:    sql.NullString{"nwtsty", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
		},
		{
			LocationID:   3,
			DocumentID:   sql.NullInt64{987654332, true},
			KeySymbol:    sql.NullString{"nwt", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
		},
		{
			LocationID:    4,
			BookNumber:    sql.NullInt32{1, true},
			ChapterNumber: sql.NullInt32{1, true},
			KeySymbol:     sql.NullString{"nwtsty", true},
			MepsLanguage:  sql.NullInt32{Int32: 2, Valid: true},
		},
	},
	Note:   []*model.Note{nil},
	Tag:    []*model.Tag{nil},
	TagMap: []*model.TagMap{nil},
	UserMark: []*model.UserMark{
		nil,
		{
			UserMarkID:   1,
			ColorIndex:   1,
			LocationID:   2,
			StyleIndex:   0,
			UserMarkGUID: "This marking has a different location depending on nwt/nwtsty",
			Version:      1,
		},
		{
			UserMarkID:   2,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "SomeGUID",
			Version:      1,
		},
		{
			UserMarkID:   3,
			ColorIndex:   1,
			LocationID:   3,
			StyleIndex:   0,
			UserMarkGUID: "This should stay in nwt because of different DocumentIDs",
			Version:      1,
		},
	},
}
