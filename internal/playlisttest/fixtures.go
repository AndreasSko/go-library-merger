// Package playlisttest builds small, portable playlist archives for regression tests.
package playlisttest

import (
	"archive/zip"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

type Playlist struct {
	Name  string
	Items []string
}
type Scenario struct {
	Name              string
	Left, Right, Want []Playlist
}

func Cases() []Scenario {
	return []Scenario{
		{"front", []Playlist{{"Common", []string{"A", "B", "C"}}}, []Playlist{{"Common", []string{"X", "A", "B", "C"}}}, []Playlist{{"Common", []string{"X", "A", "B", "C"}}}},
		{"divergent", []Playlist{{"Common", []string{"photo", "A", "B", "C"}}}, []Playlist{{"Common", []string{"A", "B", "video", "C"}}}, []Playlist{{"Common", []string{"photo", "A", "B", "video", "C"}}}},
		{"separate", []Playlist{{"Left", []string{"A", "B", "C"}}}, []Playlist{{"Right", []string{"X", "photo", "video"}}}, []Playlist{{"Left", []string{"A", "B", "C"}}, {"Right", []string{"X", "photo", "video"}}}},
		{"common-plus-new", []Playlist{{"Common", []string{"photo", "A", "B", "C"}}, {"Left", []string{"X"}}}, []Playlist{{"Common", []string{"A", "B", "video", "C"}}, {"Right", []string{"Y"}}}, []Playlist{{"Common", []string{"photo", "A", "B", "video", "C"}}, {"Left", []string{"X"}}, {"Right", []string{"Y"}}}},
		{"repeat", []Playlist{{"Common", []string{"A", "B", "A"}}}, []Playlist{{"Common", []string{"X", "A", "B", "A"}}}, []Playlist{{"Common", []string{"X", "A", "B", "A"}}}},
	}
}
func Action(label string) int {
	return map[string]int{"A": 0, "B": 1, "C": 2, "X": 3, "photo": 2, "video": 3}[label]
}

// Build deliberately uses sparse location IDs and values beyond signed int32.
func Build(t *testing.T, lists []Playlist) string {
	t.Helper()
	dir := t.TempDir()
	db := &model.Database{
		IndependentMediaDir: dir,
		Tag:                 []*model.Tag{nil}, TagMap: []*model.TagMap{nil},
		PlaylistItem: []*model.PlaylistItem{nil}, IndependentMedia: []*model.IndependentMedia{nil},
		Location: make([]*model.Location, 150),
	}
	mediaIDs := map[string]int{}
	addMedia := func(name string) int {
		if id, ok := mediaIDs[name]; ok {
			return id
		}
		blob := []byte("test media: " + name)
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), blob, 0600))
		id := len(db.IndependentMedia)
		mediaIDs[name] = id
		db.IndependentMedia = append(db.IndependentMedia, &model.IndependentMedia{IndependentMediaID: id, OriginalFilename: name, FilePath: name, MimeType: "application/octet-stream", Hash: fmt.Sprintf("%x", sha256.Sum256(blob))})
		return id
	}
	for _, list := range lists {
		tag := len(db.Tag)
		db.Tag = append(db.Tag, &model.Tag{TagID: tag, TagType: 2, Name: list.Name})
		for pos, label := range list.Items {
			id := len(db.PlaylistItem)
			thumb := label + ".jpg"
			addMedia(thumb)
			db.PlaylistItem = append(db.PlaylistItem, &model.PlaylistItem{PlaylistItemID: id, Label: label, Accuracy: 1, EndAction: Action(label), StartTrimOffsetTicks: sql.NullInt64{Int64: 2500000000, Valid: true}, ThumbnailFilePath: sql.NullString{String: thumb, Valid: true}})
			db.TagMap = append(db.TagMap, &model.TagMap{TagMapID: len(db.TagMap), PlaylistItemID: sql.NullInt32{Int32: int32(id), Valid: true}, TagID: tag, Position: pos})
			if label == "photo" || label == "video" {
				media := addMedia(label + ".bin")
				db.PlaylistItemIndependentMediaMap = append(db.PlaylistItemIndependentMediaMap, &model.PlaylistItemIndependentMediaMap{PlaylistItemID: id, IndependentMediaID: media, DurationTicks: 5000000000})
			} else {
				loc := 100 + int(label[0]-'A')
				db.Location[loc] = &model.Location{LocationID: loc, KeySymbol: sql.NullString{String: label, Valid: true}, LocationType: 1, Title: sql.NullString{String: label, Valid: true}}
				db.PlaylistItemLocationMap = append(db.PlaylistItemLocationMap, &model.PlaylistItemLocationMap{PlaylistItemID: id, LocationID: loc, MajorMultimediaType: 2, BaseDurationTicks: sql.NullInt64{Int64: 5000000000, Valid: true}})
			}
		}
	}
	path := filepath.Join(dir, "fixture.jwlibrary")
	require.NoError(t, db.ExportJWLBackup(path))
	return path
}

