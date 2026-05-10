//go:build !windows
// +build !windows

package cmd

import (
	"bytes"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/AndreasSko/go-library-merger/model"
	expect "github.com/Netflix/go-expect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_merge(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	emptyFilename := filepath.Join(tmp, "empty.jwlibrary")
	leftFilename := filepath.Join(tmp, "left.jwlibrary")
	rightFilename := filepath.Join(tmp, "right.jwlibrary")
	mergedFilename := filepath.Join(tmp, "merged.jwlibrary")
	leftMultiCollisionFilename := filepath.Join(tmp, "leftMultiCollision.jwlibrary")
	rightMultiCollisionFilename := filepath.Join(tmp, "rightMultiCollision.jwlibrary")
	assert.NoError(t, emptyDB.ExportJWLBackup(emptyFilename))
	assert.NoError(t, leftDB.ExportJWLBackup(leftFilename))
	assert.NoError(t, rightDB.ExportJWLBackup(rightFilename))
	assert.NoError(t, leftMultiCollision.ExportJWLBackup(leftMultiCollisionFilename))
	assert.NoError(t, rightMultiCollision.ExportJWLBackup(rightMultiCollisionFilename))

	// Merge against empty DB and see if result is still the same
	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			_, err := c.ExpectString("🎉 Finished merging!")
			assert.NoError(t, err)
			_, err = c.ExpectEOF()
			assert.NoError(t, err)
		},
		func(t *testing.T, c *expect.Console) {
			merge(leftFilename, emptyFilename, mergedFilename,
				terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()})
			merged := &model.Database{}
			merged.ImportJWLBackup(mergedFilename)
			assert.True(t, leftDB.Equals(merged))
		})

	// Merge while selecting all right
	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			c.ExpectString("📑 Merging Bookmarks")
			c.SendLine(string(terminal.KeyArrowDown))

			c.ExpectString("✍️  Merging InputFields")
			c.SendLine(string(terminal.KeyArrowDown))

			c.ExpectString("🖍  Merging Markings")
			c.SendLine(string(terminal.KeyArrowDown))

			c.ExpectString("📝 Merging Notes")
			c.SendLine(string(terminal.KeyArrowDown))

			c.ExpectEOF()
		},
		func(t *testing.T, c *expect.Console) {
			merge(leftFilename, rightFilename, mergedFilename,
				terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()})
			merged := &model.Database{}
			merged.ImportJWLBackup(mergedFilename)
			assert.True(t, mergedAllRightDB.Equals(merged))
		})

	// Merge while selecting all left
	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			c.ExpectString("📑 Merging Bookmarks")
			c.SendLine("")

			c.ExpectString("✍️  Merging InputFields")
			c.SendLine("")

			c.ExpectString("🖍  Merging Markings")
			c.SendLine("")

			c.ExpectString("📝 Merging Notes")
			c.SendLine("")

			c.ExpectEOF()
		},
		func(t *testing.T, c *expect.Console) {
			merge(leftFilename, rightFilename, mergedFilename,
				terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()})
			merged := &model.Database{}
			merged.ImportJWLBackup(mergedFilename)
			assert.True(t, mergedAllLeftDB.Equals(merged))
		})

	// Merge with auto resolution: chooseRight for Bookmarks & Markings,
	// chooseNewest for Notes
	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			c.ExpectString("🎉 Finished merging!")
			c.ExpectEOF()
		},
		func(t *testing.T, c *expect.Console) {
			BookmarkResolver = "chooseRight"
			MarkingResolver = "chooseRight"
			NoteResolver = "chooseNewest"
			InputFieldResolver = "chooseRight"
			merge(leftFilename, rightFilename, mergedFilename,
				terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()})
			merged := &model.Database{}
			merged.ImportJWLBackup(mergedFilename)
			assert.True(t, mergedAllRightDB.Equals(merged))
		})

	// Merge multiCollision with autoresolver all right
	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			_, err := c.ExpectString("🎉 Finished merging!")
			assert.NoError(t, err)
			c.ExpectEOF()
		},
		func(t *testing.T, c *expect.Console) {
			MarkingResolver = "chooseRight"
			merge(leftMultiCollisionFilename,
				rightMultiCollisionFilename,
				mergedFilename,
				terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()})
			merged := &model.Database{}
			merged.ImportJWLBackup(mergedFilename)
			assert.True(t, rightMultiCollision.Equals(merged))
		})

	// Merge while selecting all left and also migrate left to nwtsty
	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			c.ExpectString("📑 Merging Bookmarks")
			c.SendLine("")

			c.ExpectString("✍️  Merging InputFields")
			c.SendLine("")

			c.ExpectString("🖍  Merging Markings")
			c.SendLine("")

			c.ExpectString("📝 Merging Notes")
			c.SendLine("")

			c.ExpectEOF()
		},
		func(t *testing.T, c *expect.Console) {
			leftNwtFilename := filepath.Join(tmp, "leftNwt.jwlibrary")
			rightNwtFilename := filepath.Join(tmp, "rightNwt.jwlibrary")
			assert.NoError(t, leftNwtDB.ExportJWLBackup(leftNwtFilename))
			assert.NoError(t, rightNwtDB.ExportJWLBackup(rightNwtFilename))

			merge(leftNwtFilename, rightNwtFilename, mergedFilename,
				terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()})
			merged := &model.Database{}
			merged.ImportJWLBackup(mergedFilename)
			assert.True(t, mergedAllLeftNwtDB.Equals(merged))
		})

	// Merge and test proper nwt migration while not migrating locations with different DocIds
	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			c.ExpectEOF()
		},
		func(t *testing.T, c *expect.Console) {
			leftNwtWithDifferentDocIDFilename := filepath.Join(tmp, "leftNwtWithDifferentDocIDFilename.jwlibrary")
			assert.NoError(t, leftDBNwtWithDifferentDocID.ExportJWLBackup(leftNwtWithDifferentDocIDFilename))
			rightNwtWithDifferentDocIDFilename := filepath.Join(tmp, "rightNwtWithDifferentDocIDFilename.jwlibrary")
			assert.NoError(t, rightDBNwtWithDifferentDocID.ExportJWLBackup(rightNwtWithDifferentDocIDFilename))
			mergedFilename := filepath.Join(tmp, "mergedNwtWithDifferentDocIDFilename.jwlibrary")
			assert.NoError(t, mergedDBNwtWithDifferentDocID.ExportJWLBackup(mergedFilename))

			merge(leftNwtWithDifferentDocIDFilename, rightNwtWithDifferentDocIDFilename, mergedFilename,
				terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()})
			merged := &model.Database{}
			merged.ImportJWLBackup(mergedFilename)
			assert.True(t, mergedDBNwtWithDifferentDocID.Equals(merged))
		})

	// Merge with marking containing no BlockRanges and selecting all left
	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			c.ExpectString("📑 Merging Bookmarks")
			c.SendLine("")

			c.ExpectString("✍️  Merging InputFields")
			c.SendLine("")

			c.ExpectString("🖍  Merging Markings")
			c.SendLine("")

			c.ExpectString("📝 Merging Notes")
			c.SendLine("")

			_, err := c.ExpectString("🎉 Finished merging!")
			assert.NoError(t, err)
			c.ExpectEOF()
		},
		func(t *testing.T, c *expect.Console) {
			leftDBEmptyBRFilename := filepath.Join(tmp, "leftDBEmptyBRFilename.jwlibrary")
			assert.NoError(t, leftDBEmptyBR.ExportJWLBackup(leftDBEmptyBRFilename))
			RightDBEmptyBRFilename := filepath.Join(tmp, "RightDBEmptyBRFilename.jwlibrary")
			assert.NoError(t, RightDBEmptyBR.ExportJWLBackup(RightDBEmptyBRFilename))
			mergedAllLeftDBEmptyBRFilename := filepath.Join(tmp, "mergedAllLeftDBEmptyBRFilename.jwlibrary")
			assert.NoError(t, mergedAllLeftDBEmptyBR.ExportJWLBackup(mergedAllLeftDBEmptyBRFilename))

			merge(leftDBEmptyBRFilename,
				RightDBEmptyBRFilename,
				mergedFilename,
				terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()})
			merged := &model.Database{}
			merged.ImportJWLBackup(mergedAllLeftDBEmptyBRFilename)
			assert.True(t, mergedAllLeftDBEmptyBR.Equals(merged))
		})

	// Attempt to merge with playlists, while skipPlaylists is set to false
	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			c.ExpectString("Error: failed to import left backup: merging of playlists is not supported yet. " +
				"Enable SkipPlaylists flag to skip this safety check")
			c.ExpectEOF()
		},
		func(t *testing.T, c *expect.Console) {
			merge("../model/testdata/backup_withPlaylist.jwlibrary",
				"../model/testdata/backup_withPlaylist.jwlibrary",
				mergedFilename,
				terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()})
		},
	)

	// Merge with playlists and skipPlaylists set to true
	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			c.ExpectString("🎉 Finished merging!")
			c.ExpectEOF()
		},
		func(t *testing.T, c *expect.Console) {
			SkipPlaylists = true
			defer func() { SkipPlaylists = false }()
			merge("../model/testdata/backup_withPlaylist.jwlibrary",
				"../model/testdata/backup_withPlaylist.jwlibrary",
				mergedFilename,
				terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()})
		},
	)
}

