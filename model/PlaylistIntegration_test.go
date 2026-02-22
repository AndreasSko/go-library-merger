package model_test

import (
	"path/filepath"
	"testing"

	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

func TestPlaylistCopyAndEquality(t *testing.T) {
	p := playlisttest.Build(t, playlisttest.Cases()[1].Left)
	a, b := &model.Database{TempDir: t.TempDir()}, &model.Database{TempDir: t.TempDir()}
	require.NoError(t, a.ImportJWLBackup(p))
	require.NoError(t, b.ImportJWLBackup(p))
	defer func() { require.NoError(t, a.CleanupMedia()) }()
	defer func() { require.NoError(t, b.CleanupMedia()) }()
	require.NotEqual(t, a.IndependentMediaDir, b.IndependentMediaDir)
	require.True(t, a.Equals(b))
	cp := model.MakeDatabaseCopy(a)
	require.True(t, a.Equals(cp))
	require.Equal(t, a.PlaylistItem[1].UniqueKey(), cp.PlaylistItem[1].UniqueKey())
	originalKey := a.PlaylistItem[2].UniqueKey()
	for _, loc := range cp.Location {
		if loc != nil && loc.Title.String == "A" {
			loc.KeySymbol.String = "migrated"
		}
	}
	require.NotEqual(t, originalKey, cp.PlaylistItem[2].UniqueKey())
	require.Equal(t, originalKey, a.PlaylistItem[2].UniqueKey())
	cp.PlaylistItem[1].Label = "changed"
	require.NotEqual(t, a.PlaylistItem[1].Label, cp.PlaylistItem[1].Label)
}

func TestSkipPlaylistsAndInvalidExport(t *testing.T) {
	p := playlisttest.Build(t, playlisttest.Cases()[0].Left)
	db := &model.Database{SkipPlaylists: true, TempDir: t.TempDir()}
	require.NoError(t, db.ImportJWLBackup(p))
	require.True(t, db.ContainsPlaylists)
	require.Empty(t, db.PlaylistItem)
	require.Empty(t, db.IndependentMedia)
	require.Empty(t, db.IndependentMediaDir)
	db = &model.Database{TempDir: t.TempDir()}
	require.NoError(t, db.ImportJWLBackup(p))
	defer func() { require.NoError(t, db.CleanupMedia()) }()
	db.PlaylistItem = nil
	require.ErrorContains(t, db.ExportJWLBackup(filepath.Join(t.TempDir(), "bad.jwlibrary")), "references missing PlaylistItem")
}

func TestPlaylistExportRejectsMissingMembershipOrContent(t *testing.T) {
	fixture := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Graph", Items: []string{"A"}}})
	for _, membership := range []bool{false, true} {
		db := &model.Database{TempDir: t.TempDir()}
		require.NoError(t, db.ImportJWLBackup(fixture))
		defer func() { require.NoError(t, db.CleanupMedia()) }()
		if membership {
			db.TagMap = nil
		} else {
			db.PlaylistItemLocationMap = nil
		}
		require.ErrorContains(t, db.ExportJWLBackup(filepath.Join(t.TempDir(), "invalid.jwlibrary")), "invalid backup playlist relationships")
	}
}
