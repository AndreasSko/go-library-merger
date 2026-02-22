package gomobile

import (
	"errors"

	"github.com/AndreasSko/go-library-merger/merger"
	"github.com/AndreasSko/go-library-merger/model"
)

// DatabaseWrapper wraps the left, right, and merged
// Database structs so they can be used with Gomobile.
type DatabaseWrapper struct {
	TempDir string

	left   *model.Database
	right  *model.Database
	merged *model.Database

	// skipPlaylists allows to skip prevention of merging if playlists exist in the database.
	// It is meant as a temporary workaround until merging of playlists is implemented.
	skipPlaylists bool

	// Temporary databases allow to run a merge function multiple
	// times without changing the content of the original databases.
	leftTmp  *model.Database
	rightTmp *model.Database
}

// ImportJWLBackup imports a .jwlibrary backup file into the struct
// on the given side.
func (dbw *DatabaseWrapper) ImportJWLBackup(filename string, side string) error {
	db := &model.Database{
		TempDir:       dbw.TempDir,
		SkipPlaylists: dbw.skipPlaylists,
	}

	if err := db.ImportJWLBackup(filename); err != nil {
		return err
	}

	switch side {
	case "leftSide":
		dbw.left = db
	case "rightSide":
		dbw.right = db
	default:
		return errors.New("only leftSide and rightSide are valid for importing backups")
	}

	return nil
}

// SkipPlaylists allows to skip the check if playlists exist in the database.
// It is meant as a temporary workaround until merging of playlists is implemented.
func (dbw *DatabaseWrapper) SkipPlaylists(skipPlaylists bool) {
	dbw.skipPlaylists = skipPlaylists
}

// Init initializes the DatabaseWrapper to prepare for subsequent
// function calls. Should be called after ImportJWLBackup.
func (dbw *DatabaseWrapper) Init() {
	dbw.leftTmp = model.MakeDatabaseCopy(dbw.left)
	dbw.rightTmp = model.MakeDatabaseCopy(dbw.right)
	dbw.merged = &model.Database{
		TempDir: dbw.TempDir,
	}
	merger.PrepareDatabasesPreMerge(dbw.leftTmp, dbw.rightTmp)
}

// DBIsLoaded indicates if a DB on the given side has been loaded.
func (dbw *DatabaseWrapper) DBIsLoaded(side string) bool {
	switch side {
	case "leftSide":
		return dbw.left != nil
	case "rightSide":
		return dbw.right != nil
	case "mergeSide":
		return dbw.merged != nil
	}

	return false
}

// DBContainsPlaylists indicates if a DB on the given side contains playlists.
func (dbw *DatabaseWrapper) DBContainsPlaylists(side string) bool {
	if !dbw.DBIsLoaded(side) {
		return false
	}

	switch side {
	case "leftSide":
		return dbw.left.ContainsPlaylists
	case "rightSide":
		return dbw.right.ContainsPlaylists
	}

	return false
}

// ExportMerged exports the merged database to filename.
func (dbw *DatabaseWrapper) ExportMerged(filename string) error {
	merger.PrepareDatabasesPostMerge(dbw.merged)
	return dbw.merged.ExportJWLBackup(filename)
}
