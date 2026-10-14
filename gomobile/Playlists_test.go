package gomobile

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
	"github.com/AndreasSko/go-library-merger/model"
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
				out := filepath.Join(t.TempDir(), "merged.jwlibrary")
				mergePlaylistArchives(t, playlisttest.Build(t, left), playlisttest.Build(t, right), out, nil)
				playlisttest.Verify(t, out, tc.Want)
			})
		}
	}
}

// playlistConflictInputs imports both sides and provides the retryable merge
// stage used by conflict tests. Membership/export remain explicit at each call.
func playlistConflictInputs(t *testing.T, path string) (*DatabaseWrapper, *MergeConflictsWrapper, func() error) {
	t.Helper()
	db := &DatabaseWrapper{TempDir: t.TempDir()}
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, db.ImportJWLBackup(path, "leftSide"))
	require.NoError(t, db.ImportJWLBackup(path, "rightSide"))
	conflicts := &MergeConflictsWrapper{}
	conflicts.InitDBWrapper(db)
	run := func() error {
		require.NoError(t, db.Init())
		require.NoError(t, db.MergeLocations())
		require.NoError(t, db.MergeTags())
		return db.MergePlaylists("", conflicts)
	}
	return db, conflicts, run
}

// mergePlaylistArchives runs all merge steps like the app does.
func mergePlaylistArchives(t *testing.T, left, right, output string, mutate func(*DatabaseWrapper)) {
	t.Helper()
	db := &DatabaseWrapper{TempDir: t.TempDir()}
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, db.ImportJWLBackup(left, "leftSide"))
	require.NoError(t, db.ImportJWLBackup(right, "rightSide"))
	if mutate != nil {
		mutate(db)
	}
	require.NoError(t, db.Init())
	conflicts := &MergeConflictsWrapper{}
	require.NoError(t, db.MergeLocations())
	require.NoError(t, db.MergeBookmarks("", conflicts))
	require.NoError(t, db.MergeInputFields("", conflicts))
	require.NoError(t, db.MergeTags())
	require.NoError(t, db.MergeUserMarkAndBlockRange("", conflicts))
	require.NoError(t, db.MergeNotes("", conflicts))
	require.NoError(t, db.MergePlaylists("", conflicts))
	require.NoError(t, db.MergeTagMaps())
	require.NoError(t, db.ExportMerged(output))
}

// reimportSide exports a side and imports it again, so changes made to it in
// memory are validated like a real backup.
func reimportSide(t *testing.T, db *DatabaseWrapper, side string) {
	t.Helper()
	source := db.left
	if side == "rightSide" {
		source = db.right
	}
	path := filepath.Join(t.TempDir(), "side.jwlibrary")
	require.NoError(t, source.ExportJWLBackup(path))
	require.NoError(t, db.ImportJWLBackup(path, side))
}

func importedPlaylist(t *testing.T, path string) *model.Database {
	t.Helper()
	db := &model.Database{TempDir: t.TempDir()}
	require.NoError(t, db.ImportJWLBackup(path))
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return db
}

func orderedItems(db *model.Database) []*model.PlaylistItem {
	maps := []*model.TagMap{}
	for _, row := range db.TagMap {
		if row != nil && row.PlaylistItemID.Valid {
			maps = append(maps, row)
		}
	}
	sort.Slice(maps, func(i, j int) bool { return maps[i].Position < maps[j].Position })
	items := make([]*model.PlaylistItem, 0, len(maps))
	for _, row := range maps {
		items = append(items, db.PlaylistItem[row.PlaylistItemID.Int32])
	}
	return items
}

