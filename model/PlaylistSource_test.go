package model

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlaylistSourceURLUsesImportedVideoIdentity(t *testing.T) {
	video := func() *Location {
		return &Location{LocationType: 3,
			KeySymbol:    sql.NullString{String: "jwb-142", Valid: true},
			Track:        sql.NullInt32{Int32: 1, Valid: true},
			MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}}
	}
	for _, tc := range []struct {
		name string
		edit func(*Location)
		kind int
		want string
	}{
		{"English video", func(*Location) {}, 2, "https://www.jw.org/finder?lank=pub-jwb-142_1_VIDEO&wtlocale=E"},
		{"German video", func(v *Location) { v.MepsLanguage.Int32 = 2 }, 2, "https://www.jw.org/finder?lank=pub-jwb-142_1_VIDEO&wtlocale=X"},
		{"title is not identity", func(v *Location) { v.Title = sql.NullString{String: "An edited title", Valid: true} }, 2, "https://www.jw.org/finder?lank=pub-jwb-142_1_VIDEO&wtlocale=E"},
		{"unknown language", func(v *Location) { v.MepsLanguage.Int32 = 999 }, 2, ""},
		{"absent language", func(v *Location) { v.MepsLanguage.Valid = false }, 2, ""},
		{"absent track", func(v *Location) { v.Track.Valid = false }, 2, ""},
		{"negative track", func(v *Location) { v.Track.Int32 = -1 }, 2, ""},
		{"zero track", func(v *Location) { v.Track.Int32 = 0 }, 2, ""},
		{"absent symbol", func(v *Location) { v.KeySymbol.Valid = false }, 2, ""},
		{"injected symbol", func(v *Location) { v.KeySymbol.String = "jwb-142&docid=1" }, 2, ""},
		{"dated identity", func(v *Location) { v.IssueTagNumber = 20261000 }, 2, ""},
		{"special edition", func(v *Location) { v.Edition = sql.NullString{String: "2026", Valid: true} }, 2, ""},
		{"specialty", func(v *Location) { v.Specialty = sql.NullString{String: "special", Valid: true} }, 2, ""},
		{"audio is not video", func(*Location) {}, 1, ""},
		{"unknown multimedia type", func(*Location) {}, 99, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			location := video()
			tc.edit(location)
			require.Equal(t, tc.want, PlaylistSourceURL(location, tc.kind))
		})
	}
	require.Empty(t, PlaylistSourceURL(nil, 2))
}

func TestPlaylistSourceURLKeepsLargeDocumentIDs(t *testing.T) {
	location := &Location{DocumentID: sql.NullInt64{Int64: 4052170279, Valid: true},
		MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}}
	require.Equal(t, "https://www.jw.org/finder?docid=4052170279&wtlocale=E", PlaylistSourceURL(location, 0))
	location.DocumentID.Valid = false
	require.Empty(t, PlaylistSourceURL(location, 0))
	location.DocumentID = sql.NullInt64{Int64: -1, Valid: true}
	require.Empty(t, PlaylistSourceURL(location, 0))
}
