//go:build !windows
// +build !windows

package gomobile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testdataDir = filepath.Join("..", "model", "testdata")
var backupFile = filepath.Join(testdataDir, "backup.jwlibrary")

func TestDatabaseWrapper_ImportJWLBackup(t *testing.T) {
	tests := []struct {
		name       string
		dbw        *DatabaseWrapper
		filename   string
		side       string
		wantErr    assert.ErrorAssertionFunc
		assertions func(t *testing.T, dbWrapper *DatabaseWrapper)
	}{
		{
			name:     "import right",
			dbw:      &DatabaseWrapper{},
			filename: backupFile,
			side:     "rightSide",
			wantErr:  assert.NoError,
			assertions: func(t *testing.T, dbWrapper *DatabaseWrapper) {
				assert.Len(t, dbWrapper.right.BlockRange, 5)
				assert.Len(t, dbWrapper.right.Bookmark, 3)
				assert.Len(t, dbWrapper.right.Location, 9)
				assert.Len(t, dbWrapper.right.Note, 3)
				assert.Len(t, dbWrapper.right.Tag, 3)
				assert.Len(t, dbWrapper.right.TagMap, 3)
				assert.Len(t, dbWrapper.right.UserMark, 5)
				assert.Equal(t, dbWrapper.TempDir, dbWrapper.right.TempDir)
			},
		},
		{
			name: "contains playlists, skip playlists set, import left",
			dbw: &DatabaseWrapper{
				TempDir:       t.TempDir(),
				skipPlaylists: true,
			},
			filename: filepath.Join(testdataDir, "backup_withPlaylist.jwlibrary"),
			side:     "leftSide",
			wantErr:  assert.NoError,
			assertions: func(t *testing.T, dbWrapper *DatabaseWrapper) {
				assert.Len(t, dbWrapper.left.BlockRange, 5)
				assert.Len(t, dbWrapper.left.Bookmark, 4)
				assert.Len(t, dbWrapper.left.Location, 9)
				assert.Len(t, dbWrapper.left.Note, 3)
				assert.Len(t, dbWrapper.left.Tag, 5)
				assert.Len(t, dbWrapper.left.TagMap, 5)
				assert.Len(t, dbWrapper.left.UserMark, 5)
				assert.Equal(t, dbWrapper.TempDir, dbWrapper.left.TempDir)
			},
		},
		{
			name:     "contains playlists, skip playlists not set",
			dbw:      &DatabaseWrapper{},
			filename: filepath.Join(testdataDir, "backup_withPlaylist.jwlibrary"),
			side:     "leftSide",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(tt, err, "merging of playlists is not supported yet")
			},
		},
		{
			name:     "wrong file",
			dbw:      &DatabaseWrapper{},
			filename: "wrongFile",
			side:     "leftSide",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(tt, err, "no such file or directory")
			},
		},
		{
			name:     "wrong side",
			dbw:      &DatabaseWrapper{},
			filename: backupFile,
			side:     "wrongSide",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(tt, err, "only leftSide and rightSide are valid for importing backups")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.dbw.TempDir = t.TempDir()
			err := tt.dbw.ImportJWLBackup(tt.filename, tt.side)
			t.Cleanup(func() { assert.NoError(t, tt.dbw.Close()) })
			tt.wantErr(t, err)

			if tt.assertions != nil {
				tt.assertions(t, tt.dbw)
			}
		})
	}
}

func TestDatabaseWrapper_SkipPlaylists(t *testing.T) {
	tests := []struct {
		name         string
		dbWrapper    *DatabaseWrapper
		skipPlaylist bool
	}{
		{
			name:         "skipPlaylists is false",
			dbWrapper:    &DatabaseWrapper{},
			skipPlaylist: false,
		},
		{
			name:         "skipPlaylists is true",
			dbWrapper:    &DatabaseWrapper{},
			skipPlaylist: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.dbWrapper.SkipPlaylists(tt.skipPlaylist)
			assert.Equal(t, tt.skipPlaylist, tt.dbWrapper.skipPlaylists)
		})
	}
}

