package gomobile

import (
	"context"
	"sync"

	"github.com/AndreasSko/go-library-merger/publication"
)

// DownloadManager keeps all the information of a running download, enabling it
// to check progress and also cancel the download if necessary
type DownloadManager struct {
	mu        sync.RWMutex
	progress  DownloadProgress
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
		prgrsChan: make(chan publication.Progress),
		ctx:       ctx,
		cancel:    cancel,
	}

	// The worker owns publication state; consumers receive immutable snapshots.
	go func() {
		result := make(chan error, 1)
		go func() { result <- publication.DownloadCatalog(dm.ctx, dm.prgrsChan, dst) }()
		for progress := range dm.prgrsChan {
			dm.mu.Lock()
			dm.progress.Size = progress.Size
			dm.progress.BytesComplete = progress.BytesComplete
			dm.progress.BytesPerSecond = progress.BytesPerSecond
			dm.progress.Progress = progress.Progress
			dm.mu.Unlock()
		}
		err := <-result
		dm.mu.Lock()
		dm.err = err
		dm.progress.Done = true
		dm.mu.Unlock()
	}()

	return dm
}

// GetProgress returns a snapshot that stays unchanged as the download advances.
func (dm *DownloadManager) GetProgress() *DownloadProgress {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	snapshot := dm.progress
	return &snapshot
}

// CancelDownload cancels a running download
func (dm *DownloadManager) CancelDownload() {
	dm.cancel()
	dm.mu.Lock()
	dm.progress.Canceled = true
	dm.mu.Unlock()
}

// DownloadSuccessful indicates if the download has been successful
func (dm *DownloadManager) DownloadSuccessful() bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.progress.Done && dm.err == nil && !dm.progress.Canceled
}

// Error returns possible errors of a download as a string
func (dm *DownloadManager) Error() string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
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