// https://github.com/AlecAivazis/survey/blob/master/survey_posix_test.go
func RunCmdTest(t *testing.T, procedure func(*testing.T, *expect.Console), test func(*testing.T, *expect.Console)) {
	// Multiplex output to a buffer as well for the raw bytes.
	buf := new(bytes.Buffer)
	c, err := expect.NewConsole(expect.WithStdout(buf))
	require.Nil(t, err)
	defer c.Close()

	donec := make(chan struct{})
	go func() {
		defer close(donec)
		procedure(t, c)
	}()

	test(t, c)

	// Close the slave end of the pty, and read the remaining bytes from the master end.
	c.Tty().Close()
	<-donec

	t.Logf("Raw output: %q", buf.String())
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
			DocumentID:   sql.NullInt32{1102021811, true},
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
			DocumentID:   sql.NullInt32{1102021811, true},
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
			DocumentID:   sql.NullInt32{1102021811, true},
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
			DocumentID:   sql.NullInt32{1102021811, true},
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
			DocumentID:   sql.NullInt32{1102021811, true},
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
			DocumentID:   sql.NullInt32{1102021811, true},
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
			DocumentID:   sql.NullInt32{1102021811, true},
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
			DocumentID:   sql.NullInt32{1102021811, true},
			KeySymbol:    sql.NullString{"lffi", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
		},
		{
			LocationID:   2,
			DocumentID:   sql.NullInt32{123456789, true},
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
			DocumentID:   sql.NullInt32{987654332, true},
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
			DocumentID:   sql.NullInt32{1102021811, true},
			KeySymbol:    sql.NullString{"lffi", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
		},
		{
			LocationID:   2,
			DocumentID:   sql.NullInt32{123456789, true},
			KeySymbol:    sql.NullString{"nwtsty", true},
			MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
		},
		{
			LocationID:   3,
			DocumentID:   sql.NullInt32{987654332, true},
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

var leftDBEmptyBR = &model.Database{
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
			DocumentID:   sql.NullInt32{1102021811, true},
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
		{
			UserMarkID:   2,
			ColorIndex:   1,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "No BlockRanges",
			Version:      1,
		},
	},
}

var RightDBEmptyBR = &model.Database{
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
			DocumentID:   sql.NullInt32{1102021811, true},
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
			Created:      "2020-09-20T13:52:25+00:00",
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
		{
			UserMarkID:   4,
			ColorIndex:   2,
			LocationID:   1,
			StyleIndex:   0,
			UserMarkGUID: "No BlockRanges",
			Version:      1,
		},
	},
}

var mergedAllLeftDBEmptyBR = &model.Database{
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
			DocumentID:   sql.NullInt32{1102021811, true},
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
			TagMapID: 3,
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