func TestDatabaseWrapper_Init(t *testing.T) {
	dbw := &DatabaseWrapper{
		TempDir: t.TempDir(),
	}

	assert.NoError(t, dbw.ImportJWLBackup(backupFile, "leftSide"))
	t.Cleanup(func() { assert.NoError(t, dbw.Close()) })
	assert.NoError(t, dbw.ImportJWLBackup(backupFile, "rightSide"))
	t.Cleanup(func() { assert.NoError(t, dbw.Close()) })

	assert.NoError(t, dbw.Init())

	assert.True(t, dbw.merged.Equals(&model.Database{
		TempDir: dbw.TempDir,
	}))
	assert.True(t, dbw.leftTmp.Equals(dbw.left))
	assert.True(t, dbw.rightTmp.Equals(dbw.right))

	dbw.leftTmp.Bookmark[1].Title = "Tweaked"
	assert.False(t, dbw.leftTmp.Equals(dbw.left))

	dbw.merged = model.MakeDatabaseCopy(dbw.left)
	assert.NoError(t, dbw.Init())
	assert.True(t, dbw.merged.Equals(&model.Database{
		TempDir: dbw.TempDir,
	}))
	assert.True(t, dbw.leftTmp.Equals(dbw.left))
	assert.True(t, dbw.rightTmp.Equals(dbw.right))
	assert.Equal(t, dbw.TempDir, dbw.merged.TempDir)
}

func TestDatabaseWrapper_DBIsLoaded(t *testing.T) {
	dbw := &DatabaseWrapper{}

	assert.False(t, dbw.DBIsLoaded("leftSide"))
	assert.False(t, dbw.DBIsLoaded("rightSide"))
	assert.False(t, dbw.DBIsLoaded("mergeSide"))

	assert.NoError(t, dbw.ImportJWLBackup(backupFile, "leftSide"))
	t.Cleanup(func() { assert.NoError(t, dbw.Close()) })
	assert.True(t, dbw.DBIsLoaded("leftSide"))
	assert.False(t, dbw.DBIsLoaded("rightSide"))
	assert.False(t, dbw.DBIsLoaded("mergeSide"))

	assert.NoError(t, dbw.ImportJWLBackup(backupFile, "rightSide"))
	t.Cleanup(func() { assert.NoError(t, dbw.Close()) })
	assert.True(t, dbw.DBIsLoaded("leftSide"))
	assert.True(t, dbw.DBIsLoaded("rightSide"))
	assert.False(t, dbw.DBIsLoaded("mergeSide"))

	dbw.merged = model.MakeDatabaseCopy(dbw.left)
	assert.True(t, dbw.DBIsLoaded("mergeSide"))

	assert.False(t, dbw.DBIsLoaded("wrongSide"))
}

func TestDatabaseWrapper_ExportMerged(t *testing.T) {
	tmp := t.TempDir()

	dbw := &DatabaseWrapper{}
	dbw.merged = &model.Database{}

	assert.NoError(t, dbw.merged.ImportJWLBackup(backupFile))
	t.Cleanup(func() { assert.NoError(t, dbw.merged.Close()) })

	newBackup := filepath.Join(tmp, "test.jwlibrary")
	assert.NoError(t, dbw.ExportMerged(newBackup))

	newDB := &model.Database{}
	assert.NoError(t, newDB.ImportJWLBackup(newBackup))
	t.Cleanup(func() { assert.NoError(t, newDB.Close()) })
	assert.True(t, dbw.merged.Equals(newDB))
}

