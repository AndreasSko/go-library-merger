package publication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/codeclysm/extract/v3"
	"github.com/pkg/errors"
)

// ManifestURL is the URL to the publication manifest
var ManifestURL = "https://app.jw-cdn.org/catalogs/publications/v4/manifest.json"

// CatalogURL is the URL to the publication catalog
var CatalogURL = "https://app.jw-cdn.org/catalogs/publications/v4/%s/catalog.db.gz"

// Progress represents the progress of a running download
type Progress struct {
	Size           int64
	BytesComplete  int64
	BytesPerSecond float64
	Progress       float64
	Duration       time.Duration
	ETA            time.Time
	Done           bool
}

type catalogManifest struct {
	Version int    `json:"version"`
	Current string `json:"current"`
}

// CatalogNeedsUpdate checks if catalog.db located at path is still up-to-date.
// For now it just makes sure that it is younger than one month.
// If it can't find a file at path, it returns true
func CatalogNeedsUpdate(path string) bool {
	stat, err := os.Stat(path)
	if err == nil {
		old := time.Now().Add(-time.Hour * 24 * 30)
		if !stat.ModTime().Before(old) {
			return false
		}
	}
	return true
}

// CatalogExists checks if catalog.db exists at path
func CatalogExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CatalogSize returns the size of the catalog.db at path
func CatalogSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}

	return info.Size()
}

// DownloadCatalog downloads the newest catalog.db and saves it at dst.
// The prgrs channel informs about the progress of the download.
func DownloadCatalog(ctx context.Context, prgrs chan Progress, dst string) error {
	if prgrs != nil {
		defer close(prgrs)
	}

	// Create tmp folder and place all files there
	tmp, err := os.MkdirTemp("", "library-merger")
	if err != nil {
		return errors.Wrap(err, "Error while creating temporary directory")
	}
	defer os.RemoveAll(tmp)

	// Fetch manifest, so we can generate the catalogURL
	mfst, err := fetchManifest(ctx)
	if err != nil {
		return errors.Wrap(err, "Could not fetch catalog manifest")
	}
	url := fmt.Sprintf(CatalogURL, mfst.Current)

	client := grab.NewClient()
	req, err := grab.NewRequest(tmp, url)
	if err != nil {
		return errors.Wrapf(err, "Error while creating request for %s", url)
	}
	req = req.WithContext(ctx)
	resp := client.Do(req)

	progress := Progress{}

	// Send a progress over the prgrsChan every 250 milliseconds
	t := time.NewTicker(250 * time.Millisecond)
	defer t.Stop()
Loop:
	for {
		progress := Progress{
			Size:           resp.Size(),
			BytesComplete:  resp.BytesComplete(),
			BytesPerSecond: resp.BytesPerSecond(),
			Progress:       resp.Progress(),
			Duration:       resp.Duration(),
			ETA:            resp.ETA(),
		}
		select {
		case <-t.C:
			select {
			case prgrs <- progress:
				continue
			default:
				continue
			}
		case <-resp.Done:
			break Loop
		}
	}
	if err := resp.Err(); err != nil {
		progress.Done = true
		select {
		case prgrs <- progress:
		default:
		}
		return errors.Wrapf(err, "Error while downloading catalog from %s", url)
	}

	// Extract and save at dst
	data, err := os.ReadFile(resp.Filename)
	if err != nil {
		return errors.Wrap(err, "Error while reading catalog.db.gz")
	}
	buffer := bytes.NewBuffer(data)
	err = extract.Gz(ctx, buffer, dst, nil)
	if err != nil {
		return errors.Wrap(err, "Error while extracting catalog.db")
	}

	progress.Done = true
	select {
	case prgrs <- progress:
	default:
	}

	return nil
}

// fetchManifest fetches the latest manifest from manifestURL
func fetchManifest(ctx context.Context) (catalogManifest, error) {
	req, err := http.NewRequest(http.MethodGet, ManifestURL, nil)
	if err != nil {
		return catalogManifest{}, errors.Wrapf(err, "Error while creating new request for %s", ManifestURL)
	}
	req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return catalogManifest{}, errors.Wrapf(err, "Could not download catalog manifest from %s", ManifestURL)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return catalogManifest{}, errors.Wrap(err, "Error while reading response body for catalog manifest")
	}

	mfst := catalogManifest{}
	err = json.Unmarshal([]byte(body), &mfst)
	if err != nil {
		return catalogManifest{}, errors.Wrap(err, "Could not unmarshall catalog manifest file")
	}

	return mfst, nil
}
