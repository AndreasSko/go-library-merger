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

func cosmeticPlaylistArchives(t *testing.T, settings bool) (string, string, []byte) {
	t.Helper()
	left := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Cosmetic", Items: []string{"A"}}})
	db := &model.Database{TempDir: t.TempDir()}
	require.NoError(t, db.ImportJWLBackup(left))
	t.Cleanup(func() { require.NoError(t, db.CleanupMedia()) })
	media := db.IndependentMedia[1]
	original, err := os.ReadFile(filepath.Join(db.IndependentMediaDir, media.FilePath))
	require.NoError(t, err)
	changed := []byte("cosmetic CLI thumbnail")
	originalHash := media.UniqueKey()
	media.Hash = fmt.Sprintf("%x", sha256.Sum256(changed))
	require.NoError(t, os.WriteFile(filepath.Join(db.IndependentMediaDir, media.FilePath), changed, 0600))
	if settings {
		db.PlaylistItem[1].Accuracy = 2
	}
	right := filepath.Join(t.TempDir(), "right.jwlibrary")
	require.NoError(t, db.ExportJWLBackup(right))
	want := original
	if media.UniqueKey() < originalHash {
		want = changed
	}
	return left, right, want
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

func TestPlaylistCLINoninteractiveCosmeticAndExplicitSettingsResolvers(t *testing.T) {
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
			left, right, want := cosmeticPlaylistArchives(t, tc.settings)
			for _, reverse := range []bool{false, true} {
				a, b := left, right
				if reverse {
					a, b = b, a
				}
				output := filepath.Join(t.TempDir(), "merged.jwlibrary")
				require.NoError(t, merge(a, b, output, stdio))
				playlisttest.Verify(t, output, []playlisttest.Playlist{{Name: "Cosmetic", Items: []string{"A"}}})
				got := &model.Database{TempDir: t.TempDir()}
				require.NoError(t, got.ImportJWLBackup(output))
				defer func() { require.NoError(t, got.CleanupMedia()) }()
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
			left, right, _ := cosmeticPlaylistArchives(t, true)
			output := filepath.Join(t.TempDir(), "unresolved.jwlibrary")
			err := merge(left, right, output, stdio)
			require.ErrorContains(t, err, tc.wantError)
			_, err = os.Stat(output)
			require.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}
