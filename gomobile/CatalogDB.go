package gomobile

import (
	"context"

	"github.com/AndreasSko/go-library-merger/publication"
)

// DownloadManager keeps all the information of a running download, enabling it
// to check progress and also cancel the download if necessary
type DownloadManager struct {
	Progress  *DownloadProgress
	prgrsChan chan publication.Progress
	ctx       context.Context
	cancel    context.CancelFunc
	err       error
}

// DownloadProgress represents the progress of a running download
type DownloadProgress struct {
	Size           int64
	BytesComplete  int64
	BytesPerSecond float64
	Progress       float64
	Done           bool
	Canceled       bool
}

// DownloadCatalog downloads the newest catalog.db and saves it at dst. The
// returned DownloadManager allows to keep track and manage the running download
func DownloadCatalog(dst string) *DownloadManager {
	ctx, cancel := context.WithCancel(context.Background())
	dm := &DownloadManager{
		Progress:  &DownloadProgress{},
		prgrsChan: make(chan publication.Progress),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Start download in sub-goroutine, while monitoring its progress
	go func() {
		done := make(chan struct{})
		go func() {
			err := publication.DownloadCatalog(dm.ctx, dm.prgrsChan, dst)
			if err != nil {
				dm.err = err
			}
			done <- struct{}{}
		}()
		for progress := range dm.prgrsChan {
			dm.Progress.Size = progress.Size
			dm.Progress.BytesComplete = progress.BytesComplete
			dm.Progress.BytesPerSecond = progress.BytesPerSecond
			dm.Progress.Progress = progress.Progress
		}
		<-done
		dm.Progress.Done = true
	}()

	return dm
}

// CancelDownload cancels a running download
func (dm *DownloadManager) CancelDownload() {
	dm.cancel()
	dm.Progress.Canceled = true
}

// DownloadSuccessful indicates if the download has been successful
func (dm *DownloadManager) DownloadSuccessful() bool {
	return dm.Progress.Done && dm.err == nil && !dm.Progress.Canceled
}

// Error returns possible errors of a download as a string
func (dm *DownloadManager) Error() string {
	if dm.err != nil {
		return dm.err.Error()
	}
	return ""
}

// CatalogNeedsUpdate checks if catalog.db located at path is still up-to-date.
// For now it just makes sure that it is younger than one month.
// If it can't find a file at path, it returns true
func CatalogNeedsUpdate(path string) bool {
	return publication.CatalogNeedsUpdate(path)
}

// CatalogExists checks if catalog.db exists at path
func CatalogExists(path string) bool {
	return publication.CatalogExists(path)
}

// CatalogSize returns the size of the catalog.db at path
func CatalogSize(path string) int64 {
	return publication.CatalogSize(path)
}