func TestDatabaseWrapper_DBContainsPlaylists(t *testing.T) {
	tests := []struct {
		name      string
		dbWrapper *DatabaseWrapper
		side      string
		want      bool
	}{
		{
			name: "no playlists",
			dbWrapper: &DatabaseWrapper{
				left: &model.Database{
					ContainsPlaylists: false,
				},
				right: &model.Database{
					ContainsPlaylists: false,
				},
			},
			side: "leftSide",
			want: false,
		},
		{
			name: "Playlists on right side, check left",
			dbWrapper: &DatabaseWrapper{
				left: &model.Database{
					ContainsPlaylists: false,
				},
				right: &model.Database{
					ContainsPlaylists: true,
				},
			},
			side: "leftSide",
			want: false,
		},
		{
			name: "Playlists on right side, check right",
			dbWrapper: &DatabaseWrapper{
				left: &model.Database{
					ContainsPlaylists: false,
				},
				right: &model.Database{
					ContainsPlaylists: true,
				},
			},
			side: "rightSide",
			want: true,
		},
		{
			name: "Wrong side",
			dbWrapper: &DatabaseWrapper{
				left: &model.Database{
					ContainsPlaylists: false,
				},
				right: &model.Database{
					ContainsPlaylists: true,
				},
			},
			side: "wrong",
			want: false,
		},
		{
			name:      "DB not loaded",
			dbWrapper: &DatabaseWrapper{},
			side:      "rightSide",
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dbWrapper.DBContainsPlaylists(tt.side)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDatabaseWrapperWorkingDirectoryLifecycle(t *testing.T) {
	for _, tc := range []struct {
		name  string
		check func(*testing.T, *DatabaseWrapper)
	}{
		{"invalid side allocates nothing", func(t *testing.T, db *DatabaseWrapper) {
			require.Error(t, db.ImportJWLBackup(filepath.Join("..", "model", "testdata", "backup.jwlibrary"), "invalid"))
			entries, err := os.ReadDir(db.TempDir)
			require.NoError(t, err)
			require.Empty(t, entries)
			require.Nil(t, db.left)
			require.Nil(t, db.right)
		}},
		{"reinitializing closes output but keeps inputs", func(t *testing.T, db *DatabaseWrapper) {
			path := filepath.Join("..", "model", "testdata", "backup.jwlibrary")
			require.NoError(t, db.ImportJWLBackup(path, "leftSide"))
			require.NoError(t, db.ImportJWLBackup(path, "rightSide"))
			left, right := db.left.WorkingDir, db.right.WorkingDir
			require.NoError(t, db.Init())
			require.NoError(t, db.merged.CreateWorkingDirectory())
			output := db.merged.WorkingDir
			require.NoError(t, db.Init())
			require.NoDirExists(t, output)
			require.DirExists(t, left)
			require.DirExists(t, right)
			require.Equal(t, left, db.leftTmp.WorkingDir)
			require.Equal(t, right, db.rightTmp.WorkingDir)
		}},
		{"close releases session and is idempotent", func(t *testing.T, db *DatabaseWrapper) {
			path := filepath.Join("..", "model", "testdata", "backup.jwlibrary")
			require.NoError(t, db.ImportJWLBackup(path, "leftSide"))
			require.NoError(t, db.ImportJWLBackup(path, "rightSide"))
			require.NoError(t, db.Init())
			require.NoError(t, db.merged.CreateWorkingDirectory())
			dirs := []string{db.left.WorkingDir, db.right.WorkingDir, db.merged.WorkingDir}
			require.NoError(t, db.Close())
			for _, dir := range dirs {
				require.NoDirExists(t, dir)
			}
			for _, side := range []string{"leftSide", "rightSide", "mergeSide"} {
				require.False(t, db.DBIsLoaded(side))
			}
			for _, owner := range []*model.Database{db.left, db.right, db.merged, db.leftTmp, db.rightTmp} {
				require.Nil(t, owner)
			}
			require.NoError(t, db.Close())
			require.Error(t, db.Init())
			require.DirExists(t, db.TempDir, "the caller owns the parent directory")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &DatabaseWrapper{TempDir: t.TempDir()}
			t.Cleanup(func() { require.NoError(t, db.Close()) })
			tc.check(t, db)
		})
	}
}
