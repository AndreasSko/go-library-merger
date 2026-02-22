//go:build !windows
// +build !windows

package cmd

import (
	"path/filepath"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/AndreasSko/go-library-merger/model"
	expect "github.com/Netflix/go-expect"
	"github.com/stretchr/testify/assert"
)

func Test_purge(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	inputFilename := filepath.Join(tmp, "left.jwlibrary")
	assert.NoError(t, leftDB.ExportJWLBackup(inputFilename))

	RunCmdTest(t,
		func(t *testing.T, c *expect.Console) {
			_, err := c.ExpectString("🎉 Done")
			assert.NoError(t, err)
			_, err = c.ExpectEOF()
			assert.NoError(t, err)
		},
		func(t *testing.T, c *expect.Console) {
			Tables = "Note,Tag,TagMap"

			outputFilename := filepath.Join(tmp, "1.jwlibrary")
			purge(inputFilename, outputFilename, terminal.Stdio{In: c.Tty(), Out: c.Tty(), Err: c.Tty()})

			want := model.MakeDatabaseCopy(leftDB)
			want.Note = []*model.Note{nil}
			want.Tag = []*model.Tag{nil}
			want.TagMap = []*model.TagMap{nil}

			output := &model.Database{}
			assert.NoError(t, output.ImportJWLBackup(outputFilename))
			assert.True(t, want.Equals(output))
		})
}
