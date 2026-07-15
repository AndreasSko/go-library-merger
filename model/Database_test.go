package model

import (
	"archive/zip"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestDatabase_Fetch(t *testing.T) {
	db := &Database{
		Location: []*Location{
			nil,
			{
				LocationID: 1,
				Title:      sql.NullString{"#1", true},
			},
			nil,
			{
				LocationID: 3,
				Title:      sql.NullString{"#3", true},
			},
		},
		Bookmark: []*Bookmark{
			nil,
			{
				BookmarkID: 1,
				Title:      "#1",
			},
			nil,
		},
	}

	assert.Equal(t, "#1", db.FetchFromTable("Location", 1).(*Location).Title.String)
	assert.Equal(t, nil, db.FetchFromTable("Location", 2))
	assert.Equal(t, nil, db.FetchFromTable("Location", 4))
	assert.Equal(t, nil, db.FetchFromTable("Location", 400))
	assert.Equal(t, "#1", db.FetchFromTable("Bookmark", 1).(*Bookmark).Title)
	assert.PanicsWithValue(t, "Table notexists does not exist in Database", func() {
		db.FetchFromTable("notexists", 2)
	})
}

func TestDatabase_PurgeTables(t *testing.T) {
	type args struct {
		tables []string
	}
	tests := []struct {
		name        string
		db          *Database
		args        args
		errContains string
		wantDB      *Database
	}{
		{
			name: "No DB",
			args: args{
				tables: []string{"Bookmark"},
			},
			errContains: "Database is nil",
		},
		{
			name: "One table doesn't exist",
			db: &Database{
				Bookmark: []*Bookmark{
					nil,
					{BookmarkID: 1},
					nil,
					{BookmarkID: 3},
				},
			},
			args: args{
				tables: []string{"Bookmark", "NonExistent"},
			},
			errContains: "table NonExistent does not exist in database",
		},
		{
			name: "Purge two tables, keep rest",
			db: &Database{
				BlockRange: []*BlockRange{
					nil,
					{BlockRangeID: 1},
					{BlockRangeID: 2},
				},
				Bookmark: []*Bookmark{
					nil,
					{BookmarkID: 1},
					nil,
					{BookmarkID: 3},
				},
				InputField: []*InputField{
					nil,
					nil,
				},
				Location: []*Location{
					nil,
					nil,
					{LocationID: 2},
				},
				Note: []*Note{
					nil,
					{NoteID: 1},
				},
				Tag: []*Tag{
					nil,
					{TagID: 1},
				},
				TagMap: []*TagMap{
					nil,
					{TagMapID: 1},
				},
				UserMark: []*UserMark{
					nil,
					{UserMarkID: 1},
					{UserMarkID: 2},
					{UserMarkID: 3},
				},
			},
			args: args{
				tables: []string{"UserMark", "Location", ""},
			},
			wantDB: &Database{
				BlockRange: []*BlockRange{
					nil,
					{BlockRangeID: 1},
					{BlockRangeID: 2},
				},
				Bookmark: []*Bookmark{
					nil,
					{BookmarkID: 1},
					nil,
					{BookmarkID: 3},
				},
				InputField: []*InputField{
					nil,
					nil,
				},
				Location: []*Location{nil},
				Note: []*Note{
					nil,
					{NoteID: 1},
				},
				Tag: []*Tag{
					nil,
					{TagID: 1},
				},
				TagMap: []*TagMap{
					nil,
					{TagMapID: 1},
				},
				UserMark: []*UserMark{nil},
			},
		},
		{
			name: "Purge table that's nil in Database",
			db: &Database{
				Bookmark: []*Bookmark{
					nil,
					{BookmarkID: 1},
					nil,
					{BookmarkID: 3},
				},
			},
			args: args{
				tables: []string{"BlockRange"},
			},
			wantDB: &Database{
				Bookmark: []*Bookmark{
					nil,
					{BookmarkID: 1},
					nil,
					{BookmarkID: 3},
				},
				BlockRange: []*BlockRange{nil},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.db.PurgeTables(tt.args.tables)
			if tt.errContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			assert.NoError(t, err)
			assert.True(t, tt.wantDB.Equals(tt.db))
		})
	}
}

func TestMakeDatabaseCopy(t *testing.T) {
	db := &Database{
		TempDir: "a-temp-dir",
	}

	path := filepath.Join("testdata", userDataFilename)
	assert.NoError(t, db.importSQLite(path))

	dbCp := MakeDatabaseCopy(db)
	assert.Equal(t, db.TempDir, dbCp.TempDir)
	assertEqualNotDeepSame(t, db.BlockRange, dbCp.BlockRange)
	assertEqualNotDeepSame(t, db.Bookmark, dbCp.Bookmark)
	assertEqualNotDeepSame(t, db.InputField, dbCp.InputField)
	assertEqualNotDeepSame(t, db.Location, dbCp.Location)
	assertEqualNotDeepSame(t, db.Note, dbCp.Note)
	assertEqualNotDeepSame(t, db.Tag, dbCp.Tag)
	assertEqualNotDeepSame(t, db.TagMap, dbCp.TagMap)
	assertEqualNotDeepSame(t, db.UserMark, dbCp.UserMark)
}

// assertEqualNotDeepSame asserts that the entries of two slices are equal
// but point to different memory addresses (so not the same).
func assertEqualNotDeepSame(t *testing.T, expected interface{}, actual interface{}) {
	expectedRefl := reflect.ValueOf(expected)
	actualRefl := reflect.ValueOf(actual)

	assert.Equal(t, expectedRefl.Len(), actualRefl.Len())

	for i := 0; i < expectedRefl.Len(); i++ {
		if expectedRefl.Index(i).IsNil() || actualRefl.Index(i).IsNil() {
			assert.Equal(t, expectedRefl.Index(i).IsNil(), actualRefl.Index(i).IsNil())
			continue
		}
		assert.Equal(t, expectedRefl.Index(i).Elem().Interface(), actualRefl.Index(i).Elem().Interface())
		assert.NotEqual(t, expectedRefl.Index(i).Pointer(), actualRefl.Index(i).Pointer())
	}
}

func Test_getTableEntryCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := mock.NewRows([]string{"Count"}).AddRow(123)
	mock.ExpectQuery("SELECT Count\\(\\*\\) FROM PlaylistItem").WillReturnRows(rows)

	res, err := getTableEntryCount(db, "PlaylistItem")
	assert.NoError(t, err)
	assert.Equal(t, 123, res)

	rows = mock.NewRows([]string{"Count"}).AddRow(0)
	mock.ExpectQuery("SELECT Count\\(\\*\\) FROM InputField").WillReturnRows(rows)

	res, err = getTableEntryCount(db, "InputField")
	assert.NoError(t, err)
	assert.Equal(t, 0, res)
}