func TestPlaylistRepeatedOccurrenceInsertionExportsAndRemerges(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse=%t", reverse), func(t *testing.T) {
			two := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Repeated", Items: []string{"video", "video"}}})
			three := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Repeated", Items: []string{"video", "video", "video"}}})
			left, right := two, three
			if reverse {
				left, right = right, left
			}
			output := filepath.Join(t.TempDir(), "merged.jwlibrary")
			mergePlaylistArchives(t, left, right, output, func(db *DatabaseWrapper) {
				short, long := db.left, db.right
				if len(short.PlaylistItem) > len(long.PlaylistItem) {
					short, long = long, short
				}
				short.PlaylistItem[1].StartTrimOffsetTicks.Int64 = 0
				short.PlaylistItem[2].StartTrimOffsetTicks.Int64 = 20
				long.PlaylistItem[1].StartTrimOffsetTicks.Int64 = 0
				long.PlaylistItem[2].StartTrimOffsetTicks.Int64 = 10
				long.PlaylistItem[2].EndAction = 1
				long.PlaylistItem[3].StartTrimOffsetTicks.Int64 = 20
			})
			got := importedPlaylist(t, output)
			items := orderedItems(got)
			require.Len(t, items, 3)
			require.Equal(t, []int64{0, 10, 20}, []int64{
				items[0].StartTrimOffsetTicks.Int64,
				items[1].StartTrimOffsetTicks.Int64,
				items[2].StartTrimOffsetTicks.Int64,
			})
			again := filepath.Join(t.TempDir(), "again.jwlibrary")
			mergePlaylistArchives(t, output, output, again, nil)
			repeated := orderedItems(importedPlaylist(t, again))
			require.Len(t, repeated, 3)
			for i := range items {
				require.Equal(t, items[i].StartTrimOffsetTicks, repeated[i].StartTrimOffsetTicks)
				require.Equal(t, items[i].EndAction, repeated[i].EndAction)
			}
			require.Equal(t, 1, items[1].EndAction)
		})
	}
}

func TestPlaylistMediaPathCollisionExportsAndRemerges(t *testing.T) {
	var firstPaths []string
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse=%t", reverse), func(t *testing.T) {
			first := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Media", Items: []string{"video"}}})
			second := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Media", Items: []string{"video"}}})
			left, right := first, second
			if reverse {
				left, right = right, left
			}
			output := filepath.Join(t.TempDir(), "merged.jwlibrary")
			mergePlaylistArchives(t, left, right, output, func(db *DatabaseWrapper) {
				side, source := "rightSide", db.right
				if reverse {
					side, source = "leftSide", db.left
				}
				media := source.IndependentMedia[2]
				payload := []byte("different video payload")
				require.NoError(t, os.WriteFile(filepath.Join(source.WorkingDir, media.FilePath), payload, 0600))
				media.Hash = fmt.Sprintf("%x", sha256.Sum256(payload))
				reimportSide(t, db, side)
			})
			got := importedPlaylist(t, output)
			require.Len(t, orderedItems(got), 2)
			paths, hashes := map[string]bool{}, map[string]bool{}
			hashByID := map[int]string{}
			for _, media := range got.IndependentMedia {
				if media != nil {
					hashByID[media.IndependentMediaID] = media.Hash
				}
				if media == nil || media.OriginalFilename != "video.bin" {
					continue
				}
				paths[media.FilePath] = true
				blob, err := os.ReadFile(filepath.Join(got.WorkingDir, media.FilePath))
				require.NoError(t, err)
				hash := fmt.Sprintf("%x", sha256.Sum256(blob))
				require.Equal(t, media.Hash, hash)
				hashes[hash] = true
			}
			require.Len(t, paths, 2)
			require.Len(t, hashes, 2)
			pathList := make([]string, 0, len(paths))
			for path := range paths {
				pathList = append(pathList, path)
			}
			sort.Strings(pathList)
			if firstPaths == nil {
				firstPaths = pathList
			} else {
				require.Equal(t, firstPaths, pathList, "remapped paths must not depend on input direction")
			}
			linkedHashes := map[string]bool{}
			for _, link := range got.PlaylistItemIndependentMediaMap {
				if link != nil {
					linkedHashes[hashByID[link.IndependentMediaID]] = true
				}
			}
			require.Equal(t, hashes, linkedHashes)
			for _, item := range orderedItems(got) {
				require.True(t, item.ThumbnailFilePath.Valid)
				_, err := os.Stat(filepath.Join(got.WorkingDir, item.ThumbnailFilePath.String))
				require.NoError(t, err)
			}
			again := filepath.Join(t.TempDir(), "again.jwlibrary")
			mergePlaylistArchives(t, output, output, again, nil)
			remerged := importedPlaylist(t, again)
			require.Len(t, orderedItems(remerged), 2)
			remergedPaths := []string{}
			for _, media := range remerged.IndependentMedia {
				if media != nil && media.OriginalFilename == "video.bin" {
					remergedPaths = append(remergedPaths, media.FilePath)
				}
			}
			sort.Strings(remergedPaths)
			require.Equal(t, pathList, remergedPaths, "self-remerge must not churn media paths")
		})
	}
}

