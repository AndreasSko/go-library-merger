//go:build !windows
// +build !windows

package gomobile

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AndreasSko/go-library-merger/publication"
	"github.com/stretchr/testify/assert"
)

func TestDownloadCatalog(t *testing.T) {
	tmp := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.String(), "manifest.json") {
			rw.Write([]byte(`{"version": 1, "current": "164a1c4b-4dbd-4909-8f88-8e7a18c562f2"}`))
		} else {
			data, err := os.ReadFile(filepath.Join("../publication/testdata", "catalog.db.gz"))
			assert.NoError(t, err)
			rw.Write(data)
		}
	}))
	defer server.Close()

	publication.ManifestURL = server.URL + "/catalogs/publications/v4/manifest.json"
	publication.CatalogURL = server.URL + "/catalogs/publications/v4/%s/catalog.db.gz"

	for range []int{0, 1} {
		dm := DownloadCatalog(filepath.Join(tmp, "catalog.db"))
		for {
			assert.NoError(t, dm.err)
			if dm.Progress.Done == true {
				break
			}
		}
		assert.Equal(t, "7ebe98db8b5edd1ab901b7d6b43647fd35790b2a332c43739efdf9383d590651",
			hashFile(filepath.Join(tmp, "catalog.db")))
		assert.True(t, dm.DownloadSuccessful())
	}

	// Test error
	publication.CatalogURL = "https://notvaliddomain.com/%s"
	dm := DownloadCatalog(filepath.Join(tmp, "catalog.db"))
	time.Sleep(5 * time.Second)
	assert.Error(t, dm.err)
	assert.NotEqual(t, "", dm.Error())
	assert.True(t, dm.Progress.Done)
	assert.False(t, dm.DownloadSuccessful())
}

func TestDownloadManager_CancelDownload(t *testing.T) {
	tmp := t.TempDir()

	dm := DownloadCatalog(filepath.Join(tmp, "catalog.db"))
	dm.CancelDownload()
	time.Sleep(time.Second)
	assert.True(t, dm.Progress.Done)
	assert.NotEqual(t, "", dm.Error())
	assert.True(t, dm.Progress.Canceled)
	assert.False(t, dm.DownloadSuccessful())
	assert.False(t, publication.CatalogExists(filepath.Join(tmp, "catalog.db")))
}

func TestCatalogNeedsUpdate(t *testing.T) {
	tmp := t.TempDir()

	assert.True(t, CatalogNeedsUpdate("not-valid-path"))

	filePath := filepath.Join(tmp, "catalog.db")
	_, err := os.Create(filePath)
	assert.NoError(t, err)

	assert.False(t, CatalogNeedsUpdate(filePath))

	os.Chtimes(filePath, time.Now(), time.Now().Add(-time.Hour*24*31))
	assert.True(t, CatalogNeedsUpdate(filePath))
}

func TestCatalogExists(t *testing.T) {
	tmp := t.TempDir()

	filePath := filepath.Join(tmp, "catalog.db")
	_, err := os.Create(filePath)
	assert.NoError(t, err)

	assert.False(t, CatalogExists("not-valid-path"))
	assert.True(t, CatalogExists(filePath))
}

func TestCatalogSize(t *testing.T) {
	assert.Equal(t, int64(77824), CatalogSize(filepath.Join("../publication/testdata", "catalog.db")))
	assert.Equal(t, int64(0), CatalogSize("not-valid-path"))
}

func hashFile(path string) string {
	f, _ := os.Open(path)
	hasher := sha256.New()
	io.Copy(hasher, f)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