func Test_getSliceCapacity(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := mock.NewRows([]string{"TagId"}).AddRow(3)
	mock.ExpectQuery("SELECT TagId FROM Tag ORDER BY TagId DESC LIMIT 1").WillReturnRows(rows)
	res, err := getSliceCapacity(db, &Tag{})
	assert.NoError(t, err)
	assert.Equal(t, 4, res)

	// Test with empty DB
	rows = mock.NewRows([]string{"TagId"})
	mock.ExpectQuery("SELECT TagId FROM Tag ORDER BY TagId DESC LIMIT 1").WillReturnRows(rows)
	res, err = getSliceCapacity(db, &Tag{})
	assert.NoError(t, err)
	assert.Equal(t, 1, res)

	// Test with Type that does not have an ID
	rows = mock.NewRows([]string{"Count"}).AddRow(5)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM InputField").WillReturnRows(rows)
	res, err = getSliceCapacity(db, &InputField{})
	assert.NoError(t, err)
	assert.Equal(t, 6, res)
}

func Test_fetchFromSQLite(t *testing.T) {
	path := filepath.Join("testdata", userDataFilename)
	sqlite, err := sql.Open("sqlite3", path+"?immutable=1")
	if err != nil {
		t.Fatal(errors.Wrap(err, "Error while opening SQLite database"))
	}
	defer sqlite.Close()

	blockRange, err := fetchFromSQLite(sqlite, &BlockRange{})
	assert.NoError(t, err)
	assert.Len(t, blockRange, 5)
	assert.Equal(t, &BlockRange{3, 2, 13, sql.NullInt32{Int32: 0, Valid: true}, sql.NullInt32{Int32: 14, Valid: true}, 3}, blockRange[3])

	bookmark, err := fetchFromSQLite(sqlite, &Bookmark{})
	assert.NoError(t, err)
	assert.Len(t, bookmark, 4)
	assert.Equal(t, &Bookmark{2, 3, 7, 4, "Philippians 4", sql.NullString{String: "12I know how to be low on provisions and how to have an abundance. In everything and in all circumstances I have learned the secret of both how to be full and how to hunger, both how to have an abundance and how to do without. ", Valid: true}, 0, sql.NullInt32{}}, bookmark[2])

	inputField, err := fetchFromSQLite(sqlite, &InputField{})
	assert.NoError(t, err)
	assert.Len(t, inputField, 4)
	assert.Equal(t, &InputField{8, "tt71", "First other..", 3}, inputField[3])

	location, err := fetchFromSQLite(sqlite, &Location{})
	assert.NoError(t, err)
	assert.Len(t, location, 9)
	assert.Equal(t, &Location{4, sql.NullInt32{Int32: 66, Valid: true}, sql.NullInt32{Int32: 21, Valid: true}, sql.NullInt64{}, sql.NullInt32{}, 0, sql.NullString{String: "nwtsty", Valid: true}, sql.NullInt32{Int32: 2, Valid: true}, 0, sql.NullString{String: "Offenbarung 21", Valid: true}, sql.NullString{}, sql.NullString{}}, location[4])

	note, err := fetchFromSQLite(sqlite, &Note{})
	assert.NoError(t, err)
	assert.Len(t, note, 3)
	assert.Equal(t, &Note{2, "F75A18EE-FC17-4E0B-ABB6-CC16DABE9610", sql.NullInt32{Int32: 3, Valid: true}, sql.NullInt32{Int32: 3, Valid: true}, sql.NullString{String: "For all things I have the strength through the one who gives me power.", Valid: true}, sql.NullString{String: "", Valid: true}, "2020-04-14T18:42:14+00:00", "2020-04-14T18:42:14+00:00", 2, sql.NullInt32{Int32: 13, Valid: true}}, note[2])

	tag, err := fetchFromSQLite(sqlite, &Tag{})
	assert.NoError(t, err)
	assert.Len(t, tag, 3)
	assert.Equal(t, &Tag{2, 1, "Strengthening"}, tag[2])

	tagMap, err := fetchFromSQLite(sqlite, &TagMap{})
	assert.NoError(t, err)
	assert.Len(t, tagMap, 3)
	assert.Equal(t, &TagMap{2, sql.NullInt32{Int32: 0, Valid: false}, sql.NullInt32{Int32: 0, Valid: false}, sql.NullInt32{Int32: 2, Valid: true}, 2, 1}, tagMap[2])

	userMark, err := fetchFromSQLite(sqlite, &UserMark{})
	assert.NoError(t, err)
	assert.Len(t, userMark, 5)
	assert.Equal(t, &UserMark{2, 1, 2, 0, "2C5E7B4A-4997-4EDA-9CFF-38A7599C487B", 1}, userMark[2])
}

