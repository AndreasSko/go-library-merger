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

func TestPlaylistThumbnailEditConflictResolvesAndRemerges(t *testing.T) {
	for _, samePath := range []bool{false, true} {
		for _, reverse := range []bool{false, true} {
			for _, side := range []string{"leftSide", "rightSide"} {
				t.Run(fmt.Sprintf("samePath=%t/reverse=%t/choose=%s", samePath, reverse, side), func(t *testing.T) {
					fixture := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Thumbnails", Items: []string{"A"}}})
					db := &DatabaseWrapper{TempDir: t.TempDir()}
					t.Cleanup(func() { require.NoError(t, db.Close()) })
					require.NoError(t, db.ImportJWLBackup(fixture, "leftSide"))
					require.NoError(t, db.ImportJWLBackup(fixture, "rightSide"))
					originalBlob, err := db.PlaylistThumbnail("leftSide", 1)
					require.NoError(t, err)
					changedBlob := []byte("different thumbnail for the same playlist item")
					media := db.right.IndependentMedia[1]
					if !samePath {
						media.FilePath = "changed.jpg"
						db.right.PlaylistItem[1].ThumbnailFilePath.String = media.FilePath
					}
					media.Hash = fmt.Sprintf("%x", sha256.Sum256(changedBlob))
					require.NoError(t, os.WriteFile(filepath.Join(db.right.IndependentMediaDir, media.FilePath), changedBlob, 0600))
					require.NoError(t, db.right.LinkPlaylistItems())
					if reverse {
						db.left, db.right = db.right, db.left
					}
					want := originalBlob
					if (side == "rightSide") != reverse {
						want = changedBlob
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
						if (entry.side == "rightSide") != reverse {
							expected = changedBlob
						}
						require.Equal(t, expected, blob)
					}
					require.NoError(t, conflicts.SolveConflict(conflict.Key, side))
					require.NoError(t, run())
					require.NoError(t, db.MergeTagMaps())
					output := filepath.Join(t.TempDir(), "resolved.jwlibrary")
					require.NoError(t, db.ExportMerged(output))
					playlisttest.Verify(t, output, []playlisttest.Playlist{{Name: "Thumbnails", Items: []string{"A"}}})
					got := importedPlaylist(t, output)
					item := orderedItems(got)[0]
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
	db := &DatabaseWrapper{}
	_, err := db.PlaylistThumbnail("unknown", 1)
	require.Error(t, err)
	_, err = db.PlaylistThumbnail("leftSide", 1)
	require.Error(t, err)
	fixture := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Thumbnails", Items: []string{"A"}}})
	require.NoError(t, db.ImportJWLBackup(fixture, "leftSide"))
	defer func() { require.NoError(t, db.Close()) }()
	_, err = db.PlaylistThumbnail("leftSide", -1)
	require.Error(t, err)
	_, err = db.PlaylistThumbnail("leftSide", 99)
	require.Error(t, err)
	db.left.PlaylistItem[1].ThumbnailFilePath.Valid = false
	blob, err := db.PlaylistThumbnail("leftSide", 1)
	require.NoError(t, err)
	require.True(t, len(blob) > 8)
	require.Equal(t, []byte{137, 80, 78, 71, 13, 10, 26, 10}, blob[:8])
	db.left.PlaylistItem[1].ThumbnailFilePath.String = "../outside"
	db.left.PlaylistItem[1].ThumbnailFilePath.Valid = true
	_, err = db.PlaylistThumbnail("leftSide", 1)
	require.Error(t, err)
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
