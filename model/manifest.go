package model

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

const (
	version                   = 1
	supportedSchemaVersionMin = 13
	supportedSchemaVersionMax = 14
)

type manifest struct {
	CreationDate   string         `json:"creationDate"`
	UserDataBackup userDataBackup `json:"userDataBackup"`
	Name           string         `json:"name"`
	Type           int            `json:"type"`
	Version        int            `json:"version"`
}
type userDataBackup struct {
	LastModifiedDate string `json:"lastModifiedDate"`
	Hash             string `json:"hash"`
	DatabaseName     string `json:"databaseName"`
	SchemaVersion    int    `json:"schemaVersion"`
	DeviceName       string `json:"deviceName"`
}

// importManifest imports a manifest.json at path
func (mfst *manifest) importManifest(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	blob, _ := io.ReadAll(file)

	err = json.Unmarshal([]byte(blob), &mfst)
	if err != nil {
		return errors.Wrap(err, "Could not unmarshall backup manifest file")
	}

	return nil
}

// validateManifest checks if the backup file is compatible by validating the manifest
func (mfst *manifest) validateManifest() error {
	if mfst.Version > version {
		return fmt.Errorf("manifest version is too new. Should be %d is %d. "+
			"Make sure you use the latest version of the merger", version, mfst.Version)
	}
	if mfst.Version < version {
		return fmt.Errorf("manifest version is too old. Should be %d is %d. "+
			"You might need to upgrade to a newer version of JW Library first", version, mfst.Version)
	}

	if mfst.UserDataBackup.SchemaVersion > supportedSchemaVersionMax {
		return fmt.Errorf("schema version is too new. Should be up to %d is %d. "+
			"Make sure you use the latest version of the merger", supportedSchemaVersionMax, mfst.UserDataBackup.SchemaVersion)
	}
	if mfst.UserDataBackup.SchemaVersion < supportedSchemaVersionMin {
		return fmt.Errorf("schema version is too old. Should be at least %d is %d. "+
			"You might need to upgrade to a newer version of JW Library first", supportedSchemaVersionMin, mfst.UserDataBackup.SchemaVersion)
	}

	return nil
}

// generateManifest generates a manifest from the given information, which can
// later be exported
func generateManifest(backupName string, dbFile string) (*manifest, error) {
	// Get SHA256 of SQLite file
	f, err := os.Open(dbFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while opening SQLite file %s to calculate hash", dbFile)
	}
	defer f.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		log.Fatal(err)
	}
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	mfst := &manifest{
		CreationDate: time.Now().Format("2006-01-02"),
		UserDataBackup: userDataBackup{
			LastModifiedDate: time.Now().Format("2006-01-02T15:04:05-07:00"),
			Hash:             hash,
			DatabaseName:     filepath.Base(dbFile),
			SchemaVersion:    supportedSchemaVersionMax,
			DeviceName:       "library-merger",
		},
		Name:    backupName,
		Type:    0,
		Version: version,
	}

	return mfst, nil
}

// exportManifest exports a manifest at path
func (mfst *manifest) exportManifest(path string) error {
	bytes, err := json.Marshal(mfst)
	if err != nil {
		return errors.Wrap(err, "Error while marshalling manifest")
	}

	if err := os.WriteFile(path, bytes, 0o644); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error while saving manifest file at %v", path))
	}

	return nil
}
