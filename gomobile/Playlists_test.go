package gomobile

import (
	"path/filepath"
	"testing"

	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
	"github.com/stretchr/testify/require"
)

func TestPlaylistArchives(t *testing.T) {
	for _, tc := range playlisttest.Cases() {
		for _, reverse := range []bool{false, true} {
			name := tc.Name
			if reverse {
				name += "-reverse"
			}
			t.Run(name, func(t *testing.T) {
				left, right := tc.Left, tc.Right
				if reverse {
					left, right = right, left
				}
				db := &DatabaseWrapper{TempDir: t.TempDir()}
				t.Cleanup(func() { require.NoError(t, db.Close()) })
				require.NoError(t, db.ImportJWLBackup(playlisttest.Build(t, left), "leftSide"))
				require.NoError(t, db.ImportJWLBackup(playlisttest.Build(t, right), "rightSide"))
				require.True(t, db.DBContainsPlaylists("leftSide"))
				db.Init()
				out := filepath.Join(t.TempDir(), "merged.jwlibrary")
				require.ErrorContains(t, db.ExportMerged(out), "playlists have not been merged")
				conflicts := &MergeConflictsWrapper{}
				require.NoError(t, db.MergeLocations())
				require.NoError(t, db.MergeBookmarks("", conflicts))
				require.NoError(t, db.MergeInputFields("", conflicts))
				require.NoError(t, db.MergeTags())
				require.NoError(t, db.MergeUserMarkAndBlockRange("", conflicts))
				require.NoError(t, db.MergeNotes("", conflicts))
				// Exercise the existing caller flow without requiring a new method call.
				require.NoError(t, db.MergeTagMaps())
				require.NoError(t, db.ExportMerged(out))
				playlisttest.Verify(t, out, tc.Want)
			})
		}
	}
}
