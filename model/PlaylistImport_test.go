package model_test

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/AndreasSko/go-library-merger/internal/playlisttest"
	"github.com/AndreasSko/go-library-merger/merger"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

func TestPlaylistCopyAndEquality(t *testing.T) {
	p := playlisttest.Build(t, playlisttest.Cases()[1].Left)
	a, b := &model.Database{TempDir: t.TempDir()}, &model.Database{TempDir: t.TempDir()}
	require.NoError(t, a.ImportJWLBackup(p))
	require.NoError(t, b.ImportJWLBackup(p))
	defer func() { require.NoError(t, a.Close()) }()
	defer func() { require.NoError(t, b.Close()) }()
	require.NotEqual(t, a.WorkingDir, b.WorkingDir)
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

func TestSkipUnsupportedPlaylistsKeepsSupportedOnes(t *testing.T) {
	p := playlisttest.Build(t, playlisttest.Cases()[0].Left)
	db := &model.Database{SkipUnsupportedPlaylists: true, TempDir: t.TempDir()}
	require.NoError(t, db.ImportJWLBackup(p))
	require.False(t, db.ContainsUnsupportedPlaylists)
	require.NotEmpty(t, db.PlaylistItem)
	require.NotEmpty(t, db.IndependentMedia)
	require.NoError(t, db.Close())
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
			p := playlisttest.Rewrite(t, playlisttest.Build(t, []playlisttest.Playlist{{Name: "Portable", Items: []string{"A"}}}), tc.query, nil)
			db := &model.Database{TempDir: t.TempDir()}
			require.ErrorContains(t, db.ImportJWLBackup(p), "not supported yet for table "+tc.table)
			omitted := &model.Database{TempDir: t.TempDir(), SkipUnsupportedPlaylists: true}
			require.NoError(t, omitted.ImportJWLBackup(p))
			require.True(t, omitted.ContainsUnsupportedPlaylists)
			require.Empty(t, omitted.PlaylistItem)
			require.Empty(t, omitted.IndependentMedia)
			exported := filepath.Join(t.TempDir(), "omitted.jwlibrary")
			require.NoError(t, omitted.ExportJWLBackup(exported))
			reread := &model.Database{TempDir: t.TempDir()}
			require.NoError(t, reread.ImportJWLBackup(exported))
			require.False(t, reread.ContainsUnsupportedPlaylists)
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
			p := playlisttest.Rewrite(t, playlisttest.Build(t, []playlisttest.Playlist{{Name: "Portable", Items: []string{"A", "B"}}}), tc.query, nil)
			db := &model.Database{TempDir: t.TempDir()}
			require.ErrorContains(t, db.ImportJWLBackup(p), tc.want)
		})
	}
}

func TestPlaylistInvalidPayloadsReject(t *testing.T) {
	for _, kind := range []string{"missing", "corrupt"} {
		t.Run(kind, func(t *testing.T) {
			p := playlisttest.Rewrite(t, playlisttest.Build(t, []playlisttest.Playlist{{Name: "Portable", Items: []string{"photo"}}}), "", func(blobs map[string][]byte) {
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
			p := playlisttest.Rewrite(t, playlisttest.Build(t, []playlisttest.Playlist{{Name: "Portable", Items: []string{"photo"}}}), "UPDATE IndependentMedia SET FilePath='"+path+"' WHERE FilePath='photo.bin'", nil)
			db := &model.Database{TempDir: t.TempDir()}
			require.ErrorContains(t, db.ImportJWLBackup(p), "invalid independent media path")
		})
	}
}

func TestUnsafeArchivePathsReject(t *testing.T) {
	for _, entry := range []string{"../outside.bin", "/outside.bin", "nested/file.bin", "nested\\file.bin"} {
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
			output, err := writer.CreateHeader(&header)
			require.NoError(t, err)
			_, err = output.Write([]byte("outside"))
			require.NoError(t, err)
			require.NoError(t, writer.Close())
			require.NoError(t, f.Close())
			db := &model.Database{TempDir: t.TempDir()}
			require.Error(t, db.ImportJWLBackup(path))
			require.Empty(t, db.WorkingDir)
		})
	}
}

func additivePlaylist(t *testing.T, sequence string) *model.Database {
	t.Helper()
	db := &model.Database{Tag: []*model.Tag{nil, {TagID: 1, TagType: 2, Name: "Repeated"}}, Location: []*model.Location{nil, {LocationID: 1, LocationType: 1, KeySymbol: sql.NullString{String: "A", Valid: true}}, {LocationID: 2, LocationType: 1, KeySymbol: sql.NullString{String: "B", Valid: true}}}, PlaylistItem: []*model.PlaylistItem{nil}, TagMap: []*model.TagMap{nil}, PlaylistItemLocationMap: []*model.PlaylistItemLocationMap{nil}}
	for position, content := range sequence {
		id, location := position+1, 1
		if content == 'B' {
			location = 2
		}
		db.PlaylistItem = append(db.PlaylistItem, &model.PlaylistItem{PlaylistItemID: id, Label: string(content), Accuracy: 1})
		db.TagMap = append(db.TagMap, &model.TagMap{TagMapID: id, TagID: 1, Position: position, PlaylistItemID: sql.NullInt32{Int32: int32(id), Valid: true}})
		db.PlaylistItemLocationMap = append(db.PlaylistItemLocationMap, &model.PlaylistItemLocationMap{PlaylistItemID: id, LocationID: location, MajorMultimediaType: 2})
	}
	require.NoError(t, db.LinkPlaylistItems())
	return db
}

func assertAdditivePlaylist(t *testing.T, parent, child string) {
	t.Helper()
	for _, reverse := range []bool{false, true} {
		left, right := additivePlaylist(t, parent), additivePlaylist(t, child)
		if reverse {
			left, right = right, left
		}
		merged := &model.Database{Tag: left.Tag, Location: left.Location}
		require.NoError(t, merger.MergePlaylists(left, right, merged, nil), "reverse=%t", reverse)
		maps, _, err := merger.MergeTagMaps(left.TagMap, right.TagMap, nil)
		require.NoError(t, err)
		ordered := []*model.TagMap{}
		for _, row := range maps {
			if row != nil && row.PlaylistItemID.Valid {
				ordered = append(ordered, row)
			}
		}
		sort.Slice(ordered, func(i, j int) bool { return ordered[i].Position < ordered[j].Position })
		var got strings.Builder
		for _, row := range ordered {
			got.WriteString(merged.PlaylistItem[row.PlaylistItemID.Int32].Label)
		}
		require.Equal(t, child, got.String(), "reverse=%t", reverse)
		require.Len(t, merged.PlaylistItem, len(child)+1)
	}
}

// Every one-insertion child contains its complete parent in order. Identical
// repetitions must not invent a reorder, discard a copy, or add an extra copy.
func TestPlaylistAdditiveSequencesExhaustive(t *testing.T) {
	pairs := 0
	for length := 0; length <= 4; length++ {
		for mask := 0; mask < 1<<length; mask++ {
			parent := ""
			for position := 0; position < length; position++ {
				if mask&(1<<position) == 0 {
					parent += "A"
				} else {
					parent += "B"
				}
			}
			children := map[string]bool{}
			for position := 0; position <= length; position++ {
				for _, content := range []string{"A", "B"} {
					children[parent[:position]+content+parent[position:]] = true
				}
			}
			for child := range children {
				pairs++
				t.Run(fmt.Sprintf("%s_to_%s", parent, child), func(t *testing.T) { assertAdditivePlaylist(t, parent, child) })
			}
		}
	}
	require.Equal(t, 160, pairs)
}
