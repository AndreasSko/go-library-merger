package merger

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

func TestPlaylistContradictoryOrderKeepsLeftOrder(t *testing.T) {
	sequence := func(ids ...int32) []*model.TagMap {
		out := []*model.TagMap{nil}
		for pos, id := range ids {
			out = append(out, &model.TagMap{TagMapID: pos + 1, TagID: 1, PlaylistItemID: sql.NullInt32{Int32: id, Valid: true}, Position: pos})
		}
		return out
	}
	order := func(merged []*model.TagMap) []int32 {
		ids := []int32{}
		for _, tm := range merged {
			if tm != nil {
				ids = append(ids, tm.PlaylistItemID.Int32)
			}
		}
		return ids
	}
	merged, _, err := MergeTagMaps(sequence(1, 2), sequence(2, 1), nil)
	require.NoError(t, err)
	require.Equal(t, []int32{1, 2}, order(merged))

	// Item 3 is new on the right and stays after item 2, its left neighbor there.
	merged, _, err = MergeTagMaps(sequence(1, 2, 4), sequence(4, 2, 3, 1), nil)
	require.NoError(t, err)
	require.Equal(t, []int32{1, 2, 3, 4}, order(merged))
}

func TestCopyMediaUsesMatchingContent(t *testing.T) {
	left, right := &model.Database{WorkingDir: t.TempDir()}, &model.Database{WorkingDir: t.TempDir()}
	// Equal content A uses different paths. The right backup reuses the left
	// filename for B. Copying from the first path that exists would corrupt B.
	left.IndependentMedia = []*model.IndependentMedia{{FilePath: "foo", Hash: "A"}}
	right.IndependentMedia = []*model.IndependentMedia{{FilePath: "bar", Hash: "A"}, {FilePath: "foo", Hash: "B"}}
	require.NoError(t, os.WriteFile(filepath.Join(left.WorkingDir, "foo"), []byte("A"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(right.WorkingDir, "bar"), []byte("A"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(right.WorkingDir, "foo"), []byte("B"), 0600))
	out := t.TempDir()
	require.NoError(t, CopyMergedIndependentMedia(right.IndependentMedia, left, right, out))
	blob, err := os.ReadFile(filepath.Join(out, "foo"))
	require.NoError(t, err)
	require.Equal(t, "B", string(blob))
	// Left-only content must still copy when the other archive has no media.
	require.NoError(t, CopyMergedIndependentMedia(left.IndependentMedia, left, &model.Database{}, t.TempDir()))
}

func TestRemapAndCopyMediaPathCollision(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse=%t", reverse), func(t *testing.T) {
			leftDir, rightDir := t.TempDir(), t.TempDir()
			leftHash, rightHash := "bbbb", "aaaa"
			require.NoError(t, os.WriteFile(filepath.Join(leftDir, "clip.mp4"), []byte("left"), 0600))
			require.NoError(t, os.WriteFile(filepath.Join(rightDir, "clip.mp4"), []byte("right"), 0600))
			left := &model.Database{WorkingDir: leftDir, IndependentMedia: []*model.IndependentMedia{nil, {IndependentMediaID: 1, FilePath: "clip.mp4", Hash: leftHash}}}
			right := &model.Database{WorkingDir: rightDir, IndependentMedia: []*model.IndependentMedia{nil, {IndependentMediaID: 1, FilePath: "clip.mp4", Hash: rightHash}}}
			first, second := left, right
			if reverse {
				first, second = right, left
			}
			merged, _, err := MergeIndependentMedia(first.IndependentMedia, second.IndependentMedia)
			require.NoError(t, err)
			_, err = remapIndependentMediaPaths(merged)
			require.NoError(t, err)
			paths := map[string]string{}
			for _, media := range merged {
				if media != nil {
					paths[media.Hash] = media.FilePath
				}
			}
			require.Equal(t, "clip.mp4", paths[rightHash])
			digest := fmt.Sprintf("%x", sha256.Sum256([]byte(leftHash)))
			require.Equal(t, "clip-"+digest[:12]+".mp4", paths[leftHash])

			out := t.TempDir()
			require.NoError(t, CopyMergedIndependentMedia(merged, left, right, out))
			rightBytes, err := os.ReadFile(filepath.Join(out, paths[rightHash]))
			require.NoError(t, err)
			leftBytes, err := os.ReadFile(filepath.Join(out, paths[leftHash]))
			require.NoError(t, err)
			require.Equal(t, "right", string(rightBytes))
			require.Equal(t, "left", string(leftBytes))
		})
	}
}

