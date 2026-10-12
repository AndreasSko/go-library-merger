package gomobile

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

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
		db.Init()
		require.NoError(t, db.MergeLocations())
		require.NoError(t, db.MergeTags())
		return db.MergePlaylists("", conflicts)
	}
	return db, conflicts, run
}

func mergePlaylistRegression(t *testing.T, left, right, output string, mutate func(*DatabaseWrapper)) {
	t.Helper()
	db := &DatabaseWrapper{TempDir: t.TempDir()}
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, db.ImportJWLBackup(left, "leftSide"))
	require.NoError(t, db.ImportJWLBackup(right, "rightSide"))
	if mutate != nil {
		mutate(db)
	}
	db.Init()
	require.NoError(t, db.MergeLocations())
	require.NoError(t, db.MergeTags())
	require.NoError(t, db.MergePlaylists("", &MergeConflictsWrapper{}))
	require.NoError(t, db.MergeTagMaps())
	require.NoError(t, db.ExportMerged(output))
}

func importedPlaylist(t *testing.T, path string) *model.Database {
	t.Helper()
	db := &model.Database{TempDir: t.TempDir()}
	require.NoError(t, db.ImportJWLBackup(path))
	t.Cleanup(func() { require.NoError(t, db.CleanupMedia()) })
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
			mergePlaylistRegression(t, left, right, output, func(db *DatabaseWrapper) {
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
			mergePlaylistRegression(t, output, output, again, nil)
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
			mergePlaylistRegression(t, left, right, output, func(db *DatabaseWrapper) {
				source := db.right
				if reverse {
					source = db.left
				}
				media := source.IndependentMedia[2]
				payload := []byte("different video payload")
				require.NoError(t, os.WriteFile(filepath.Join(source.IndependentMediaDir, media.FilePath), payload, 0600))
				media.Hash = fmt.Sprintf("%x", sha256.Sum256(payload))
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
				blob, err := os.ReadFile(filepath.Join(got.IndependentMediaDir, media.FilePath))
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
				_, err := os.Stat(filepath.Join(got.IndependentMediaDir, item.ThumbnailFilePath.String))
				require.NoError(t, err)
			}
			again := filepath.Join(t.TempDir(), "again.jwlibrary")
			mergePlaylistRegression(t, output, output, again, nil)
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
			mergePlaylistRegression(t, left, right, output, func(db *DatabaseWrapper) {
				for _, source := range []*model.Database{db.left, db.right} {
					item := source.PlaylistItem[1]
					media := source.IndependentMedia[1]
					if item.Label == "B" {
						require.NoError(t, os.Rename(
							filepath.Join(source.IndependentMediaDir, media.FilePath),
							filepath.Join(source.IndependentMediaDir, "A.jpg"),
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
				blob, err := os.ReadFile(filepath.Join(got.IndependentMediaDir, item.ThumbnailFilePath.String))
				require.NoError(t, err)
				require.Equal(t, expected[item.Label], fmt.Sprintf("%x", sha256.Sum256(blob)))
			}
		})
	}
}
