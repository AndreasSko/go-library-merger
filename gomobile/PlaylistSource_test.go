package gomobile

import (
	"database/sql"
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

func TestPlaylistConflictSourceURLUsesTheOriginalSideMapping(t *testing.T) {
	location := &model.Location{LocationType: 3,
		KeySymbol:    sql.NullString{String: "jwb-142", Valid: true},
		Track:        sql.NullInt32{Int32: 1, Valid: true},
		MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}}
	related := model.Related{Location: location}
	item := &model.PlaylistItem{PlaylistItemID: 19}
	source := &model.Database{PlaylistItemLocationMap: []*model.PlaylistItemLocationMap{
		nil, {PlaylistItemID: 1, MajorMultimediaType: 1}, {PlaylistItemID: 19, MajorMultimediaType: 2},
	}}
	want := "https://www.jw.org/finder?lank=pub-jwb-142_1_VIDEO&wtlocale=E"
	ambiguous := model.MakeDatabaseCopy(source)
	ambiguous.PlaylistItemLocationMap = append(ambiguous.PlaylistItemLocationMap, &model.PlaylistItemLocationMap{PlaylistItemID: 19, MajorMultimediaType: 2})
	for _, tc := range []struct {
		name    string
		item    model.Model
		related model.Related
		source  *model.Database
		want    string
	}{
		{"original side mapping", item, related, source, want},
		{"missing source", item, related, nil, ""},
		{"audio mapping", &model.PlaylistItem{PlaylistItemID: 1}, related, source, ""},
		{"missing mapping", &model.PlaylistItem{PlaylistItemID: 99}, related, source, ""},
		{"ambiguous mapping", item, related, ambiguous, ""},
		{"location conflict carries its own mapping", &model.PlaylistItemLocationMap{PlaylistItemID: 500, MajorMultimediaType: 2}, related, nil, want},
		{"audio location conflict", &model.PlaylistItemLocationMap{MajorMultimediaType: 1}, related, nil, ""},
		{"unrelated model", &model.Note{}, related, source, ""},
		{"missing location", item, model.Related{}, source, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, playlistConflictSourceURL(tc.item, tc.related, tc.source))
		})
	}
}