func TestPlaylistThumbnailPathCollisionKeepsEachPayload(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse=%t", reverse), func(t *testing.T) {
			a := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Media", Items: []string{"A"}}})
			b := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Media", Items: []string{"B"}}})
			left, right := a, b
			if reverse {
				left, right = right, left
			}
			output := filepath.Join(t.TempDir(), "merged.jwlibrary")
			expected := map[string]string{}
			mergePlaylistArchives(t, left, right, output, func(db *DatabaseWrapper) {
				for _, source := range []*model.Database{db.left, db.right} {
					item := source.PlaylistItem[1]
					media := source.IndependentMedia[1]
					if item.Label == "B" {
						require.NoError(t, os.Rename(
							filepath.Join(source.WorkingDir, media.FilePath),
							filepath.Join(source.WorkingDir, "A.jpg"),
						))
						media.FilePath = "A.jpg"
						item.ThumbnailFilePath.String = "A.jpg"
					}
					expected[item.Label] = media.Hash
				}
			})
			got := importedPlaylist(t, output)
			hashByPath := map[string]string{}
			for _, media := range got.IndependentMedia {
				if media != nil {
					hashByPath[media.FilePath] = media.Hash
				}
			}
			for _, item := range orderedItems(got) {
				require.Equal(t, expected[item.Label], hashByPath[item.ThumbnailFilePath.String])
				blob, err := os.ReadFile(filepath.Join(got.WorkingDir, item.ThumbnailFilePath.String))
				require.NoError(t, err)
				require.Equal(t, expected[item.Label], fmt.Sprintf("%x", sha256.Sum256(blob)))
			}
		})
	}
}

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

func TestPlaylistMissingThumbnailIsNoConflict(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(map[bool]string{false: "forward", true: "reverse"}[reverse], func(t *testing.T) {
			short := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Repeated", Items: []string{"video", "video"}}})
			long := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Repeated", Items: []string{"video", "video", "video"}}})
			left, right := short, long
			if reverse {
				left, right = right, left
			}
			output := filepath.Join(t.TempDir(), "repeated.jwlibrary")
			mergePlaylistArchives(t, left, right, output, func(db *DatabaseWrapper) {
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
				_, err := got.PlaylistThumbnail(item.ID())
				require.NoError(t, err)
			}
			again := filepath.Join(t.TempDir(), "again.jwlibrary")
			mergePlaylistArchives(t, output, short, again, nil)
			require.Len(t, orderedItems(importedPlaylist(t, again)), 3)
		})
	}
}

