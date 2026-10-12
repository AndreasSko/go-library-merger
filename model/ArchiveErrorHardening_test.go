package model

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestReadFailuresReturnErrors(t *testing.T) {
	// A directory opens successfully, but cannot be read as file contents.
	for _, tc := range []struct {
		name string
		read func(string) error
		want string
	}{
		{"manifest contents", func(path string) error { return (&manifest{}).importManifest(path) }, "Could not read"},
		{"database hash", func(path string) error { _, err := generateManifest("fictional", path); return err }, "calculate hash"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.read(t.TempDir())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("read failure: %v", err)
			}
		})
	}
}

type failingZipWriter struct{ err error }

func (w failingZipWriter) Write(p []byte) (int, error) { return 0, w.err }

func TestZipCloseFailureIsReturned(t *testing.T) {
	failure := errors.New("central directory write failed")
	// Empty archives write their bytes only on Close.
	if err := writeZip(failingZipWriter{failure}, nil); !errors.Is(err, failure) {
		t.Fatalf("close failure: %v", err)
	}
}

func TestFailedZipPreservesDestination(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "backup.jwlibrary")
	original := []byte("previous good backup")
	if err := os.WriteFile(target, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := zipFiles(target, []string{filepath.Join(dir, "missing")}); err == nil {
		t.Fatal("expected input error")
	}
	got, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(got, original) {
		t.Fatalf("existing backup changed: %q %v", got, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary output leaked: %v %v", entries, err)
	}
}

func TestZipSuccessfulReplacement(t *testing.T) {
	dir := t.TempDir()
	target, input := filepath.Join(dir, "backup.jwlibrary"), filepath.Join(dir, "fictional.txt")
	if err := os.WriteFile(target, []byte("previous backup"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, []byte("fictional contents"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := zipFiles(target, []string{input}); err != nil {
		t.Fatal(err)
	}
	archive, err := zip.OpenReader(target)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := archive.Close(); err != nil {
			t.Errorf("closing archive: %v", err)
		}
	}()
	if len(archive.File) != 1 || archive.File[0].Name != "fictional.txt" {
		t.Fatal("replacement contents differ")
	}
	reader, err := archive.File[0].Open()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Errorf("closing reader: %v", err)
		}
	}()
	contents, err := io.ReadAll(reader)
	if err != nil || string(contents) != "fictional contents" {
		t.Fatalf("replacement contents: %q %v", contents, err)
	}
}

func TestFailedZipRenameRemovesTemporaryOutput(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "existing-directory")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	if err := zipFiles(target, nil); err == nil {
		t.Fatal("expected rename failure")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 || !entries[0].IsDir() {
		t.Fatalf("destination changed or temporary output leaked: %v %v", entries, err)
	}
}