func Test_isNullableMismatch(t *testing.T) {
	type args struct {
		err  error
		rows func() *sql.Rows
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Nullable mismatch with int",
			args: args{
				err: fmt.Errorf(`Scan error on column index 2, name "NotNullable": converting NULL to int is unsupported`),
				rows: func() *sql.Rows {
					return mockSQLRows(t,
						sqlmock.NewColumn("ID").OfType("INTEGER", 1),
						sqlmock.NewColumn("Nullable").OfType("INTEGER", sql.NullInt32{}).Nullable(true),
						sqlmock.NewColumn("NotNullable").OfType("INTEGER", nil))
				},
			},
			want: true,
		},
		{
			name: "Nullable mismatch with text",
			args: args{
				err: fmt.Errorf(`Scan error on column index 1, name "NotNullable": converting NULL to string is unsupported`),
				rows: func() *sql.Rows {
					return mockSQLRows(t,
						sqlmock.NewColumn("ID").OfType("INTEGER", 1),
						sqlmock.NewColumn("NotNullable").OfType("TEXT", nil),
						sqlmock.NewColumn("Nullable").OfType("INTEGER", sql.NullInt32{}).Nullable(true))
				},
			},
			want: true,
		},
		{
			name: "No nullable mismatch",
			args: args{
				err: fmt.Errorf(`Scan error on column index 1, name "Nullable": converting NULL to int is unsupported`),
				rows: func() *sql.Rows {
					return mockSQLRows(t,
						sqlmock.NewColumn("ID").OfType("INTEGER", 1),
						sqlmock.NewColumn("Nullable").OfType("INTEGER", sql.NullInt32{}).Nullable(true),
						sqlmock.NewColumn("NotNullable").OfType("INTEGER", 3))
				},
			},
			want: false,
		},
		{
			name: "Empty",
		},
		{
			name: "Different error",
			args: args{
				err: fmt.Errorf(`Scan error on column index 1, name "Nullable": converting mock to int is unsupported`),
				rows: func() *sql.Rows {
					return mockSQLRows(t,
						sqlmock.NewColumn("ID").OfType("INTEGER", 1),
						sqlmock.NewColumn("Nullable").OfType("INTEGER", sql.NullInt32{}).Nullable(true),
						sqlmock.NewColumn("NotNullable").OfType("INTEGER", 3))
				},
			},
		},
		{
			name: "No error",
			args: args{
				rows: func() *sql.Rows {
					return mockSQLRows(t,
						sqlmock.NewColumn("ID").OfType("INTEGER", 1),
						sqlmock.NewColumn("Nullable").OfType("INTEGER", sql.NullInt32{}).Nullable(true),
						sqlmock.NewColumn("NotNullable").OfType("INTEGER", 3))
				},
			},
		},
		{
			name: "No rows",
			args: args{
				err: fmt.Errorf(`Scan error on column index 1, name "Nullable": converting NULL to int is unsupported`),
			},
		},
		{
			name: "Column name not matching column index",
			args: args{
				err: fmt.Errorf(`Scan error on column index 2, name "WrongName": converting NULL to int is unsupported`),
				rows: func() *sql.Rows {
					return mockSQLRows(t,
						sqlmock.NewColumn("ID").OfType("INTEGER", 1),
						sqlmock.NewColumn("Nullable").OfType("INTEGER", sql.NullInt32{}).Nullable(true),
						sqlmock.NewColumn("NotNullable").OfType("INTEGER", 3))
				},
			},
			want: false,
		},
		{
			name: "Type mismatch",
			args: args{
				err: fmt.Errorf(`Scan error on column index 2, name "NotNullable": converting NULL to string is unsupported`),
				rows: func() *sql.Rows {
					return mockSQLRows(t,
						sqlmock.NewColumn("ID").OfType("INTEGER", 1),
						sqlmock.NewColumn("Nullable").OfType("INTEGER", sql.NullInt32{}).Nullable(true),
						sqlmock.NewColumn("NotNullable").OfType("INTEGER", 3))
				},
			},
			want: false,
		},
		{
			name: "Negative index",
			args: args{
				err: fmt.Errorf(`Scan error on column index -5, name "NotNullable": converting NULL to string is unsupported`),
				rows: func() *sql.Rows {
					return mockSQLRows(t,
						sqlmock.NewColumn("ID").OfType("INTEGER", 1),
						sqlmock.NewColumn("Nullable").OfType("INTEGER", sql.NullInt32{}).Nullable(true),
						sqlmock.NewColumn("NotNullable").OfType("TEXT", nil))
				},
			},
			want: false,
		},
		{
			name: "More columns in error than in rows - don't panic with index out of range",
			args: args{
				err: fmt.Errorf(`Scan error on column index 2, name "NotNullable": converting NULL to int is unsupported`),
				rows: func() *sql.Rows {
					return mockSQLRows(t,
						sqlmock.NewColumn("ID").OfType("int", 1),
						sqlmock.NewColumn("Nullable").OfType("int", sql.NullInt32{}).Nullable(true))
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.rows == nil {
				tt.args.rows = func() *sql.Rows { return nil }
			}
			got := isNullableMismatch(tt.args.err, tt.args.rows())
			assert.Equal(t, tt.want, got)
		})
	}
}

