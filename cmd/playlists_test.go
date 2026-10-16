package cmd

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
	"github.com/AndreasSko/go-library-merger/merger"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

func TestPlaylistArchives(t *testing.T) {
	for _, tc := range playlisttest.Cases() {
		// The shared Gomobile table covers individual ordering and union cases.
		// Exercise the CLI archive pipeline with combined changes and repeats.
		if tc.Name != "common-plus-new" && tc.Name != "repeat" {
			continue
		}
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
				a, b := playlisttest.Build(t, left), playlisttest.Build(t, right)
				dest := filepath.Join(t.TempDir(), "merged.jwlibrary")
				output, err := os.Create(filepath.Join(t.TempDir(), "output.log"))
				require.NoError(t, err)
				defer func() { require.NoError(t, output.Close()) }()
				require.NoError(t, merge(a, b, dest, terminal.Stdio{In: os.Stdin, Out: output, Err: output}))
				playlisttest.Verify(t, dest, tc.Want)
				// Self-merge and merging with empty input must preserve all rows and media.
				again := filepath.Join(t.TempDir(), "again.jwlibrary")
				require.NoError(t, merge(dest, dest, again, terminal.Stdio{In: os.Stdin, Out: output, Err: output}))
				playlisttest.Verify(t, again, tc.Want)
				empty := playlisttest.Build(t, nil)
				require.NoError(t, merge(dest, empty, again, terminal.Stdio{In: os.Stdin, Out: output, Err: output}))
				playlisttest.Verify(t, again, tc.Want)
			})
		}
	}
}

// thumbnailPlaylistArchives returns two backups whose only playlist item has
// different thumbnails, and the thumbnail of each side.
func thumbnailPlaylistArchives(t *testing.T, settings bool) (string, string, []byte, []byte) {
	t.Helper()
	left := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Settings", Items: []string{"A"}}})
	db := &model.Database{TempDir: t.TempDir()}
	require.NoError(t, db.ImportJWLBackup(left))
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	media := db.IndependentMedia[1]
	original, err := os.ReadFile(filepath.Join(db.WorkingDir, media.FilePath))
	require.NoError(t, err)
	changed := []byte("changed CLI thumbnail")
	media.Hash = fmt.Sprintf("%x", sha256.Sum256(changed))
	require.NoError(t, os.WriteFile(filepath.Join(db.WorkingDir, media.FilePath), changed, 0600))
	if settings {
		db.PlaylistItem[1].Accuracy = 2
	}
	right := filepath.Join(t.TempDir(), "right.jwlibrary")
	require.NoError(t, db.ExportJWLBackup(right))
	return left, right, original, changed
}

// playlistCLI gives each case isolated I/O and restores the command flag.
func playlistCLI(t *testing.T, resolver string) (terminal.Stdio, string) {
	t.Helper()
	previous := PlaylistResolver
	PlaylistResolver = resolver
	t.Cleanup(func() { PlaylistResolver = previous })
	stdin, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stdin.Close()) })
	log, err := os.Create(filepath.Join(t.TempDir(), "log.txt"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, log.Close()) })
	return terminal.Stdio{In: stdin, Out: log, Err: log}, log.Name()
}

func TestPlaylistCLIResolvers(t *testing.T) {
	for _, tc := range []struct {
		name, resolver string
		settings       bool
	}{
		{"automatic thumbnail", "", false},
		{"choose left settings", "chooseLeft", true},
		{"choose right settings", "chooseRight", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdio, log := playlistCLI(t, tc.resolver)
			left, right, leftBlob, rightBlob := thumbnailPlaylistArchives(t, tc.settings)
			for _, reverse := range []bool{false, true} {
				a, b, aBlob, bBlob := left, right, leftBlob, rightBlob
				if reverse {
					a, b, aBlob, bBlob = b, a, bBlob, aBlob
				}
				// Without a conflict, the left side wins.
				want := aBlob
				if tc.resolver == "chooseRight" {
					want = bBlob
				}
				output := filepath.Join(t.TempDir(), "merged.jwlibrary")
				require.NoError(t, merge(a, b, output, stdio))
				playlisttest.Verify(t, output, []playlisttest.Playlist{{Name: "Settings", Items: []string{"A"}}})
				got := &model.Database{TempDir: t.TempDir()}
				require.NoError(t, got.ImportJWLBackup(output))
				defer func() { require.NoError(t, got.Close()) }()
				blob, err := got.PlaylistThumbnail(1)
				require.NoError(t, err)
				require.Equal(t, want, blob)
				accuracy := 1
				if tc.settings && (tc.resolver == "chooseRight") != reverse {
					accuracy = 2
				}
				require.Equal(t, accuracy, got.PlaylistItem[1].Accuracy)
			}
			bytes, err := os.ReadFile(log)
			require.NoError(t, err)
			require.NotContains(t, string(bytes), "Select which side")
		})
	}
}

func TestPlaylistCLIShowsOnlyChangedSettings(t *testing.T) {
	left := &model.PlaylistItem{Label: "Recognizable item", Accuracy: 1, EndAction: 3}
	right := model.MakeModelCopy(left).(*model.PlaylistItem)
	right.Accuracy = 2
	a, b := playlistConflictText(merger.MergeConflict{Left: left, Right: right}, nil)
	require.Equal(t, "Accuracy: 1", a)
	require.Equal(t, "Accuracy: 2", b)
	a, b = playlistConflictText(merger.MergeConflict{Left: &model.PlaylistItemIndependentMediaMap{DurationTicks: 10}, Right: &model.PlaylistItemIndependentMediaMap{DurationTicks: 20}}, nil)
	require.Equal(t, "DurationTicks: 10", a)
	require.Equal(t, "DurationTicks: 20", b)
}

func TestPlaylistCLIRequiredDecisionEOFAndInvalidResolverDoNotExport(t *testing.T) {
	for _, tc := range []struct{ name, resolver, wantError string }{
		{"missing decision at EOF", "", "conflict selection canceled or unavailable"},
		{"unsupported newest choice", "chooseNewest", "playlist resolver must be chooseLeft or chooseRight"},
		{"unknown choice", "unknown", "playlist resolver must be chooseLeft or chooseRight"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdio, _ := playlistCLI(t, tc.resolver)
			left, right, _, _ := thumbnailPlaylistArchives(t, true)
			output := filepath.Join(t.TempDir(), "unresolved.jwlibrary")
			err := merge(left, right, output, stdio)
			require.ErrorContains(t, err, tc.wantError)
			_, err = os.Stat(output)
			require.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}
