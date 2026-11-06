package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_zipFiles(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	require.NoError(t, os.WriteFile(input, []byte("content"), 0o644))

	output := filepath.Join(dir, "backup.jwlibrary")
	require.NoError(t, zipFiles(output, []string{input}))
	assert.FileExists(t, output)

	// A failed export must not leave a broken backup behind
	failed := filepath.Join(dir, "failed.jwlibrary")
	err := zipFiles(failed, []string{input, filepath.Join(dir, "missing.txt")})
	assert.Error(t, err)
	assert.NoFileExists(t, failed)
}
