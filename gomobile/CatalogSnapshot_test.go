package gomobile

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloadProgressSnapshotsAreIndependent(t *testing.T) {
	dm := &DownloadManager{cancel: func() {}}
	snapshot := dm.GetProgress()
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 1000; n++ {
				dm.mu.Lock()
				dm.progress.BytesComplete++
				dm.mu.Unlock()
				own := dm.GetProgress()
				own.Done = true
				own.BytesComplete = -1
				_ = dm.DownloadSuccessful()
				_ = dm.Error()
			}
		}()
	}
	wg.Wait()
	require.Equal(t, int64(0), snapshot.BytesComplete)
	require.Equal(t, int64(8000), dm.GetProgress().BytesComplete)
	require.False(t, dm.GetProgress().Done)
	dm.CancelDownload()
	require.True(t, dm.GetProgress().Canceled)
	require.False(t, snapshot.Canceled)
}
