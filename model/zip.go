package model

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zip"
)

// Build beside the destination so a failed export cannot replace a good backup.
func zipFiles(filename string, files []string) error {
	output, err := os.CreateTemp(filepath.Dir(filename), ".library-merger-*.zip")
	if err != nil {
		return err
	}
	temporary := output.Name()
	defer func() { _ = os.Remove(temporary) }()
	err = writeZip(output, files)
	if err == nil {
		err = output.Sync()
	}
	err = errors.Join(err, output.Close())
	if err != nil {
		return err
	}
	return os.Rename(temporary, filename)
}

// Closing the writer emits the central directory and can be the first write
// failure, even when every entry's contents were successfully copied.
func writeZip(output io.Writer, files []string) error {
	writer := zip.NewWriter(output)
	for _, file := range files {
		if err := addFileToZip(writer, file); err != nil {
			return errors.Join(err, writer.Close())
		}
	}
	return writer.Close()
}

func addFileToZip(writer *zip.Writer, filename string) (err error) {
	input, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, input.Close()) }()
	info, err := input.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate
	entry, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(entry, input)
	return err
}
