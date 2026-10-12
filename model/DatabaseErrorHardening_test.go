package model

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMalformedModelIDsReturnErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		id   int64
	}{
		{name: "negative", id: -1},
		{name: "zero", id: 0},
		{name: "integer overflow", id: 1<<63 - 1},
		{name: "allocation limit", id: 10_000_000},
		{name: "too sparse", id: 1_000_001},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := sql.Open("sqlite3", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Errorf("closing db: %v", err)
				}
			}()
			if _, err := db.Exec("CREATE TABLE Tag (TagId INTEGER, Type INTEGER, Name TEXT); INSERT INTO Tag VALUES (?, 0, 'fictional')", tc.id); err != nil {
				t.Fatal(err)
			}
			if _, err := fetchFromSQLite(db, &Tag{}); err == nil {
				t.Fatalf("accepted malformed ID %d", tc.id)
			}
		})
	}
}

func TestInvalidLocationLookupDoesNotPanic(t *testing.T) {
	db := &Database{Location: []*Location{nil}}
	if db.FetchFromTable("Location", -1) != nil {
		t.Fatal("negative lookup returned model")
	}
	for _, tc := range []struct {
		name string
		id   int
	}{
		{name: "negative", id: -1},
		{name: "zero", id: 0},
		{name: "absent", id: 1},
		{name: "out of range", id: 1000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db.UserMark = []*UserMark{nil, {UserMarkID: 1, LocationID: tc.id}}
			if err := db.ExportJWLBackup(filepath.Join(t.TempDir(), "bad.jwlibrary")); err == nil {
				t.Fatalf("accepted missing location %d", tc.id)
			}
		})
	}
}

func TestImportRejectsMissingUserMarkLocation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "malformed.db")
	if err := createEmptySQLiteDB(path); err != nil {
		t.Fatal(err)
	}
	sqlite, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sqlite.Exec("INSERT INTO UserMark (UserMarkId, ColorIndex, LocationId, StyleIndex, UserMarkGuid, Version) VALUES (1, 1, 999, 0, 'fictional', 1)"); err != nil {
		t.Fatal(err)
	}
	if err := sqlite.Close(); err != nil {
		t.Fatal(err)
	}
	if err := (&Database{}).importSQLite(path); err == nil || !strings.Contains(err.Error(), "references missing Location") {
		t.Fatalf("invalid reference: %v", err)
	}
}

func TestOrdinarySparseIDsRemainSupported(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("closing db: %v", err)
		}
	}()
	if _, err := db.Exec("CREATE TABLE Tag (TagId INTEGER, Type INTEGER, Name TEXT); INSERT INTO Tag VALUES (1, 0, 'first'), (400, 0, 'last')"); err != nil {
		t.Fatal(err)
	}
	rows, err := fetchFromSQLite(db, &Tag{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 401 || rows[1].ID() != 1 || rows[400].ID() != 400 || rows[399] != nil {
		t.Fatalf("sparse ID identity changed")
	}
}

func TestModelAllocationBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		count, maximum, minimum int64
		want                    int
		reject                  bool
	}{
		{"empty", 0, 0, 0, 1, false},
		{"small-sparse-floor", 1, 1_000_000, 1_000_000, 1_000_001, false},
		{"large-with-ordinary-gaps", 50_000, 3_200_000, 1, 3_200_001, false},
		{"large-too-sparse", 50_000, 3_200_001, 1, 0, true},
		{"absolute-bound", 1_000_000, 10_000_000, 1, 0, true},
		{"negative-among-positive", 2, 10, -1, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Errorf("closing db: %v", err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("database expectations: %v", err)
				}
			}()
			mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count", "max", "min"}).AddRow(tc.count, tc.maximum, tc.minimum))
			mock.ExpectClose()
			got, err := getSliceCapacity(db, &Tag{})
			if tc.reject {
				if err == nil {
					t.Fatal("unsafe allocation accepted")
				}
			} else if err != nil || got != tc.want {
				t.Fatalf("ordinary gaps rejected: %d %v", got, err)
			}
		})
	}
}