func mockSQLRows(t *testing.T, columns ...*sqlmock.Column) *sql.Rows {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("Placeholder").WillReturnRows(mock.NewRowsWithColumnDefinition(columns...))
	result, err := db.Query("Placeholder")
	assert.NoError(t, err)
	return result
}

func TestDatabase_importSQLite(t *testing.T) {
	tests := []struct {
		name       string
		db         *Database
		filename   string
		wantErr    assert.ErrorAssertionFunc
		assertions func(t *testing.T, db *Database)
	}{
		{
			name:     "Regular import",
			db:       &Database{},
			filename: filepath.Join("testdata", userDataFilename),
			wantErr:  assert.NoError,
			assertions: func(t *testing.T, db *Database) {
				assert.Len(t, db.BlockRange, 5)
				assert.Len(t, db.Bookmark, 4)
				assert.Len(t, db.InputField, 4)
				assert.Len(t, db.Location, 9)
				assert.Len(t, db.Note, 3)
				assert.Len(t, db.Tag, 3)
				assert.Len(t, db.TagMap, 3)
				assert.Len(t, db.UserMark, 5)
			},
		},
		{
			name:     "Playlists included, error!",
			db:       &Database{},
			filename: filepath.Join("testdata", "userData_withPlaylist.db"),
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "merging of playlists is not supported yet")
			},
			assertions: func(t *testing.T, db *Database) {},
		},
		{
			name: "Playlists included, skip them!",
			db: &Database{
				SkipPlaylists: true,
			},
			filename: filepath.Join("testdata", "userData_withPlaylist.db"),
			wantErr:  assert.NoError,
			assertions: func(t *testing.T, db *Database) {
				assert.Nil(t, db.Tag[3])
				assert.NotNil(t, db.Tag[4])
				assert.Nil(t, db.TagMap[3])
				assert.Nil(t, db.TagMap[4])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.db.importSQLite(tt.filename)
			tt.wantErr(t, err)
			tt.assertions(t, tt.db)
		})
	}
}

