package model_test

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/AndreasSko/go-library-merger/merger"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

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

func TestPlaylistLongRepetitiveInsertion(t *testing.T) {
	parent := strings.Repeat("AB", 500)
	assertAdditivePlaylist(t, parent, parent+"A")
}
