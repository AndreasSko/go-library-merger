//go:build !windows

package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
	"github.com/AndreasSko/go-library-merger/model"
	expect "github.com/Netflix/go-expect"
	"github.com/stretchr/testify/require"
)

func TestPlaylistCLIInteractiveConflict(t *testing.T) {
	for _, right := range []bool{false, true} {
		t.Run(map[bool]string{false: "left", true: "right"}[right], func(t *testing.T) {
			leftPath, rightPath, want, rightBlob := thumbnailPlaylistArchives(t, true)
			if right {
				want = rightBlob
			}
			previous := PlaylistResolver
			PlaylistResolver = ""
			t.Cleanup(func() { PlaylistResolver = previous })
			console, err := expect.NewConsole(expect.WithDefaultTimeout(5 * time.Second))
			require.NoError(t, err)
			defer func() { require.NoError(t, console.Close()) }()
			done := make(chan struct {
				text string
				err  error
			}, 1)
			go func() {
				text, err := console.ExpectString("Select which side should be chosen:")
				if err != nil {
					done <- struct {
						text string
						err  error
					}{text, err}
					return
				}
				answer := ""
				if right {
					answer = string(terminal.KeyArrowDown)
				}
				_, err = console.SendLine(answer)
				if err == nil {
					_, err = console.ExpectString("🎉 Finished merging!")
				}
				done <- struct {
					text string
					err  error
				}{text, err}
			}()
			output := filepath.Join(t.TempDir(), "interactive.jwlibrary")
			require.NoError(t, merge(leftPath, rightPath, output, terminal.Stdio{In: console.Tty(), Out: console.Tty(), Err: console.Tty()}))
			interaction := <-done
			require.NoError(t, interaction.err)
			text := interaction.text
			require.Contains(t, text, "Playlist: Settings")
			require.Contains(t, text, "Item: A")
			require.Contains(t, text, "Source: A")
			require.Contains(t, text, "Accuracy:")
			require.NotContains(t, text, "StartTrimOffsetTicks:")
			require.NotContains(t, text, "Thumbnail")
			playlisttest.Verify(t, output, []playlisttest.Playlist{{Name: "Settings", Items: []string{"A"}}})
			got := &model.Database{TempDir: t.TempDir()}
			require.NoError(t, got.ImportJWLBackup(output))
			defer func() { require.NoError(t, got.Close()) }()
			accuracy := 1
			if right {
				accuracy = 2
			}
			require.Equal(t, accuracy, got.PlaylistItem[1].Accuracy)
			blob, err := got.PlaylistThumbnail(1)
			require.NoError(t, err)
			require.Equal(t, want, blob)
		})
	}
}

func TestPlaylistCLIInteractiveCancelDoesNotExport(t *testing.T) {
	left, right, _, _ := thumbnailPlaylistArchives(t, true)
	previous := PlaylistResolver
	PlaylistResolver = ""
	t.Cleanup(func() { PlaylistResolver = previous })
	console, err := expect.NewConsole(expect.WithDefaultTimeout(5 * time.Second))
	require.NoError(t, err)
	defer func() { require.NoError(t, console.Close()) }()
	stdio := terminal.Stdio{In: console.Tty(), Out: console.Tty(), Err: console.Tty()}
	// Disable the PTY driver's signal handling before sending Ctrl-C: the
	// prompt renders before survey itself enters raw mode.
	reader := terminal.NewRuneReader(stdio)
	require.NoError(t, reader.SetTermMode())
	defer func() { require.NoError(t, reader.RestoreTermMode()) }()
	done := make(chan error, 1)
	go func() {
		_, err := console.ExpectString("Select which side should be chosen:")
		if err == nil {
			_, err = console.Send(string(terminal.KeyInterrupt))
		}
		done <- err
	}()
	output := filepath.Join(t.TempDir(), "canceled.jwlibrary")
	err = merge(left, right, output, stdio)
	require.ErrorIs(t, err, terminal.InterruptErr)
	require.NoError(t, <-done)
	_, err = os.Stat(output)
	require.ErrorIs(t, err, os.ErrNotExist)
}
