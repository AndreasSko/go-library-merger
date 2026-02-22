package model_test

import (
	"archive/zip"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

// rewritePlaylistArchive changes synthetic input and keeps the manifest valid,
// so rejection must come from the relationship/payload contract under test.
func rewritePlaylistArchive(t *testing.T, input, query string, mutate func(map[string][]byte)) string {
	t.Helper()
	z, err := zip.OpenReader(input)
	require.NoError(t, err)
	blobs := map[string][]byte{}
	for _, f := range z.File {
		r, e := f.Open()
		require.NoError(t, e)
		blobs[f.Name], e = io.ReadAll(r)
		require.NoError(t, e)
		require.NoError(t, r.Close())
	}
	require.NoError(t, z.Close())
	dbpath := filepath.Join(t.TempDir(), "userData.db")
	require.NoError(t, os.WriteFile(dbpath, blobs["userData.db"], 0600))
	c, e := sql.Open("sqlite3", dbpath)
	require.NoError(t, e)
	if query != "" {
		_, e = c.Exec(query)
		require.NoError(t, e)
	}
	require.NoError(t, c.Close())
	blobs["userData.db"], e = os.ReadFile(dbpath)
	require.NoError(t, e)
	var manifest map[string]interface{}
	require.NoError(t, json.Unmarshal(blobs["manifest.json"], &manifest))
	manifest["userDataBackup"].(map[string]interface{})["hash"] = fmt.Sprintf("%x", sha256.Sum256(blobs["userData.db"]))
	blobs["manifest.json"], e = json.Marshal(manifest)
	require.NoError(t, e)
	if mutate != nil {
		mutate(blobs)
	}
	out := filepath.Join(t.TempDir(), "negative.jwlibrary")
	f, e := os.Create(out)
	require.NoError(t, e)
	w := zip.NewWriter(f)
	for name, data := range blobs {
		entry, e := w.Create(name)
		require.NoError(t, e)
		_, e = entry.Write(data)
		require.NoError(t, e)
	}
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())
	return out
}

func TestPlaylistUnsupportedTablesRejectIndependently(t *testing.T) {
	cases := []struct{ table, query string }{
		{"PlaylistItemMarker", "INSERT INTO PlaylistItemMarker VALUES(1,1,'clip',0,100,0)"},
		// Deliberately orphan these maps so the parent-table rejection cannot mask
		// a missing guard for either child table.
		{"PlaylistItemMarkerBibleVerseMap", "INSERT INTO PlaylistItemMarkerBibleVerseMap VALUES(1,1)"},
		{"PlaylistItemMarkerParagraphMap", "INSERT INTO PlaylistItemMarkerParagraphMap VALUES(1,1,1,1)"},
	}
	for _, tc := range cases {
		t.Run(tc.table, func(t *testing.T) {
			p := rewritePlaylistArchive(t, playlisttest.Build(t, []playlisttest.Playlist{{Name: "Portable", Items: []string{"A"}}}), tc.query, nil)
			db := &model.Database{TempDir: t.TempDir()}
			require.ErrorContains(t, db.ImportJWLBackup(p), "not supported yet for table "+tc.table)
			omitted := &model.Database{TempDir: t.TempDir(), SkipPlaylists: true}
			require.NoError(t, omitted.ImportJWLBackup(p))
			require.True(t, omitted.ContainsPlaylists)
			require.Empty(t, omitted.PlaylistItem)
			require.Empty(t, omitted.IndependentMedia)
			exported := filepath.Join(t.TempDir(), "omitted.jwlibrary")
			require.NoError(t, omitted.ExportJWLBackup(exported))
			reread := &model.Database{TempDir: t.TempDir()}
			require.NoError(t, reread.ImportJWLBackup(exported))
			require.False(t, reread.ContainsPlaylists)
		})
	}
}

