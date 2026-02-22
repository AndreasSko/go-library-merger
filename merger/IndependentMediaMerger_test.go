package merger

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

func TestCopyMediaUsesMatchingContent(t *testing.T) {
	left, right := &model.Database{IndependentMediaDir: t.TempDir()}, &model.Database{IndependentMediaDir: t.TempDir()}
	// Equal content A uses different paths. The right backup reuses the left
	// filename for B. Copying from the first path that exists would corrupt B.
	left.IndependentMedia = []*model.IndependentMedia{{FilePath: "foo", Hash: "A"}}
	right.IndependentMedia = []*model.IndependentMedia{{FilePath: "bar", Hash: "A"}, {FilePath: "foo", Hash: "B"}}
	require.NoError(t, os.WriteFile(filepath.Join(left.IndependentMediaDir, "foo"), []byte("A"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(right.IndependentMediaDir, "bar"), []byte("A"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(right.IndependentMediaDir, "foo"), []byte("B"), 0600))
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
			left := &model.Database{IndependentMediaDir: leftDir, IndependentMedia: []*model.IndependentMedia{nil, {IndependentMediaID: 1, FilePath: "clip.mp4", Hash: leftHash}}}
			right := &model.Database{IndependentMediaDir: rightDir, IndependentMedia: []*model.IndependentMedia{nil, {IndependentMediaID: 1, FilePath: "clip.mp4", Hash: rightHash}}}
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

func TestRemapMediaAvoidsReservedGeneratedPath(t *testing.T) {
	hash := "bbbb"
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(hash)))
	reserved := "clip-" + digest[:12] + ".mp4"
	media := []*model.IndependentMedia{
		{FilePath: "clip.mp4", Hash: "aaaa"},
		{FilePath: "clip.mp4", Hash: hash},
		{FilePath: reserved, Hash: "cccc"},
	}
	_, err := remapIndependentMediaPaths(media)
	require.NoError(t, err)
	paths := []string{media[0].FilePath, media[1].FilePath, media[2].FilePath}
	sort.Strings(paths)
	require.Equal(t, []string{reserved, "clip-" + digest[:16] + ".mp4", "clip.mp4"}, paths)
}

func TestRemapMediaCaseInsensitivePaths(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		left := &model.Database{IndependentMediaDir: t.TempDir(), IndependentMedia: []*model.IndependentMedia{nil, {IndependentMediaID: 1, FilePath: "Clip.JPG", Hash: "aaaa"}}}
		right := &model.Database{IndependentMediaDir: t.TempDir(), IndependentMedia: []*model.IndependentMedia{nil, {IndependentMediaID: 1, FilePath: "clip.jpg", Hash: "bbbb"}}}
		require.NoError(t, os.WriteFile(filepath.Join(left.IndependentMediaDir, "Clip.JPG"), []byte("left"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(right.IndependentMediaDir, "clip.jpg"), []byte("right"), 0600))
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
