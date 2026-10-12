package merger

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

func TestPlaylistOriginalExplicitSolutionStillUsesCosmeticPolicy(t *testing.T) {
	for _, side := range []MergeSide{LeftSide, RightSide} {
		t.Run(string(side), func(t *testing.T) {
			fixture := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Original solution", Items: []string{"A"}}})
			left, right := &model.Database{TempDir: t.TempDir()}, &model.Database{TempDir: t.TempDir()}
			require.NoError(t, left.ImportJWLBackup(fixture))
			require.NoError(t, right.ImportJWLBackup(fixture))
			defer func() { require.NoError(t, left.CleanupMedia()) }()
			defer func() { require.NoError(t, right.CleanupMedia()) }()
			media := right.IndependentMedia[1]
			originalHash := left.IndependentMedia[1].UniqueKey()
			blob, err := left.PlaylistThumbnail(1)
			require.NoError(t, err)
			changed := []byte("other thumbnail for original solution")
			media.Hash = fmt.Sprintf("%x", sha256.Sum256(changed))
			require.NoError(t, os.WriteFile(filepath.Join(right.IndependentMediaDir, media.FilePath), changed, 0600))
			right.PlaylistItem[1].Accuracy++
			want := blob
			if media.UniqueKey() < originalHash {
				want = changed
			}
			merged := &model.Database{TempDir: t.TempDir()}
			defer func() { require.NoError(t, merged.CleanupMedia()) }()
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
		})
	}
}

func TestPlaylistOriginalExplicitSolutionKeepsRemappedEqualContentAlias(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		for _, side := range []MergeSide{LeftSide, RightSide} {
			t.Run(fmt.Sprintf("reverse%t/%s", reverse, side), func(t *testing.T) {
				fixture := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Original alias", Items: []string{"A"}}})
				left, right := &model.Database{TempDir: t.TempDir()}, &model.Database{TempDir: t.TempDir()}
				require.NoError(t, left.ImportJWLBackup(fixture))
				require.NoError(t, right.ImportJWLBackup(fixture))
				defer func() { require.NoError(t, left.CleanupMedia()) }()
				defer func() { require.NoError(t, right.CleanupMedia()) }()
				for _, alias := range []struct {
					db   *model.Database
					path string
				}{{left, "z.jpg"}, {right, "a.jpg"}} {
					media := alias.db.IndependentMedia[1]
					require.NoError(t, os.Rename(filepath.Join(alias.db.IndependentMediaDir, media.FilePath), filepath.Join(alias.db.IndependentMediaDir, alias.path)))
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
				defer func() { require.NoError(t, merged.CleanupMedia()) }()
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
				defer func() { require.NoError(t, readback.CleanupMedia()) }()
				got, err = readback.PlaylistThumbnail(1)
				require.NoError(t, err)
				require.Equal(t, want, got)
				require.Equal(t, chosen.Accuracy, readback.PlaylistItem[1].Accuracy)
			})
		}
	}
}