func TestPlaylistMalformedRelationshipsReject(t *testing.T) {
	cases := []struct{ name, query, want string }{
		{"missing item", "UPDATE TagMap SET PlaylistItemId=999 WHERE PlaylistItemId=1", "invalid PlaylistItemID"},
		{"wrong tag type", "UPDATE Tag SET Type=0", "invalid playlist TagID"},
		{"duplicate membership", "INSERT INTO Tag VALUES(2,2,'Other'); INSERT INTO TagMap VALUES(3,1,NULL,NULL,2,0)", "belongs to multiple playlist entries"},
		{"missing location", "UPDATE PlaylistItemLocationMap SET LocationId=999", "invalid playlist LocationID"},
		{"no source", "DELETE FROM PlaylistItemLocationMap", "must have exactly one media or location"},
		{"multiple locations", "INSERT INTO Location(LocationId,KeySymbol,Type) VALUES(2,'second',1); INSERT INTO PlaylistItemLocationMap VALUES(1,2,2,100)", "multiple locations"},
		{"missing media", "DELETE FROM PlaylistItemLocationMap; INSERT INTO PlaylistItemIndependentMediaMap VALUES(1,999,100)", "invalid IndependentMediaID"},
		{"both source kinds", "INSERT INTO PlaylistItemIndependentMediaMap VALUES(1,1,100)", "must have exactly one media or location"},
		{"multiple media", "DELETE FROM PlaylistItemLocationMap; INSERT INTO PlaylistItemIndependentMediaMap VALUES(1,1,100),(1,2,100)", "multiple media"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := rewritePlaylistArchive(t, playlisttest.Build(t, []playlisttest.Playlist{{Name: "Portable", Items: []string{"A", "B"}}}), tc.query, nil)
			db := &model.Database{TempDir: t.TempDir()}
			require.ErrorContains(t, db.ImportJWLBackup(p), tc.want)
		})
	}
}

func TestPlaylistInvalidPayloadsReject(t *testing.T) {
	for _, kind := range []string{"missing", "corrupt"} {
		t.Run(kind, func(t *testing.T) {
			p := rewritePlaylistArchive(t, playlisttest.Build(t, []playlisttest.Playlist{{Name: "Portable", Items: []string{"photo"}}}), "", func(blobs map[string][]byte) {
				if kind == "missing" {
					delete(blobs, "photo.bin")
				} else {
					blobs["photo.bin"] = []byte("corrupted payload")
				}
			})
			db := &model.Database{TempDir: t.TempDir()}
			require.Error(t, db.ImportJWLBackup(p))
		})
	}
	for _, path := range []string{"../outside.bin", "/outside.bin", "nested/file.bin", ".", ".."} {
		t.Run(fmt.Sprintf("path=%q", path), func(t *testing.T) {
			p := rewritePlaylistArchive(t, playlisttest.Build(t, []playlisttest.Playlist{{Name: "Portable", Items: []string{"photo"}}}), "UPDATE IndependentMedia SET FilePath='"+path+"' WHERE FilePath='photo.bin'", nil)
			db := &model.Database{TempDir: t.TempDir()}
			require.ErrorContains(t, db.ImportJWLBackup(p), "invalid independent media path")
		})
	}
}

func TestPlaylistUnsafeArchiveEntriesRejectBeforeExtraction(t *testing.T) {
	for _, entry := range []string{"../outside.bin", "/outside.bin", "nested/file.bin", "nested\\file.bin", "userData.db", "USERDATA.DB", "link.bin"} {
		t.Run(entry, func(t *testing.T) {
			input := playlisttest.Build(t, []playlisttest.Playlist{{Name: "Portable", Items: []string{"A"}}})
			source, err := zip.OpenReader(input)
			require.NoError(t, err)
			defer func() { require.NoError(t, source.Close()) }()
			path := filepath.Join(t.TempDir(), "unsafe.jwlibrary")
			f, err := os.Create(path)
			require.NoError(t, err)
			writer := zip.NewWriter(f)
			for _, original := range source.File {
				header := original.FileHeader
				output, e := writer.CreateHeader(&header)
				require.NoError(t, e)
				reader, e := original.Open()
				require.NoError(t, e)
				_, e = io.Copy(output, reader)
				require.NoError(t, e)
				require.NoError(t, reader.Close())
			}
			header := zip.FileHeader{Name: entry, Method: zip.Deflate}
			if entry == "link.bin" {
				header.SetMode(os.ModeSymlink | 0777)
			}
			output, err := writer.CreateHeader(&header)
			require.NoError(t, err)
			_, err = output.Write([]byte("outside"))
			require.NoError(t, err)
			require.NoError(t, writer.Close())
			require.NoError(t, f.Close())
			db := &model.Database{TempDir: t.TempDir()}
			require.Error(t, db.ImportJWLBackup(path))
			require.Empty(t, db.IndependentMediaDir)
		})
	}
}