func TestRemapMediaCaseInsensitivePaths(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		left := &model.Database{WorkingDir: t.TempDir(), IndependentMedia: []*model.IndependentMedia{nil, {IndependentMediaID: 1, FilePath: "Clip.JPG", Hash: "aaaa"}}}
		right := &model.Database{WorkingDir: t.TempDir(), IndependentMedia: []*model.IndependentMedia{nil, {IndependentMediaID: 1, FilePath: "clip.jpg", Hash: "bbbb"}}}
		require.NoError(t, os.WriteFile(filepath.Join(left.WorkingDir, "Clip.JPG"), []byte("left"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(right.WorkingDir, "clip.jpg"), []byte("right"), 0600))
		a, b := left, right
		if reverse {
			a, b = b, a
		}
		media, _, err := MergeIndependentMedia(a.IndependentMedia, b.IndependentMedia)
		require.NoError(t, err)
		_, err = remapIndependentMediaPaths(media)
		require.NoError(t, err)
		paths := map[string]string{}
		for _, m := range media {
			if m != nil {
				paths[m.Hash] = m.FilePath
			}
		}
		require.NotEqual(t, strings.ToLower(paths["aaaa"]), strings.ToLower(paths["bbbb"]))
		out := t.TempDir()
		require.NoError(t, CopyMergedIndependentMedia(media, a, b, out))
		blob, err := os.ReadFile(filepath.Join(out, paths["aaaa"]))
		require.NoError(t, err)
		require.Equal(t, "left", string(blob))
		blob, err = os.ReadFile(filepath.Join(out, paths["bbbb"]))
		require.NoError(t, err)
		require.Equal(t, "right", string(blob))
	}
}

func TestRemapMediaAvoidsBackupReservedFilenames(t *testing.T) {
	for _, path := range []string{"userData.db", "manifest.json", "default_thumbnail.png"} {
		media := []*model.IndependentMedia{{FilePath: path, Hash: "content"}}
		paths, err := remapIndependentMediaPaths(media)
		require.NoError(t, err)
		require.NotEqual(t, path, media[0].FilePath)
		require.Equal(t, media[0].FilePath, paths["content\x00"+path])
	}
}

func TestPlaylistSolutionUsesRenamedThumbnail(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		for _, side := range []MergeSide{LeftSide, RightSide} {
			t.Run(fmt.Sprintf("reverse%t/%s", reverse, side), func(t *testing.T) {
				fixture := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Original alias", Items: []string{"A"}}})
				left, right := &model.Database{TempDir: t.TempDir()}, &model.Database{TempDir: t.TempDir()}
				require.NoError(t, left.ImportJWLBackup(fixture))
				require.NoError(t, right.ImportJWLBackup(fixture))
				defer func() { require.NoError(t, left.Close()) }()
				defer func() { require.NoError(t, right.Close()) }()
				for _, alias := range []struct {
					db   *model.Database
					path string
				}{{left, "z.jpg"}, {right, "a.jpg"}} {
					media := alias.db.IndependentMedia[1]
					require.NoError(t, os.Rename(filepath.Join(alias.db.WorkingDir, media.FilePath), filepath.Join(alias.db.WorkingDir, alias.path)))
					media.FilePath = alias.path
					alias.db.PlaylistItem[1].ThumbnailFilePath.String = alias.path
					require.NoError(t, alias.db.LinkPlaylistItems())
				}
				right.PlaylistItem[1].Accuracy++
				want, err := left.PlaylistThumbnail(1)
				require.NoError(t, err)
				if reverse {
					left, right = right, left
				}
				merged := &model.Database{TempDir: t.TempDir()}
				defer func() { require.NoError(t, merged.Close()) }()
				locations, locationIDs, err := MergeLocations(left.Location, right.Location)
				require.NoError(t, err)
				merged.Location = locations
				UpdateLRIDs(left.PlaylistItemLocationMap, right.PlaylistItemLocationMap, "LocationID", locationIDs)
				merged.Tag, _, err = MergeTags(left.Tag, right.Tag, nil)
				require.NoError(t, err)
				var conflict MergeConflictError
				require.ErrorAs(t, MergePlaylists(left, right, merged, nil), &conflict)
				require.Len(t, conflict.Conflicts, 1)
				chosen, discarded := left.PlaylistItem[1], right.PlaylistItem[1]
				if side == RightSide {
					chosen, discarded = discarded, chosen
				}
				before := model.MakeModelCopy(chosen)
				solutions := map[string]MergeSolution{}
				for key := range conflict.Conflicts {
					solutions[key] = MergeSolution{Side: side, Solution: chosen, Discarded: discarded}
				}
				require.NoError(t, MergePlaylists(left, right, merged, solutions))
				require.Equal(t, before, chosen)
				require.Equal(t, chosen.Accuracy, merged.PlaylistItem[1].Accuracy)
				got, err := merged.PlaylistThumbnail(1)
				require.NoError(t, err)
				require.Equal(t, want, got)
				merged.TagMap, _, err = MergeTagMaps(left.TagMap, right.TagMap, nil)
				require.NoError(t, err)
				require.NoError(t, PrepareDatabasesPostMerge(merged))
				output := filepath.Join(t.TempDir(), "alias.jwlibrary")
				require.NoError(t, merged.ExportJWLBackup(output))
				playlisttest.Verify(t, output, []playlisttest.Playlist{{Name: "Original alias", Items: []string{"A"}}})
				readback := &model.Database{TempDir: t.TempDir()}
				require.NoError(t, readback.ImportJWLBackup(output))
				defer func() { require.NoError(t, readback.Close()) }()
				got, err = readback.PlaylistThumbnail(1)
				require.NoError(t, err)
				require.Equal(t, want, got)
				require.Equal(t, chosen.Accuracy, readback.PlaylistItem[1].Accuracy)
			})
		}
	}
}
