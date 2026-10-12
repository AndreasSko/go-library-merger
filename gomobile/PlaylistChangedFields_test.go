package gomobile

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

func TestPlaylistMeaningfulConflictFieldsAndBothChoices(t *testing.T) {
	tests := []struct {
		field       string
		independent bool
		change      func(*model.Database)
	}{
		{"label", false, func(db *model.Database) { db.PlaylistItem[1].Label = "Changed" }},
		{"startTrimOffsetTicks", false, func(db *model.Database) { db.PlaylistItem[1].StartTrimOffsetTicks.Int64++ }},
		{"endTrimOffsetTicks", false, func(db *model.Database) {
			db.PlaylistItem[1].EndTrimOffsetTicks.Valid = true
			db.PlaylistItem[1].EndTrimOffsetTicks.Int64 = 900
		}},
		{"accuracy", false, func(db *model.Database) { db.PlaylistItem[1].Accuracy++ }},
		{"endAction", false, func(db *model.Database) { db.PlaylistItem[1].EndAction++ }},
		{"baseDurationTicks", false, func(db *model.Database) { db.PlaylistItemLocationMap[1].BaseDurationTicks.Int64++ }},
		{"majorMultimediaType", false, func(db *model.Database) { db.PlaylistItemLocationMap[1].MajorMultimediaType++ }},
		{"durationTicks", true, func(db *model.Database) { db.PlaylistItemIndependentMediaMap[1].DurationTicks++ }},
	}
	for _, tc := range tests {
		for _, side := range []string{"leftSide", "rightSide"} {
			t.Run(tc.field+"/"+side, func(t *testing.T) {
				label := "A"
				if tc.independent {
					label = "video"
				}
				fixture := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Settings", Items: []string{label}}})
				db, conflicts, run := playlistConflictInputs(t, fixture)
				require.Equal(t, 1, db.Stats("leftSide").Playlist)
				require.Equal(t, 1, db.Stats("leftSide").PlaylistItem)
				tc.change(db.right)
				require.Error(t, run())
				conflict, err := conflicts.NextConflict()
				require.NoError(t, err)
				fields, err := json.Marshal([]string{tc.field})
				require.NoError(t, err)
				require.Equal(t, string(fields), conflict.ChangedFields)
				// Both choices must carry the model identity and playlist context
				// required by the mobile conflict UI, across all three model kinds.
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
					require.Equal(t, "Settings", payload.Related.Tag.Name)
				}
				var selected struct {
					Model json.RawMessage `json:"model"`
				}
				payload := conflict.Left
				if side == "rightSide" {
					payload = conflict.Right
				}
				require.NoError(t, json.Unmarshal([]byte(payload), &selected))
				require.NoError(t, conflicts.SolveConflict(conflict.Key, side))
				require.NoError(t, run())
				require.NoError(t, db.MergeTagMaps())
				output := filepath.Join(t.TempDir(), "settings.jwlibrary")
				require.NoError(t, db.ExportMerged(output))
				require.NoError(t, db.ExportMerged(filepath.Join(t.TempDir(), "again.jwlibrary")))
				got := importedPlaylist(t, output)
				var actual model.Model
				var wanted model.Model
				switch tc.field {
				case "durationTicks":
					wanted = &model.PlaylistItemIndependentMediaMap{}
					for _, row := range got.PlaylistItemIndependentMediaMap {
						if row != nil {
							actual = row
							break
						}
					}
				case "baseDurationTicks", "majorMultimediaType":
					wanted = &model.PlaylistItemLocationMap{}
					for _, row := range got.PlaylistItemLocationMap {
						if row != nil {
							actual = row
							break
						}
					}
				default:
					wanted = &model.PlaylistItem{}
					actual = orderedItems(got)[0]
				}
				require.NoError(t, json.Unmarshal(selected.Model, wanted))
				require.Empty(t, model.PlaylistChangedFields(wanted, actual))
			})
		}
	}
}

func TestPlaylistThumbnailPresenceAndRepeatedInsertionAreCosmetic(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(map[bool]string{false: "forward", true: "reverse"}[reverse], func(t *testing.T) {
			short := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Repeated", Items: []string{"video", "video"}}})
			long := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Repeated", Items: []string{"video", "video", "video"}}})
			left, right := short, long
			if reverse {
				left, right = right, left
			}
			output := filepath.Join(t.TempDir(), "repeated.jwlibrary")
			mergePlaylistRegression(t, left, right, output, func(db *DatabaseWrapper) {
				missing := db.left
				if reverse {
					missing = db.right
				}
				for _, item := range missing.PlaylistItem {
					if item != nil {
						item.ThumbnailFilePath.Valid = false
					}
				}
			})
			got := importedPlaylist(t, output)
			require.Len(t, orderedItems(got), 3)
			for _, item := range orderedItems(got) {
				require.True(t, item.ThumbnailFilePath.Valid)
				_, err := got.PlaylistThumbnail(item.ID())
				require.NoError(t, err)
			}
			again := filepath.Join(t.TempDir(), "again.jwlibrary")
			mergePlaylistRegression(t, output, short, again, nil)
			require.Len(t, orderedItems(importedPlaylist(t, again)), 3)
		})
	}
}
