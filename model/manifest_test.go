package model

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var exampleManifest = &manifest{
	CreationDate: time.Now().Format("2006-01-02"),
	UserDataBackup: userDataBackup{
		LastModifiedDate: time.Now().Format("2006-01-02T15:04:05-07:00"),
		Hash:             "e2e09ceba668bb1ad093b2db317237451a01ae9ff435b38c840b70dc434f184f",
		DatabaseName:     userDataFilename,
		SchemaVersion:    14,
		DeviceName:       "library-merger",
	},
	Name:    "test",
	Type:    0,
	Version: 1,
}

func Test_manifest_importManifest(t *testing.T) {
	path := filepath.Join("testdata", "manifest_correct.json")

	mfst := &manifest{}
	assert.NoError(t, mfst.importManifest(path))

	expectedMfst := &manifest{
		CreationDate: "2020-04-11",
		UserDataBackup: userDataBackup{
			LastModifiedDate: "2020-04-09T05:47:26+02:00",
			Hash:             "d87a67028133cc4de5536affe1b072841def95899b7f7450a5622112b4b5e63f",
			DatabaseName:     userDataFilename,
			SchemaVersion:    14,
			DeviceName:       "iPhone",
		},
		Name:    "UserDataBackup_2020-04-11_iPhone",
		Type:    0,
		Version: 1,
	}
	assert.Equal(t, expectedMfst, mfst)

	assert.Error(t, mfst.importManifest("nonexistentpath"))
}

func Test_validateManifest1(t *testing.T) {
	path := filepath.Join("testdata", "manifest_correct.json")

	mfst := manifest{}
	assert.NoError(t, mfst.importManifest(path))
	assert.NoError(t, mfst.validateManifest())

	path = filepath.Join("testdata", "manifest_outdated.json")
	mfst = manifest{}
	assert.NoError(t, mfst.importManifest(path))
	assert.Error(t, mfst.validateManifest())
}

func Test_manifest_validateManifest2(t *testing.T) {
	tests := []struct {
		name    string
		mfst    *manifest
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Newest schema version",
			mfst: &manifest{
				UserDataBackup: userDataBackup{
					SchemaVersion: 14,
				},
				Version: 1,
			},
			wantErr: assert.NoError,
		},
		{
			name: "Older but supported schema version",
			mfst: &manifest{
				UserDataBackup: userDataBackup{
					SchemaVersion: 13,
				},
				Version: 1,
			},
			wantErr: assert.NoError,
		},
		{
			name: "Schema version too old",
			mfst: &manifest{
				UserDataBackup: userDataBackup{
					SchemaVersion: 12,
				},
				Version: 1,
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(tt, err, "schema version is too old. Should be at least 13 is 12")
			},
		},
		{
			name: "Schema version too new",
			mfst: &manifest{
				UserDataBackup: userDataBackup{
					SchemaVersion: 15,
				},
				Version: 1,
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(tt, err, "schema version is too new. Should be up to 14 is 15")
			},
		},
		{
			name: "Manifest version too old",
			mfst: &manifest{
				UserDataBackup: userDataBackup{
					SchemaVersion: 14,
				},
				Version: 0,
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(tt, err, "manifest version is too old. Should be 1 is 0")
			},
		},
		{
			name: "Manifest version too new",
			mfst: &manifest{
				UserDataBackup: userDataBackup{
					SchemaVersion: 14,
				},
				Version: 2,
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(tt, err, "manifest version is too new. Should be 1 is 2")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mfst.validateManifest()
			tt.wantErr(t, err)
		})
	}
}

func Test_generateManifest(t *testing.T) {
	dbPath := filepath.Join("testdata", userDataFilename)

	mfst, err := generateManifest("test", dbPath)
	exampleManifest.UserDataBackup.LastModifiedDate = time.Now().Format("2006-01-02T15:04:05-07:00") // Could have changed in the last second..
	assert.NoError(t, err)
	assert.Equal(t, exampleManifest, mfst)

	_, err = generateManifest("test", "nonexistent.db")
	assert.Error(t, err)
}

func Test_exportManifest(t *testing.T) {
	tmp := t.TempDir()

	path := filepath.Join(tmp, "test_manifest.json")
	fmt.Println(path)
	err := exampleManifest.exportManifest(path)
	assert.NoError(t, err)
	assert.FileExists(t, path)

	otherMfst := &manifest{}
	err = otherMfst.importManifest(path)
	assert.NoError(t, err)
	assert.Equal(t, exampleManifest, otherMfst)
}
