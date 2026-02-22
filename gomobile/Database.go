package gomobile

import (
	"errors"
	"fmt"

	"github.com/AndreasSko/go-library-merger/merger"
	"github.com/AndreasSko/go-library-merger/model"
)

// PlaylistThumbnail returns the original input's thumbnail for a conflict item.
// The input item ID remains usable even when conflict paths were remapped to
// avoid a collision with the other backup. A missing thumbnail uses the default.
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

// DatabaseWrapper wraps the left, right, and merged
// Database structs so they can be used with Gomobile.
type DatabaseWrapper struct {
	TempDir string

	left   *model.Database
	right  *model.Database
	merged *model.Database

	// skipPlaylists allows to skip prevention of merging if playlists exist in the database.
	// It is meant as a temporary workaround until merging of playlists is implemented.
	skipPlaylists   bool
	playlistsMerged bool
	tagMapsMerged   bool

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
		if dbw.left != nil {
			_ = dbw.left.CleanupMedia()
		}
		dbw.left = db
	case "rightSide":
		if dbw.right != nil {
			_ = dbw.right.CleanupMedia()
		}
		dbw.right = db
	default:
		_ = db.CleanupMedia()
		return errors.New("only leftSide and rightSide are valid for importing backups")
	}

	dbw.playlistsMerged = false
	dbw.tagMapsMerged = false
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
	if dbw.merged != nil {
		_ = dbw.merged.CleanupMedia()
	}
	dbw.playlistsMerged = false
	dbw.tagMapsMerged = false
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
	if dbw.merged == nil {
		return errors.New("merge has not been initialized")
	}
	if (!dbw.playlistsMerged || !dbw.tagMapsMerged) && ((dbw.left != nil && dbw.left.ContainsPlaylists && !dbw.skipPlaylists) || (dbw.right != nil && dbw.right.ContainsPlaylists && !dbw.skipPlaylists)) {
		return errors.New("playlists have not been merged; call MergePlaylists and MergeTagMaps before exporting")
	}
	if err := merger.PrepareDatabasesPostMerge(dbw.merged); err != nil {
		return err
	}
	return dbw.merged.ExportJWLBackup(filename)
}

// Close releases temporary media owned by this wrapper. Call after export or cancellation.
func (dbw *DatabaseWrapper) Close() error {
	var first error
	for _, db := range []*model.Database{dbw.left, dbw.right, dbw.merged} {
		if db != nil {
			if err := db.CleanupMedia(); err != nil && first == nil {
				first = err
			}
		}
	}
	return first
}