func TestDatabase_removePlaylists(t *testing.T) {
	tests := []struct {
		name   string
		db     *Database
		wantDB *Database
	}{
		{
			name: "No playlists",
			db: &Database{
				Tag: []*Tag{
					nil,
					{
						TagID:   1,
						TagType: 0,
					},
					{
						TagID:   2,
						TagType: 1,
					},
				},
				TagMap: []*TagMap{
					nil,
					{
						TagMapID:   1,
						TagID:      1,
						LocationID: sql.NullInt32{Int32: 1, Valid: true},
					},
					nil,
					{
						TagMapID:   3,
						TagID:      2,
						LocationID: sql.NullInt32{Int32: 2, Valid: true},
					},
				},
			},
			wantDB: &Database{
				Tag: []*Tag{
					nil,
					{
						TagID:   1,
						TagType: 0,
					},
					{
						TagID:   2,
						TagType: 1,
					},
				},
				TagMap: []*TagMap{
					nil,
					{
						TagMapID:   1,
						TagID:      1,
						LocationID: sql.NullInt32{Int32: 1, Valid: true},
					},
					nil,
					{
						TagMapID:   3,
						TagID:      2,
						LocationID: sql.NullInt32{Int32: 2, Valid: true},
					},
				},
			},
		},
		{
			name: "Playlist tag included",
			db: &Database{
				Tag: []*Tag{
					nil,
					{
						TagID:   1,
						TagType: 2,
					},
					{
						TagID:   2,
						TagType: 1,
					},
				},
				TagMap: []*TagMap{
					nil,
					{
						TagMapID:   1,
						TagID:      1,
						LocationID: sql.NullInt32{Int32: 1, Valid: true},
					},
					nil,
					{
						TagMapID:   3,
						TagID:      2,
						LocationID: sql.NullInt32{Int32: 2, Valid: true},
					},
				},
			},
			wantDB: &Database{
				ContainsPlaylists: true,
				Tag: []*Tag{
					nil,
					nil,
					{
						TagID:   2,
						TagType: 1,
					},
				},
				TagMap: []*TagMap{
					nil,
					{
						TagMapID:   1,
						TagID:      1,
						LocationID: sql.NullInt32{Int32: 1, Valid: true},
					},
					nil,
					{
						TagMapID:   3,
						TagID:      2,
						LocationID: sql.NullInt32{Int32: 2, Valid: true},
					},
				},
			},
		},
		{
			name: "Playlist entry in TagMap included",
			db: &Database{
				Tag: []*Tag{
					nil,
					{
						TagID:   1,
						TagType: 0,
					},
					{
						TagID:   2,
						TagType: 1,
					},
				},
				TagMap: []*TagMap{
					nil,
					{
						TagMapID:       1,
						TagID:          1,
						PlaylistItemID: sql.NullInt32{Int32: 1, Valid: true},
					},
					nil,
					{
						TagMapID:   3,
						TagID:      2,
						LocationID: sql.NullInt32{Int32: 2, Valid: true},
					},
				},
			},
			wantDB: &Database{
				ContainsPlaylists: true,
				Tag: []*Tag{
					nil,
					{
						TagID:   1,
						TagType: 0,
					},
					{
						TagID:   2,
						TagType: 1,
					},
				},
				TagMap: []*TagMap{
					nil,
					nil,
					nil,
					{
						TagMapID:   3,
						TagID:      2,
						LocationID: sql.NullInt32{Int32: 2, Valid: true},
					},
				},
			},
		},
		{
			name: "Playlist with entries included",
			db: &Database{
				Tag: []*Tag{
					nil,
					{
						TagID:   1,
						TagType: 1,
					},
					{
						TagID:   2,
						TagType: 2,
					},
					{
						TagID:   3,
						TagType: 2,
					},
				},
				TagMap: []*TagMap{
					nil,
					{
						TagMapID:       1,
						TagID:          2,
						PlaylistItemID: sql.NullInt32{Int32: 1, Valid: true},
					},
					{
						TagMapID:       2,
						TagID:          2,
						PlaylistItemID: sql.NullInt32{Int32: 2, Valid: true},
					},
					{
						TagMapID:       3,
						TagID:          3,
						PlaylistItemID: sql.NullInt32{Int32: 3, Valid: true},
					},
					{
						TagMapID:   4,
						TagID:      1,
						LocationID: sql.NullInt32{Int32: 1, Valid: true},
					},
				},
			},
			wantDB: &Database{
				ContainsPlaylists: true,
				Tag: []*Tag{
					nil,
					{
						TagID:   1,
						TagType: 1,
					},
					nil,
					nil,
				},
				TagMap: []*TagMap{
					nil,
					nil,
					nil,
					nil,
					{
						TagMapID:   4,
						TagID:      1,
						LocationID: sql.NullInt32{Int32: 1, Valid: true},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.db.removePlaylists()
			assert.Equal(t, tt.wantDB, tt.db)
		})
	}
}

func TestDatabase_ImportJWLBackup(t *testing.T) {
	db := Database{}

	path := filepath.Join("testdata", "backup.jwlibrary")
	assert.NoError(t, db.ImportJWLBackup(path))

	// As we already test the correctness in Test_fetchFromSQLite,
	// it should be sufficient to just double-check the size of the slices.
	assert.Len(t, db.BlockRange, 5)
	assert.Len(t, db.Bookmark, 3)
	assert.Len(t, db.InputField, 4)
	assert.Len(t, db.Location, 9)
	assert.Len(t, db.Note, 3)
	assert.Len(t, db.Tag, 3)
	assert.Len(t, db.TagMap, 3)
	assert.Len(t, db.UserMark, 5)
}

func TestDatabase_ExportJWLBackup(t *testing.T) {
	// Create tmp folder and place all files there
	testFolder := t.TempDir()

	// Test if import->export->import tweakes Data in wrong way
	db := Database{}

	path := filepath.Join("testdata", "backup.jwlibrary")
	assert.NoError(t, db.ImportJWLBackup(path))

	newPath := filepath.Join(testFolder, "backup.jwlibrary")
	assert.NoError(t, db.ExportJWLBackup(newPath))

	// Make sure all expected files are there
	filenames, err := filenamesInZip(newPath)
	assert.NoError(t, err)
	for _, expected := range []string{"userData.db", "manifest.json", "default_thumbnail.png"} {
		assert.Contains(t, filenames, expected)
	}

	db = Database{}
	assert.NoError(t, db.ImportJWLBackup(newPath))

	assert.Len(t, db.BlockRange, 5)
	assert.Equal(t, &BlockRange{3, 2, 13, sql.NullInt32{Int32: 0, Valid: true}, sql.NullInt32{Int32: 14, Valid: true}, 3}, db.BlockRange[3])

	assert.Len(t, db.Bookmark, 3)
	assert.Equal(t, &Bookmark{2, 3, 7, 4, "Philippians 4", sql.NullString{String: "12I know how to be low on provisions and how to have an abundance. In everything and in all circumstances I have learned the secret of both how to be full and how to hunger, both how to have an abundance and how to do without. ", Valid: true}, 0, sql.NullInt32{}}, db.Bookmark[2])

	assert.Len(t, db.InputField, 4)
	assert.Equal(t, &InputField{8, "tt71", "First other..", 3}, db.InputField[3])

	assert.Len(t, db.Location, 9)
	assert.Equal(t, &Location{4, sql.NullInt32{Int32: 66, Valid: true}, sql.NullInt32{Int32: 21, Valid: true}, sql.NullInt64{}, sql.NullInt32{}, 0, sql.NullString{String: "nwtsty", Valid: true}, sql.NullInt32{Int32: 2, Valid: true}, 0, sql.NullString{String: "Offenbarung 21", Valid: true}, sql.NullString{}, sql.NullString{}}, db.Location[4])

	assert.Len(t, db.Note, 3)
	assert.Equal(t, &Note{2, "F75A18EE-FC17-4E0B-ABB6-CC16DABE9610", sql.NullInt32{Int32: 3, Valid: true}, sql.NullInt32{Int32: 3, Valid: true}, sql.NullString{String: "For all things I have the strength through the one who gives me power.", Valid: true}, sql.NullString{String: "", Valid: true}, "2020-04-14T18:42:14+00:00", "2020-04-14T18:42:14+00:00", 2, sql.NullInt32{Int32: 13, Valid: true}}, db.Note[2])

	assert.Len(t, db.Tag, 3)
	assert.Equal(t, &Tag{2, 1, "Strengthening"}, db.Tag[2])

	assert.Len(t, db.TagMap, 3)
	assert.Equal(t, &TagMap{2, sql.NullInt32{Int32: 0, Valid: false}, sql.NullInt32{Int32: 0, Valid: false}, sql.NullInt32{Int32: 2, Valid: true}, 2, 1}, db.TagMap[2])

	assert.Len(t, db.UserMark, 5)
	assert.Equal(t, &UserMark{2, 1, 2, 0, "2C5E7B4A-4997-4EDA-9CFF-38A7599C487B", 1}, db.UserMark[2])
}

func Test_createEmptySQLiteDB(t *testing.T) {
	// Create tmp folder and place all files there
	tmp := t.TempDir()

	path := filepath.Join(tmp, userDataFilename)
	assert.NoError(t, createEmptySQLiteDB(path))

	// Test if file has correct hash
	f, err := os.Open(path)
	if err != nil {
		errors.Wrap(err, "Error while opening SQLite file to calculate hash")
	}
	defer f.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		log.Fatal(err)
	}
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	assert.Equal(t, "37ea4cc3734847a874f83206d35c10f1f49f9c899f8263ac3f4461c8f2537710", hash)
}

func TestDatabase_saveToNewSQLite(t *testing.T) {
	// Create tmp folder and place all files there
	tmp := t.TempDir()

	db := Database{
		BlockRange: []*BlockRange{{3, 2, 13, sql.NullInt32{Int32: 0, Valid: true}, sql.NullInt32{Int32: 14, Valid: true}, 3}},
		Bookmark:   []*Bookmark{{2, 3, 7, 4, "Philippians 4", sql.NullString{String: "12 I know how to be low on provisions and how to have an abundance. In everything and in all circumstances I have learned the secret of both how to be full and how to hunger, both how to have an abundance and how to do without. ", Valid: true}, 0, sql.NullInt32{}}},
		InputField: []*InputField{{8, "tt56", "First lesson completed on..", 1}, {8, "tt66", "1", 3}, {8, "tt71", "First other..", 3}},
		Location: []*Location{
			{4, sql.NullInt32{Int32: 66, Valid: true}, sql.NullInt32{Int32: 21, Valid: true}, sql.NullInt64{}, sql.NullInt32{}, 0, sql.NullString{String: "nwtsty", Valid: true}, sql.NullInt32{Int32: 2, Valid: true}, 0, sql.NullString{String: "Offenbarung 21", Valid: true}, sql.NullString{}, sql.NullString{}},
			{5, sql.NullInt32{}, sql.NullInt32{}, sql.NullInt64{Int64: 2020401, Valid: true}, sql.NullInt32{}, 0, sql.NullString{String: "w", Valid: true}, sql.NullInt32{Int32: 2, Valid: true}, 0, sql.NullString{String: "A title", Valid: true}, sql.NullString{String: "E", Valid: true}, sql.NullString{String: "r2025", Valid: true}},
		},
		Note:     []*Note{{2, "F75A18EE-FC17-4E0B-ABB6-CC16DABE9610", sql.NullInt32{Int32: 3, Valid: true}, sql.NullInt32{Int32: 3, Valid: true}, sql.NullString{String: "For all things I have the strength through the one who gives me power.", Valid: true}, sql.NullString{String: "!", Valid: true}, "2020-04-14T18:42:14+00:00", "2020-04-14T18:42:14+00:00", 2, sql.NullInt32{Int32: 13, Valid: true}}},
		Tag:      []*Tag{{2, 1, "Strengthening"}},
		TagMap:   []*TagMap{{2, sql.NullInt32{Int32: 0, Valid: false}, sql.NullInt32{Int32: 0, Valid: false}, sql.NullInt32{Int32: 2, Valid: true}, 2, 1}},
		UserMark: []*UserMark{{2, 1, 2, 0, "2C5E7B4A-4997-4EDA-9CFF-38A7599C487B", 1}},
	}
	path := filepath.Join(tmp, userDataFilename)
	assert.NoError(t, db.saveToNewSQLite(path))

	db2 := Database{}
	assert.NoError(t, db2.importSQLite(path))

	assert.Equal(t, db.BlockRange[0], db2.BlockRange[3])
	assert.Equal(t, db.Bookmark[0], db2.Bookmark[2])
	assert.Equal(t, db.InputField[2], db2.InputField[3])
	assert.Equal(t, db.Location[0], db2.Location[4])
	assert.Equal(t, db.Location[1], db2.Location[5])
	assert.Equal(t, db.Note[0], db2.Note[2])
	assert.Equal(t, db.TagMap[0], db2.TagMap[2])
	assert.Equal(t, db.UserMark[0], db2.UserMark[2])

	// Check if saving empty tables is possible
	db = Database{
		BlockRange: []*BlockRange{{3, 2, 13, sql.NullInt32{Int32: 0, Valid: true}, sql.NullInt32{Int32: 14, Valid: true}, 3}},
		Bookmark:   []*Bookmark{nil},
	}
	assert.NoError(t, db.saveToNewSQLite(path))
}

func TestDatabase_Equals(t *testing.T) {
	db1 := &Database{}
	db2 := &Database{}

	path := filepath.Join("testdata", "backup.jwlibrary")
	assert.NoError(t, db1.ImportJWLBackup(path))
	assert.False(t, db1.Equals(db2))
	assert.NoError(t, db2.ImportJWLBackup(path))
	assert.True(t, db1.Equals(db2))

	db1.Location = append(db1.Location, &Location{
		MepsLanguage: sql.NullInt32{Int32: 100, Valid: true},
	})

	assert.False(t, db1.Equals(db2))

	db3 := &Database{}
	path = filepath.Join("testdata", "backup_shuffled.jwlibrary")
	assert.NoError(t, db3.ImportJWLBackup(path))
	assert.True(t, db2.Equals(db3))
}

// filenamesInZip returns the names of all files contained in a zip file.
func filenamesInZip(path string) ([]string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	filenames := make([]string, len(r.File))
	for i, f := range r.File {
		filenames[i] = f.Name
	}

	return filenames, nil
}
