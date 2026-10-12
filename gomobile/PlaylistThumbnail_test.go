package gomobile

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

func TestPlaylistThumbnailCosmeticPolicyResolvesAndRemerges(t *testing.T) {
	for _, tc := range []struct {
		name              string
		samePath, reverse bool
		// Empty side denotes a cosmetic-only change, with no conflict choice.
		side string
	}{
		{"different paths/forward/cosmetic", false, false, ""},
		{"different paths/forward/choose left", false, false, "leftSide"},
		{"different paths/forward/choose right", false, false, "rightSide"},
		{"different paths/reverse/cosmetic", false, true, ""},
		{"different paths/reverse/choose left", false, true, "leftSide"},
		{"different paths/reverse/choose right", false, true, "rightSide"},
		{"path collision/forward/cosmetic", true, false, ""},
		{"path collision/forward/choose left", true, false, "leftSide"},
		{"path collision/forward/choose right", true, false, "rightSide"},
		{"path collision/reverse/cosmetic", true, true, ""},
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
			require.NoError(t, os.WriteFile(filepath.Join(db.right.IndependentMediaDir, media.FilePath), changedBlob, 0600))
			if tc.side != "" {
				db.right.PlaylistItem[1].Accuracy = 2
			}
			require.NoError(t, db.right.LinkPlaylistItems())
			want := originalBlob
			if media.UniqueKey() < db.left.IndependentMedia[1].UniqueKey() {
				want = changedBlob
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
			mergePlaylistRegression(t, output, output, again, nil)
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
	mergePlaylistRegression(t, fixture, fixture, output, func(db *DatabaseWrapper) {
		media := db.right.IndependentMedia[1]
		require.NoError(t, os.Rename(filepath.Join(db.right.IndependentMediaDir, media.FilePath), filepath.Join(db.right.IndependentMediaDir, "renamed.jpg")))
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
		blob, err := os.ReadFile(filepath.Join(native.IndependentMediaDir, media.FilePath))
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
	mergePlaylistRegression(t, fixture, nativeArchive, output, nil)
	got := importedPlaylist(t, output)
	require.Len(t, orderedItems(got), 1)
	require.Len(t, got.IndependentMedia, len(original.IndependentMedia))
}

func TestPlaylistExportRequiresMembershipMerge(t *testing.T) {
	fixture := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Membership", Items: []string{"A"}}})
	db := &DatabaseWrapper{TempDir: t.TempDir()}
	defer func() { require.NoError(t, db.Close()) }()
	require.NoError(t, db.ImportJWLBackup(fixture, "leftSide"))
	require.NoError(t, db.ImportJWLBackup(fixture, "rightSide"))
	db.Init()
	require.NoError(t, db.MergeLocations())
	require.NoError(t, db.MergeTags())
	require.NoError(t, db.MergePlaylists("", nil))
	out := filepath.Join(t.TempDir(), "merged.jwlibrary")
	require.ErrorContains(t, db.ExportMerged(out), "MergeTagMaps")
	require.NoError(t, db.MergeTagMaps())
	require.NoError(t, db.ExportMerged(out))
	playlisttest.Verify(t, out, []playlisttest.Playlist{{Name: "Membership", Items: []string{"A"}}})
	// Replacing an input invalidates a previously completed merge as well.
	require.NoError(t, db.ImportJWLBackup(fixture, "leftSide"))
	require.ErrorContains(t, db.ExportMerged(out), "MergeTagMaps")
}