func TestPlaylistThumbnailComesFromChosenSide(t *testing.T) {
	for _, tc := range []struct {
		name              string
		samePath, reverse bool
		// Empty side means only the thumbnail differs, so there is no conflict.
		side string
	}{
		{"different paths/forward/thumbnail only", false, false, ""},
		{"different paths/forward/choose left", false, false, "leftSide"},
		{"different paths/forward/choose right", false, false, "rightSide"},
		{"different paths/reverse/thumbnail only", false, true, ""},
		{"different paths/reverse/choose left", false, true, "leftSide"},
		{"different paths/reverse/choose right", false, true, "rightSide"},
		{"path collision/forward/thumbnail only", true, false, ""},
		{"path collision/forward/choose left", true, false, "leftSide"},
		{"path collision/forward/choose right", true, false, "rightSide"},
		{"path collision/reverse/thumbnail only", true, true, ""},
		{"path collision/reverse/choose left", true, true, "leftSide"},
		{"path collision/reverse/choose right", true, true, "rightSide"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Thumbnails", Items: []string{"A"}}})
			db, conflicts, run := playlistConflictInputs(t, fixture)
			originalBlob, err := db.PlaylistThumbnail("leftSide", 1)
			require.NoError(t, err)
			changedBlob := []byte("different thumbnail for the same playlist item")
			media := db.right.IndependentMedia[1]
			if !tc.samePath {
				media.FilePath = "changed.jpg"
				db.right.PlaylistItem[1].ThumbnailFilePath.String = media.FilePath
			}
			media.Hash = fmt.Sprintf("%x", sha256.Sum256(changedBlob))
			require.NoError(t, os.WriteFile(filepath.Join(db.right.WorkingDir, media.FilePath), changedBlob, 0600))
			if tc.side != "" {
				db.right.PlaylistItem[1].Accuracy = 2
			}
			reimportSide(t, db, "rightSide")
			leftBlob, rightBlob := originalBlob, changedBlob
			if tc.reverse {
				leftBlob, rightBlob = rightBlob, leftBlob
			}
			want := leftBlob
			if tc.side == "rightSide" {
				want = rightBlob
			}
			if tc.reverse {
				db.left, db.right = db.right, db.left
			}
			if tc.side != "" {
				require.Error(t, run())
				conflict, err := conflicts.NextConflict()
				require.NoError(t, err)
				require.JSONEq(t, `["accuracy"]`, conflict.ChangedFields)
				for _, entry := range []struct{ side, payload string }{{"leftSide", conflict.Left}, {"rightSide", conflict.Right}} {
					var tuple struct {
						Model struct {
							Type string `json:"type"`
							ID   int    `json:"playlistItemId"`
						} `json:"model"`
					}
					require.NoError(t, json.Unmarshal([]byte(entry.payload), &tuple))
					require.Equal(t, "PlaylistItem", tuple.Model.Type)
					blob, err := db.PlaylistThumbnail(entry.side, tuple.Model.ID)
					require.NoError(t, err)
					expected := originalBlob
					if (entry.side == "rightSide") != tc.reverse {
						expected = changedBlob
					}
					require.Equal(t, expected, blob)
				}
				require.NoError(t, conflicts.SolveConflict(conflict.Key, tc.side))
				require.NoError(t, run())
			} else {
				require.NoError(t, run())
				_, err := conflicts.NextConflict()
				require.Error(t, err)
			}
			require.NoError(t, db.MergeTagMaps())
			output := filepath.Join(t.TempDir(), "resolved.jwlibrary")
			require.NoError(t, db.ExportMerged(output))
			playlisttest.Verify(t, output, []playlisttest.Playlist{{Name: "Thumbnails", Items: []string{"A"}}})
			got := importedPlaylist(t, output)
			item := orderedItems(got)[0]
			expectedAccuracy := 1
			if tc.side != "" && (tc.side == "rightSide") != tc.reverse {
				expectedAccuracy = 2
			}
			require.Equal(t, expectedAccuracy, item.Accuracy)
			blob, err := got.PlaylistThumbnail(item.ID())
			require.NoError(t, err)
			require.Equal(t, want, blob)
			again := filepath.Join(t.TempDir(), "again.jwlibrary")
			mergePlaylistArchives(t, output, output, again, nil)
			remerged := importedPlaylist(t, again)
			require.Len(t, orderedItems(remerged), 1)
			blob, err = remerged.PlaylistThumbnail(orderedItems(remerged)[0].ID())
			require.NoError(t, err)
			require.Equal(t, want, blob)
		})
	}
}

func TestPlaylistThumbnailSameContentDifferentPathHasNoConflict(t *testing.T) {
	fixture := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Thumbnails", Items: []string{"A"}}})
	output := filepath.Join(t.TempDir(), "merged.jwlibrary")
	mergePlaylistArchives(t, fixture, fixture, output, func(db *DatabaseWrapper) {
		media := db.right.IndependentMedia[1]
		require.NoError(t, os.Rename(filepath.Join(db.right.WorkingDir, media.FilePath), filepath.Join(db.right.WorkingDir, "renamed.jpg")))
		media.FilePath = "renamed.jpg"
		db.right.PlaylistItem[1].ThumbnailFilePath.String = media.FilePath
		require.NoError(t, db.right.LinkPlaylistItems())
	})
	playlisttest.Verify(t, output, []playlisttest.Playlist{{Name: "Thumbnails", Items: []string{"A"}}})
}

