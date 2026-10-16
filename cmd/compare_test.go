//go:build !windows
// +build !windows

package cmd

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2/terminal"
	expect "github.com/Netflix/go-expect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_compare(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	emptyFilename := filepath.Join(tmp, "empty.jwlibrary")
	leftFilename := filepath.Join(tmp, "left.jwlibrary")
	rightFilename := filepath.Join(tmp, "right.jwlibrary")
	assert.NoError(t, emptyDB.ExportJWLBackup(emptyFilename))
	assert.NoError(t, leftDB.ExportJWLBackup(leftFilename))
	assert.NoError(t, rightDB.ExportJWLBackup(rightFilename))

	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			_, err := c.ExpectString("Backups are NOT equal")
			assert.NoError(t, err)
			_, err = c.ExpectEOF()
			assert.NoError(t, err)
		},
		func(t *testing.T, c *expect.Console) {
			require.NoError(t, compare(leftFilename, emptyFilename, terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()}))
			time.Sleep(time.Millisecond * 150) // So it does not finish before go-expect finished
		})

	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			_, err := c.ExpectString("✅ Backups are equal")
			assert.NoError(t, err)
			_, err = c.ExpectEOF()
			assert.NoError(t, err)
		},
		func(t *testing.T, c *expect.Console) {
			require.NoError(t, compare(leftFilename, leftFilename, terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()}))
			time.Sleep(time.Millisecond * 150) // So it does not finish before go-expect finished
		})

	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			_, err := c.ExpectString("Backups are equal")
			assert.NoError(t, err)
			_, err = c.ExpectEOF()
			assert.NoError(t, err)
		},
		func(t *testing.T, c *expect.Console) {
			require.NoError(t, compare(rightFilename, rightFilename, terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()}))
			time.Sleep(time.Millisecond * 150) // So it does not finish before go-expect finished
		})
}
