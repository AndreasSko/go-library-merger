package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
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