func Verify(t *testing.T, path string, want []Playlist) {
	t.Helper()
	db := &model.Database{TempDir: t.TempDir()}
	require.NoError(t, db.ImportJWLBackup(path))
	t.Cleanup(func() { require.NoError(t, db.CleanupMedia()) })
	got := []Playlist{}
	for _, tag := range db.Tag {
		if tag == nil || tag.TagType != 2 {
			continue
		}
		maps := []*model.TagMap{}
		for _, m := range db.TagMap {
			if m != nil && m.TagID == tag.TagID {
				maps = append(maps, m)
			}
		}
		sort.Slice(maps, func(i, j int) bool { return maps[i].Position < maps[j].Position })
		list := Playlist{Name: tag.Name, Items: []string{}}
		for pos, m := range maps {
			require.Equal(t, pos, m.Position)
			p := db.PlaylistItem[m.PlaylistItemID.Int32]
			list.Items = append(list.Items, p.Label)
			require.Equal(t, Action(p.Label), p.EndAction)
			require.Equal(t, int64(2500000000), p.StartTrimOffsetTicks.Int64)
		}
		got = append(got, list)
	}
	want = append([]Playlist{}, want...)
	sort.Slice(got, func(i, j int) bool { return got[i].Name < got[j].Name })
	sort.Slice(want, func(i, j int) bool { return want[i].Name < want[j].Name })
	require.Equal(t, want, got)
	for _, m := range db.PlaylistItemLocationMap {
		if m != nil {
			require.Equal(t, db.PlaylistItem[m.PlaylistItemID].Label, db.Location[m.LocationID].Title.String)
			require.Equal(t, int64(5000000000), m.BaseDurationTicks.Int64)
		}
	}
	for _, m := range db.PlaylistItemIndependentMediaMap {
		if m != nil {
			require.Equal(t, int64(5000000000), m.DurationTicks)
		}
	}
	z, err := zip.OpenReader(path)
	require.NoError(t, err)
	defer func() { require.NoError(t, z.Close()) }()
	blobs := map[string][]byte{}
	for _, f := range z.File {
		r, e := f.Open()
		require.NoError(t, e)
		blob, e := io.ReadAll(r)
		require.NoError(t, e)
		require.NoError(t, r.Close())
		blobs[f.Name] = blob
	}
	for _, m := range db.IndependentMedia {
		if m != nil {
			blob, ok := blobs[m.FilePath]
			require.True(t, ok)
			digest := sha256.Sum256(blob)
			native := ""
			for _, value := range digest {
				native += fmt.Sprintf("%x", value)
			}
			require.Contains(t, []string{fmt.Sprintf("%x", digest), native}, m.Hash)
		}
	}
	var manifest struct {
		UserDataBackup struct {
			Hash string `json:"hash"`
		} `json:"userDataBackup"`
	}
	require.NoError(t, json.Unmarshal(blobs["manifest.json"], &manifest))
	require.Equal(t, manifest.UserDataBackup.Hash, fmt.Sprintf("%x", sha256.Sum256(blobs["userData.db"])))
	file := filepath.Join(t.TempDir(), "verify.db")
	require.NoError(t, os.WriteFile(file, blobs["userData.db"], 0600))
	c, err := sql.Open("sqlite3", file)
	require.NoError(t, err)
	defer func() { require.NoError(t, c.Close()) }()
	var integrity string
	require.NoError(t, c.QueryRow("pragma integrity_check").Scan(&integrity))
	require.Equal(t, "ok", integrity)
	rows, err := c.Query("pragma foreign_key_check")
	require.NoError(t, err)
	defer func() { require.NoError(t, rows.Close()) }()
	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}
