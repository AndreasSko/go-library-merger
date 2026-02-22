package gomobile

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
	"github.com/stretchr/testify/require"
)

func TestPlaylistConflictMobileContract(t *testing.T) {
	for _, kind := range []string{"item", "location", "media"} {
		t.Run(kind, func(t *testing.T) {
			label := "A"
			if kind == "media" {
				label = "video"
			}
			path := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Common", Items: []string{label}}})
			db := &DatabaseWrapper{TempDir: t.TempDir()}
			t.Cleanup(func() { require.NoError(t, db.Close()) })
			require.NoError(t, db.ImportJWLBackup(path, "leftSide"))
			require.NoError(t, db.ImportJWLBackup(path, "rightSide"))
			require.Equal(t, 1, db.Stats("leftSide").Playlist)
			require.Equal(t, 1, db.Stats("leftSide").PlaylistItem)
			switch kind {
			case "item":
				db.right.PlaylistItem[1].EndAction = 3
			case "location":
				db.right.PlaylistItemLocationMap[1].BaseDurationTicks.Int64 = 6000000000
			case "media":
				db.right.PlaylistItemIndependentMediaMap[1].DurationTicks = 6000000000
			}
			conflicts := &MergeConflictsWrapper{}
			conflicts.InitDBWrapper(db)
			run := func() error {
				db.Init()
				require.NoError(t, db.MergeLocations())
				require.NoError(t, db.MergeTags())
				return db.MergePlaylists("", conflicts)
			}
			require.Error(t, run())
			conflict, err := conflicts.NextConflict()
			require.NoError(t, err)
			for _, data := range []string{conflict.Left, conflict.Right} {
				var payload struct {
					Model   map[string]json.RawMessage `json:"model"`
					Related struct {
						Tag struct {
							Name string `json:"name"`
						} `json:"tag"`
					} `json:"related"`
				}
				require.NoError(t, json.Unmarshal([]byte(data), &payload))
				require.Contains(t, payload.Model, "type")
				require.Contains(t, payload.Model, "playlistItemId")
				require.Equal(t, "Common", payload.Related.Tag.Name)
			}
			require.NoError(t, conflicts.SolveConflict(conflict.Key, "rightSide"))
			require.NoError(t, run())
			require.NoError(t, db.MergeTagMaps())
			require.NoError(t, db.ExportMerged(filepath.Join(t.TempDir(), "merged.jwlibrary")))
			require.NoError(t, db.ExportMerged(filepath.Join(t.TempDir(), "again.jwlibrary")))
		})
	}
}
