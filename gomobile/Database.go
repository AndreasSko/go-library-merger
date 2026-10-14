package gomobile

import (
	"errors"
	"fmt"

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
	var db **model.Database
	switch side {
	case "leftSide":
		db = &dbw.left
	case "rightSide":
		db = &dbw.right
	default:
		return errors.New("only leftSide and rightSide are valid for importing backups")
	}
	imported := *db
	if imported == nil {
		imported = &model.Database{}
	}
	imported.TempDir, imported.SkipPlaylists = dbw.TempDir, dbw.skipPlaylists
	if err := imported.ImportJWLBackup(filename); err != nil {
		return err
	}
	*db = imported
	return nil
}

// SkipPlaylists allows to skip the check if playlists exist in the database.
// It is meant as a temporary workaround until merging of playlists is implemented.
func (dbw *DatabaseWrapper) SkipPlaylists(skipPlaylists bool) {
	dbw.skipPlaylists = skipPlaylists
}

// Init initializes the DatabaseWrapper to prepare for subsequent
// function calls. Should be called after ImportJWLBackup.
func (dbw *DatabaseWrapper) Init() error {
	if dbw.left == nil || dbw.right == nil {
		return errors.New("both backup sides must be loaded")
	}
	if err := dbw.merged.Close(); err != nil {
		return err
	}
	dbw.leftTmp = model.MakeDatabaseCopy(dbw.left)
	dbw.rightTmp = model.MakeDatabaseCopy(dbw.right)
	dbw.merged = &model.Database{
		TempDir: dbw.TempDir,
	}
	merger.PrepareDatabasesPreMerge(dbw.leftTmp, dbw.rightTmp)
	return nil
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
	if dbw.merged == nil {
		return errors.New("merge has not been initialized")
	}
	if err := merger.PrepareDatabasesPostMerge(dbw.merged); err != nil {
		return err
	}
	return dbw.merged.ExportJWLBackup(filename)
}

// PlaylistThumbnail returns the thumbnail of a playlist item in the imported
// backup on the given side, or the default thumbnail if the item has none.
func (dbw *DatabaseWrapper) PlaylistThumbnail(side string, playlistItemID int) ([]byte, error) {
	var db *model.Database
	switch side {
	case "leftSide":
		db = dbw.left
	case "rightSide":
		db = dbw.right
	default:
		return nil, fmt.Errorf("invalid database side %q", side)
	}
	if db == nil {
		return nil, fmt.Errorf("database side %q is not loaded", side)
	}
	return db.PlaylistThumbnail(playlistItemID)
}

// Close releases temporary files after export or cancellation.
func (dbw *DatabaseWrapper) Close() error {
	leftErr := dbw.left.Close()
	if leftErr == nil {
		dbw.left = nil
	}
	rightErr := dbw.right.Close()
	if rightErr == nil {
		dbw.right = nil
	}
	mergedErr := dbw.merged.Close()
	if mergedErr == nil {
		dbw.merged = nil
	}
	dbw.leftTmp, dbw.rightTmp = nil, nil
	return errors.Join(leftErr, rightErr, mergedErr)
}