func TestPlaylistThumbnailAPIRejectsInvalidInputAndUsesDefault(t *testing.T) {
	db := &DatabaseWrapper{TempDir: t.TempDir()}
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	for _, tc := range []struct{ name, side string }{
		{"unknown side", "unknown"},
		{"unloaded side", "leftSide"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.PlaylistThumbnail(tc.side, 1)
			require.Error(t, err)
		})
	}
	fixture := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Thumbnails", Items: []string{"A"}}})
	require.NoError(t, db.ImportJWLBackup(fixture, "leftSide"))
	for _, tc := range []struct {
		name             string
		id               int
		path             string
		valid, wantError bool
	}{
		{"negative item ID", -1, "", false, true},
		{"missing item ID", 99, "", false, true},
		{"default thumbnail", 1, "", false, false},
		{"path traversal", 1, "../outside", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db.left.PlaylistItem[1].ThumbnailFilePath.String = tc.path
			db.left.PlaylistItem[1].ThumbnailFilePath.Valid = tc.valid
			blob, err := db.PlaylistThumbnail("leftSide", tc.id)
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Greater(t, len(blob), 8)
			require.Equal(t, []byte{137, 80, 78, 71, 13, 10, 26, 10}, blob[:8])
		})
	}
}

// Imported validation must recognize native JW Library's unpadded byte hex
// while matching the same payload against standard SHA-256 encodings.
func TestPlaylistNativeMediaHashesMatchCanonicalContent(t *testing.T) {
	fixture := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Native", Items: []string{"video"}}})
	original := importedPlaylist(t, fixture)
	native := model.MakeDatabaseCopy(original)
	different := false
	for _, media := range native.IndependentMedia {
		if media == nil {
			continue
		}
		blob, err := os.ReadFile(filepath.Join(native.WorkingDir, media.FilePath))
		require.NoError(t, err)
		digest := sha256.Sum256(blob)
		hash := ""
		for _, value := range digest {
			hash += fmt.Sprintf("%x", value)
		}
		different = different || hash != media.Hash
		media.Hash = hash
	}
	require.True(t, different, "fixture must exercise an unpadded hash")
	nativeArchive := filepath.Join(t.TempDir(), "native.jwlibrary")
	require.NoError(t, native.ExportJWLBackup(nativeArchive))
	loaded := importedPlaylist(t, nativeArchive)
	for i, media := range loaded.IndependentMedia {
		if media != nil {
			require.Equal(t, native.IndependentMedia[i].Hash, media.Hash)
			require.Equal(t, original.IndependentMedia[i].UniqueKey(), media.UniqueKey())
		}
	}
	output := filepath.Join(t.TempDir(), "merged.jwlibrary")
	mergePlaylistArchives(t, fixture, nativeArchive, output, nil)
	got := importedPlaylist(t, output)
	require.Len(t, orderedItems(got), 1)
	require.Len(t, got.IndependentMedia, len(original.IndependentMedia))
}

