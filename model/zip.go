package model

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zip"
)

// https://golangcode.com/create-zip-files-in-go/
func zipFiles(filename string, files []string) (err error) {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, newZipFile.Close())
		if err != nil {
			// Don't leave a broken backup behind that could be restored later
			_ = os.Remove(filename)
		}
	}()

	// Close writes the ZIP's central directory, so its error must not be
	// ignored: without it the file is unreadable.
	zipWriter := zip.NewWriter(newZipFile)
	for _, file := range files {
		if err := addFileToZip(zipWriter, file); err != nil {
			return errors.Join(err, zipWriter.Close())
		}
	}
	return zipWriter.Close()
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
