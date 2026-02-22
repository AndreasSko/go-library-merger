package gomobile

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

// Native and canonical encodings can coexist in an archive. They describe the
// same content, so occurrences must be assigned after payload validation has
// established canonical content identities, before any merge deduplicates them.
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
				blob, err := os.ReadFile(filepath.Join(mixed.IndependentMediaDir, original.FilePath))
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
				require.NoError(t, os.WriteFile(filepath.Join(mixed.IndependentMediaDir, duplicate.FilePath), blob, 0600))
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
						mergePlaylistRegression(t, inputs[0], inputs[1], output, nil)
						playlisttest.Verify(t, output, want)
						// A second import and self-merge must remain stable after media rows
						// are deduplicated and item/membership IDs are remapped.
						remerged := filepath.Join(t.TempDir(), "remerged.jwlibrary")
						mergePlaylistRegression(t, output, output, remerged, nil)
						playlisttest.Verify(t, remerged, want)
					})
				}
			})
		}
	}
}