func TestPlaylistMixedHashEncodingsPreserveRepeatedOccurrences(t *testing.T) {
	for _, count := range []int{2, 3} {
		for _, nativeFirst := range []bool{false, true} {
			t.Run(fmt.Sprintf("count%d_nativeFirst%t", count, nativeFirst), func(t *testing.T) {
				items := make([]string, count)
				for i := range items {
					items[i] = "video"
				}
				want := []playlisttest.Playlist{{Name: "Mixed hashes", Items: items}}
				canonicalArchive := playlisttest.Build(t, want)
				mixed := model.MakeDatabaseCopy(importedPlaylist(t, canonicalArchive))
				originalID := mixed.PlaylistItemIndependentMediaMap[1].IndependentMediaID
				original := mixed.IndependentMedia[originalID]
				blob, err := os.ReadFile(filepath.Join(mixed.WorkingDir, original.FilePath))
				require.NoError(t, err)
				digest := sha256.Sum256(blob)
				native := ""
				for _, value := range digest {
					native += fmt.Sprintf("%x", value)
				}
				require.NotEqual(t, original.Hash, native, "fixture must exercise different encodings")
				duplicate := model.MakeModelCopy(original).(*model.IndependentMedia)
				duplicate.IndependentMediaID = len(mixed.IndependentMedia)
				duplicate.FilePath = "same-content.bin"
				duplicate.OriginalFilename = duplicate.FilePath
				duplicate.Hash = native
				require.NoError(t, os.WriteFile(filepath.Join(mixed.WorkingDir, duplicate.FilePath), blob, 0600))
				mixed.IndependentMedia = append(mixed.IndependentMedia, duplicate)
				for i, mapping := range mixed.PlaylistItemIndependentMediaMap {
					if mapping != nil && (i%2 == 1) == nativeFirst {
						mapping.IndependentMediaID = duplicate.IndependentMediaID
					}
				}
				mixedArchive := filepath.Join(t.TempDir(), "mixed.jwlibrary")
				require.NoError(t, mixed.ExportJWLBackup(mixedArchive))
				loaded := importedPlaylist(t, mixedArchive)
				require.Equal(t, original.Hash, loaded.IndependentMedia[originalID].Hash)
				require.Equal(t, native, loaded.IndependentMedia[duplicate.ID()].Hash)
				require.Equal(t, loaded.IndependentMedia[originalID].UniqueKey(), loaded.IndependentMedia[duplicate.ID()].UniqueKey())
				keys := map[string]bool{}
				for _, item := range loaded.PlaylistItem {
					if item != nil {
						require.False(t, keys[item.UniqueKey()], "each intentional repeat needs a distinct identity")
						keys[item.UniqueKey()] = true
					}
				}
				require.Len(t, keys, count)
				for name, inputs := range map[string][2]string{
					"self":        {mixedArchive, mixedArchive},
					"mixed-left":  {mixedArchive, canonicalArchive},
					"mixed-right": {canonicalArchive, mixedArchive},
				} {
					t.Run(name, func(t *testing.T) {
						output := filepath.Join(t.TempDir(), "merged.jwlibrary")
						mergePlaylistArchives(t, inputs[0], inputs[1], output, nil)
						playlisttest.Verify(t, output, want)
						// A second import and self-merge must remain stable after media rows
						// are deduplicated and item/membership IDs are remapped.
						remerged := filepath.Join(t.TempDir(), "remerged.jwlibrary")
						mergePlaylistArchives(t, output, output, remerged, nil)
						playlisttest.Verify(t, remerged, want)
					})
				}
			})
		}
	}
}

func TestPlaylistConflictSourceURLUsesTheOriginalSideMapping(t *testing.T) {
	location := &model.Location{LocationType: 3,
		KeySymbol:    sql.NullString{String: "jwb-142", Valid: true},
		Track:        sql.NullInt32{Int32: 1, Valid: true},
		MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}}
	related := model.Related{Location: location}
	item := &model.PlaylistItem{PlaylistItemID: 19}
	source := &model.Database{PlaylistItemLocationMap: []*model.PlaylistItemLocationMap{
		nil, {PlaylistItemID: 1, MajorMultimediaType: 1}, {PlaylistItemID: 19, MajorMultimediaType: 2},
	}}
	want := "https://www.jw.org/finder?lank=pub-jwb-142_1_VIDEO&wtlocale=E"
	ambiguous := model.MakeDatabaseCopy(source)
	ambiguous.PlaylistItemLocationMap = append(ambiguous.PlaylistItemLocationMap, &model.PlaylistItemLocationMap{PlaylistItemID: 19, MajorMultimediaType: 2})
	for _, tc := range []struct {
		name    string
		item    model.Model
		related model.Related
		source  *model.Database
		want    string
	}{
		{"original side mapping", item, related, source, want},
		{"missing source", item, related, nil, ""},
		{"audio mapping", &model.PlaylistItem{PlaylistItemID: 1}, related, source, ""},
		{"missing mapping", &model.PlaylistItem{PlaylistItemID: 99}, related, source, ""},
		{"ambiguous mapping", item, related, ambiguous, ""},
		{"location conflict carries its own mapping", &model.PlaylistItemLocationMap{PlaylistItemID: 500, MajorMultimediaType: 2}, related, nil, want},
		{"audio location conflict", &model.PlaylistItemLocationMap{MajorMultimediaType: 1}, related, nil, ""},
		{"unrelated model", &model.Note{}, related, source, ""},
		{"missing location", item, model.Related{}, source, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, playlistConflictSourceURL(tc.item, tc.related, tc.source))
		})
	}
}
