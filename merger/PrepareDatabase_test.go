package merger

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/assert"
)

func TestPrepareDatabasesPreMerge(t *testing.T) {
	type args struct {
		left  *model.Database
		right *model.Database
	}
	tests := []struct {
		name string
		args args
		want args
	}{
		{
			name: "Do nothing",
			args: args{
				left: &model.Database{
					Bookmark: []*model.Bookmark{
						nil,
						{
							BookmarkID:            1,
							LocationID:            2,
							PublicationLocationID: 2,
						},
					},
					InputField: []*model.InputField{
						nil,
						{
							LocationID: 2,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID:    1,
							KeySymbol:     sql.NullString{"abc", true},
							ChapterNumber: sql.NullInt32{1, true},
						},
						{
							LocationID: 2,
							KeySymbol:  sql.NullString{"def", true},
						},
						{
							LocationID:    3,
							KeySymbol:     sql.NullString{"abc", true},
							ChapterNumber: sql.NullInt32{2, true},
						},
					},
					Note: []*model.Note{
						nil,
						{
							NoteID:     1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					TagMap: []*model.TagMap{
						nil,
						{
							TagMapID:   1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID: 1,
							LocationID: 2,
						},
					},
				},
				right: &model.Database{
					Bookmark: []*model.Bookmark{
						nil,
						{
							BookmarkID:            1,
							LocationID:            2,
							PublicationLocationID: 2,
						},
					},
					InputField: []*model.InputField{
						nil,
						{
							LocationID: 2,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID:    1,
							KeySymbol:     sql.NullString{"abc", true},
							ChapterNumber: sql.NullInt32{1, true},
						},
						{
							LocationID: 2,
							KeySymbol:  sql.NullString{"def", true},
						},
						{
							LocationID:    3,
							KeySymbol:     sql.NullString{"abc", true},
							ChapterNumber: sql.NullInt32{2, true},
						},
					},
					Note: []*model.Note{
						nil,
						{
							NoteID:     1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					TagMap: []*model.TagMap{
						nil,
						{
							TagMapID:   1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID: 1,
							LocationID: 2,
						},
					},
				},
			},
			want: args{
				left: &model.Database{
					Bookmark: []*model.Bookmark{
						nil,
						{
							BookmarkID:            1,
							LocationID:            2,
							PublicationLocationID: 2,
						},
					},
					InputField: []*model.InputField{
						nil,
						{
							LocationID: 2,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID:    1,
							KeySymbol:     sql.NullString{"abc", true},
							ChapterNumber: sql.NullInt32{1, true},
						},
						{
							LocationID: 2,
							KeySymbol:  sql.NullString{"def", true},
						},
						{
							LocationID:    3,
							KeySymbol:     sql.NullString{"abc", true},
							ChapterNumber: sql.NullInt32{2, true},
						},
					},
					Note: []*model.Note{
						nil,
						{
							NoteID:     1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					TagMap: []*model.TagMap{
						nil,
						{
							TagMapID:   1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID: 1,
							LocationID: 2,
						},
					},
				},
				right: &model.Database{
					Bookmark: []*model.Bookmark{
						nil,
						{
							BookmarkID:            1,
							LocationID:            2,
							PublicationLocationID: 2,
						},
					},
					InputField: []*model.InputField{
						nil,
						{
							LocationID: 2,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID:    1,
							KeySymbol:     sql.NullString{"abc", true},
							ChapterNumber: sql.NullInt32{1, true},
						},
						{
							LocationID: 2,
							KeySymbol:  sql.NullString{"def", true},
						},
						{
							LocationID:    3,
							KeySymbol:     sql.NullString{"abc", true},
							ChapterNumber: sql.NullInt32{2, true},
						},
					},
					Note: []*model.Note{
						nil,
						{
							NoteID:     1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					TagMap: []*model.TagMap{
						nil,
						{
							TagMapID:   1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID: 1,
							LocationID: 2,
						},
					},
				},
			},
		},
		{
			name: "Only remove nil entries and update location IDs",
			args: args{
				left: &model.Database{
					Bookmark: []*model.Bookmark{
						nil,
						{
							BookmarkID:            1,
							LocationID:            2,
							PublicationLocationID: 2,
						},
					},
					InputField: []*model.InputField{
						nil,
						{
							LocationID: 2,
						},
					},
					Location: []*model.Location{
						nil,
						nil,
						{
							LocationID: 2,
							KeySymbol:  sql.NullString{"def", true},
						},
						{
							LocationID:    3,
							KeySymbol:     sql.NullString{"abc", true},
							ChapterNumber: sql.NullInt32{2, true},
						},
					},
					Note: []*model.Note{
						nil,
						{
							NoteID:     1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					TagMap: []*model.TagMap{
						nil,
						{
							TagMapID:   1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID: 1,
							LocationID: 2,
						},
					},
				},
				right: &model.Database{
					Bookmark: []*model.Bookmark{
						nil,
						{
							BookmarkID:            1,
							LocationID:            2,
							PublicationLocationID: 2,
						},
					},
					InputField: []*model.InputField{
						nil,
						{
							LocationID: 2,
						},
					},
					Location: []*model.Location{
						nil,
						nil,
						{
							LocationID: 2,
							KeySymbol:  sql.NullString{"def", true},
						},
						{
							LocationID:    3,
							KeySymbol:     sql.NullString{"abc", true},
							ChapterNumber: sql.NullInt32{2, true},
						},
					},
					Note: []*model.Note{
						nil,
						{
							NoteID:     1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					TagMap: []*model.TagMap{
						nil,
						{
							TagMapID:   1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID: 1,
							LocationID: 2,
						},
					},
				},
			},
			want: args{
				left: &model.Database{
					Bookmark: []*model.Bookmark{
						nil,
						{
							BookmarkID:            1,
							LocationID:            1,
							PublicationLocationID: 1,
						},
					},
					InputField: []*model.InputField{
						nil,
						{
							LocationID: 1,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID: 1,
							KeySymbol:  sql.NullString{"def", true},
						},
						{
							LocationID:    2,
							KeySymbol:     sql.NullString{"abc", true},
							ChapterNumber: sql.NullInt32{2, true},
						},
					},
					Note: []*model.Note{
						nil,
						{
							NoteID:     1,
							LocationID: sql.NullInt32{1, true},
						},
					},
					TagMap: []*model.TagMap{
						nil,
						{
							TagMapID:   1,
							LocationID: sql.NullInt32{1, true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID: 1,
							LocationID: 1,
						},
					},
				},
				right: &model.Database{
					Bookmark: []*model.Bookmark{
						nil,
						{
							BookmarkID:            1,
							LocationID:            1,
							PublicationLocationID: 1,
						},
					},
					InputField: []*model.InputField{
						nil,
						{
							LocationID: 1,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID: 1,
							KeySymbol:  sql.NullString{"def", true},
						},
						{
							LocationID:    2,
							KeySymbol:     sql.NullString{"abc", true},
							ChapterNumber: sql.NullInt32{2, true},
						},
					},
					Note: []*model.Note{
						nil,
						{
							NoteID:     1,
							LocationID: sql.NullInt32{1, true},
						},
					},
					TagMap: []*model.TagMap{
						nil,
						{
							TagMapID:   1,
							LocationID: sql.NullInt32{1, true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID: 1,
							LocationID: 1,
						},
					},
				},
			},
		},
		{
			name: "Migrate left",
			args: args{
				left: &model.Database{
					Bookmark: []*model.Bookmark{
						nil,
						{
							BookmarkID:            1,
							LocationID:            2,
							PublicationLocationID: 2,
						},
						{
							BookmarkID:            1,
							LocationID:            4,
							PublicationLocationID: 4,
						},
					},
					InputField: []*model.InputField{
						nil,
						{
							LocationID: 2,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID:    1,
							KeySymbol:     sql.NullString{"nwt", true},
							ChapterNumber: sql.NullInt32{1, true},
						},
						{
							LocationID: 2,
							KeySymbol:  sql.NullString{"def", true},
						},
						{
							LocationID:    3,
							KeySymbol:     sql.NullString{"nwt", true},
							ChapterNumber: sql.NullInt32{2, true},
						},
						{
							LocationID:    4,
							KeySymbol:     sql.NullString{"nwtsty", true},
							ChapterNumber: sql.NullInt32{1, true},
							Title:         sql.NullString{"Keep me", true},
						},
					},
					Note: []*model.Note{
						nil,
						{
							NoteID:     1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					TagMap: []*model.TagMap{
						nil,
						{
							TagMapID:   1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "migrateMe",
						},
					},
				},
				right: &model.Database{
					Bookmark: []*model.Bookmark{
						nil,
						{
							BookmarkID:            1,
							LocationID:            2,
							PublicationLocationID: 2,
						},
					},
					InputField: []*model.InputField{
						nil,
						{
							LocationID: 2,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID:    1,
							KeySymbol:     sql.NullString{"nwtsty", true},
							ChapterNumber: sql.NullInt32{1, true},
						},
						{
							LocationID: 2,
							KeySymbol:  sql.NullString{"def", true},
						},
						{
							LocationID:    3,
							KeySymbol:     sql.NullString{"nwtsty", true},
							ChapterNumber: sql.NullInt32{2, true},
						},
					},
					Note: []*model.Note{
						nil,
						{
							NoteID:     1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					TagMap: []*model.TagMap{
						nil,
						{
							TagMapID:   1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "migrateMe",
						},
					},
				},
			},
			want: args{
				left: &model.Database{
					Bookmark: []*model.Bookmark{
						nil,
						{
							BookmarkID:            1,
							LocationID:            2,
							PublicationLocationID: 2,
						},
						{
							BookmarkID:            1,
							LocationID:            1,
							PublicationLocationID: 1,
						},
					},
					InputField: []*model.InputField{
						nil,
						{
							LocationID: 2,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID:    1,
							KeySymbol:     sql.NullString{"nwtsty", true},
							ChapterNumber: sql.NullInt32{1, true},
							Title:         sql.NullString{"Keep me", true},
						},
						{
							LocationID: 2,
							KeySymbol:  sql.NullString{"def", true},
						},
						{
							LocationID:    3,
							KeySymbol:     sql.NullString{"nwtsty", true},
							ChapterNumber: sql.NullInt32{2, true},
						},
					},
					Note: []*model.Note{
						nil,
						{
							NoteID:     1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					TagMap: []*model.TagMap{
						nil,
						{
							TagMapID:   1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "migrateMe",
						},
					},
				},
				right: &model.Database{
					Bookmark: []*model.Bookmark{
						nil,
						{
							BookmarkID:            1,
							LocationID:            2,
							PublicationLocationID: 2,
						},
					},
					InputField: []*model.InputField{
						nil,
						{
							LocationID: 2,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID:    1,
							KeySymbol:     sql.NullString{"nwtsty", true},
							ChapterNumber: sql.NullInt32{1, true},
						},
						{
							LocationID: 2,
							KeySymbol:  sql.NullString{"def", true},
						},
						{
							LocationID:    3,
							KeySymbol:     sql.NullString{"nwtsty", true},
							ChapterNumber: sql.NullInt32{2, true},
						},
					},
					Note: []*model.Note{
						nil,
						{
							NoteID:     1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					TagMap: []*model.TagMap{
						nil,
						{
							TagMapID:   1,
							LocationID: sql.NullInt32{2, true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "migrateMe",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PrepareDatabasesPreMerge(tt.args.left, tt.args.right)
			assert.Equal(t, tt.want.left, tt.args.left)
			assert.Equal(t, tt.want.right, tt.args.right)
		})
	}
}

func TestPrepareDatabasesPostMerge(t *testing.T) {
	type args struct {
		merged *model.Database
	}
	tests := []struct {
		name    string
		args    args
		want    *model.Database
		wantErr bool
	}{
		{
			name: "No duplicates",
			args: args{
				merged: &model.Database{
					BlockRange: []*model.BlockRange{
						nil,
						{
							BlockRangeID: 1,
							UserMarkID:   1,
						},
						{
							BlockRangeID: 2,
							UserMarkID:   2,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID: 1,
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "1",
						},
						{
							UserMarkID:   2,
							LocationID:   1,
							UserMarkGUID: "2",
						},
					},
				},
			},
			want: &model.Database{
				BlockRange: []*model.BlockRange{
					nil,
					{
						BlockRangeID: 1,
						UserMarkID:   1,
					},
					{
						BlockRangeID: 2,
						UserMarkID:   2,
					},
				},
				Location: []*model.Location{
					nil,
					{
						LocationID: 1,
					},
				},
				UserMark: []*model.UserMark{
					nil,
					{
						UserMarkID:   1,
						LocationID:   1,
						UserMarkGUID: "1",
					},
					{
						UserMarkID:   2,
						LocationID:   1,
						UserMarkGUID: "2",
					},
				},
			},
		},
		{
			name: "Detect something is wrong",
			args: args{
				merged: &model.Database{
					BlockRange: []*model.BlockRange{
						nil,
						{
							BlockRangeID: 1,
							UserMarkID:   1,
						},
						{
							BlockRangeID: 2,
							UserMarkID:   2,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID: 1,
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "1",
						},
						{
							UserMarkID:   2,
							LocationID:   1,
							UserMarkGUID: "1",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Cleanup duplicates",
			args: args{
				merged: &model.Database{
					BlockRange: []*model.BlockRange{
						nil,
						{
							BlockRangeID: 1,
							UserMarkID:   1,
						},
						nil,
						{
							BlockRangeID: 3,
							UserMarkID:   2,
						},
						{
							BlockRangeID: 4,
							UserMarkID:   3,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID: 1,
							DocumentID: sql.NullInt32{6789, true},
							KeySymbol:  sql.NullString{"nwtsty", true},
						},
						{
							LocationID: 2,
							DocumentID: sql.NullInt32{12345, true},
							KeySymbol:  sql.NullString{"nwt", true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "SAME",
						},
						{
							UserMarkID:   2,
							LocationID:   2,
							UserMarkGUID: "SAME",
						},
						{
							UserMarkID:   1,
							LocationID:   2,
							UserMarkGUID: "3",
						},
					},
				},
			},
			want: &model.Database{
				BlockRange: []*model.BlockRange{
					nil,
					{
						BlockRangeID: 1,
						UserMarkID:   1,
					},
					nil,
					nil,
					{
						BlockRangeID: 4,
						UserMarkID:   3,
					},
				},
				Location: []*model.Location{
					nil,
					{
						LocationID: 1,
						DocumentID: sql.NullInt32{6789, true},
						KeySymbol:  sql.NullString{"nwtsty", true},
					},
					{
						LocationID: 2,
						DocumentID: sql.NullInt32{12345, true},
						KeySymbol:  sql.NullString{"nwt", true},
					},
				},
				UserMark: []*model.UserMark{
					nil,
					{
						UserMarkID:   1,
						LocationID:   1,
						UserMarkGUID: "SAME",
					},
					nil,
					{
						UserMarkID:   1,
						LocationID:   2,
						UserMarkGUID: "3",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PrepareDatabasesPostMerge(tt.args.merged)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, tt.args.merged)
		})
	}
}

func Test_needsNwtstyMigration(t *testing.T) {
	type args struct {
		left  *model.Database
		right *model.Database
	}
	tests := []struct {
		name string
		args args
		want map[int]MergeSide
	}{
		{
			name: "Nothing to migrate",
			args: args{
				left: &model.Database{
					Location: []*model.Location{
						nil,
						{
							LocationID:   1,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   2,
							KeySymbol:    sql.NullString{"somethingElse", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						nil,
						{
							LocationID:   4,
							KeySymbol:    sql.NullString{"bla", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   5,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						nil,
						nil,
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "1",
						},
						nil,
						nil,
						{
							UserMarkID:   4,
							LocationID:   1,
							UserMarkGUID: "4",
						},
						{
							UserMarkID:   5,
							LocationID:   4,
							UserMarkGUID: "5",
						},
					},
				},
				right: &model.Database{
					Location: []*model.Location{
						nil,
						{
							LocationID:   1,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   2,
							KeySymbol:    sql.NullString{"somethingElse", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						nil,
						{
							LocationID:   4,
							KeySymbol:    sql.NullString{"bla", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   5,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   4,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						nil,
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "1",
						},
						nil,
						nil,
						{
							UserMarkID:   4,
							LocationID:   1,
							UserMarkGUID: "4",
						},
						{
							UserMarkID:   5,
							LocationID:   4,
							UserMarkGUID: "5",
						},
						{
							UserMarkID:   6,
							LocationID:   4,
							UserMarkGUID: "6",
						},
					},
				},
			},
			want: map[int]MergeSide{},
		},
		{
			name: "Partially migrate",
			args: args{
				left: &model.Database{
					Location: []*model.Location{
						nil,
						{
							LocationID:   1,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   2,
							KeySymbol:    sql.NullString{"somethingElse", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						nil,
						{
							LocationID:   4,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 3, Valid: true},
						},
						{
							LocationID:   5,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						nil,
						nil,
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "1",
						},
						nil,
						nil,
						{
							UserMarkID:   4,
							LocationID:   1,
							UserMarkGUID: "4",
						},
						{
							UserMarkID:   5,
							LocationID:   4,
							UserMarkGUID: "5",
						},
						{
							UserMarkID:   6,
							LocationID:   4,
							UserMarkGUID: "6",
						},
					},
				},
				right: &model.Database{
					Location: []*model.Location{
						nil,
						{
							LocationID:   1,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   2,
							KeySymbol:    sql.NullString{"somethingElse", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						nil,
						{
							LocationID:   4,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 3, Valid: true},
						},
						{
							LocationID:   5,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   4,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						nil,
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "1",
						},
						nil,
						nil,
						{
							UserMarkID:   4,
							LocationID:   1,
							UserMarkGUID: "4",
						},
						{
							UserMarkID:   5,
							LocationID:   4,
							UserMarkGUID: "5",
						},
					},
				},
			},
			want: map[int]MergeSide{
				1: LeftSide,
				3: RightSide,
			},
		},
		{
			name: "Migrate",
			args: args{
				left: &model.Database{
					Location: []*model.Location{
						nil,
						{
							LocationID:   1,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   2,
							KeySymbol:    sql.NullString{"somethingElse", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   3,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   4,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   5,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 3, Valid: true},
						},
						nil,
						nil,
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "1",
						},
						nil,
						nil,
						{
							UserMarkID:   4,
							LocationID:   3,
							UserMarkGUID: "4",
						},
						{
							UserMarkID:   5,
							LocationID:   5,
							UserMarkGUID: "5",
						},
					},
				},
				right: &model.Database{
					Location: []*model.Location{
						nil,
						{
							LocationID:   1,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   2,
							KeySymbol:    sql.NullString{"somethingElse", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   3,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   4,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   5,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 3, Valid: true},
						},
						nil,
						nil,
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "1",
						},
						nil,
						nil,
						{
							UserMarkID:   4,
							LocationID:   3,
							UserMarkGUID: "4",
						},
						{
							UserMarkID:   5,
							LocationID:   5,
							UserMarkGUID: "5",
						},
					},
				},
			},
			want: map[int]MergeSide{
				1: LeftSide,
				2: LeftSide,
				3: LeftSide,
			},
		},
		{
			name: "All right",
			args: args{
				left: &model.Database{
					Location: []*model.Location{
						nil,
						{
							LocationID:   1,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   2,
							KeySymbol:    sql.NullString{"somethingElse", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   3,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   4,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   5,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						nil,
						nil,
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "1",
						},
						nil,
						nil,
						{
							UserMarkID:   4,
							LocationID:   3,
							UserMarkGUID: "4",
						},
						{
							UserMarkID:   5,
							LocationID:   5,
							UserMarkGUID: "5",
						},
					},
				},
				right: &model.Database{
					Location: []*model.Location{
						nil,
						{
							LocationID:   1,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   2,
							KeySymbol:    sql.NullString{"somethingElse", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   3,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   4,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   5,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						nil,
						nil,
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "1",
						},
						nil,
						nil,
						{
							UserMarkID:   4,
							LocationID:   3,
							UserMarkGUID: "4",
						},
						{
							UserMarkID:   5,
							LocationID:   5,
							UserMarkGUID: "5",
						},
					},
				},
			},
			want: map[int]MergeSide{
				2: RightSide,
			},
		},
		{
			name: "All left",
			args: args{
				left: &model.Database{
					Location: []*model.Location{
						nil,
						{
							LocationID:   1,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   2,
							KeySymbol:    sql.NullString{"somethingElse", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   3,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   4,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   5,
							KeySymbol:    sql.NullString{"nwt", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						nil,
						nil,
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "1",
						},
						nil,
						nil,
						{
							UserMarkID:   4,
							LocationID:   3,
							UserMarkGUID: "4",
						},
						{
							UserMarkID:   5,
							LocationID:   5,
							UserMarkGUID: "5",
						},
					},
				},
				right: &model.Database{
					Location: []*model.Location{
						nil,
						{
							LocationID:   1,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   2,
							KeySymbol:    sql.NullString{"somethingElse", true},
							MepsLanguage: sql.NullInt32{Int32: 1, Valid: true},
						},
						{
							LocationID:   3,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   4,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						{
							LocationID:   5,
							KeySymbol:    sql.NullString{"nwtsty", true},
							MepsLanguage: sql.NullInt32{Int32: 2, Valid: true},
						},
						nil,
						nil,
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "1",
						},
						nil,
						nil,
						{
							UserMarkID:   4,
							LocationID:   3,
							UserMarkGUID: "4",
						},
						{
							UserMarkID:   5,
							LocationID:   5,
							UserMarkGUID: "5",
						},
					},
				},
			},
			want: map[int]MergeSide{
				2: LeftSide,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, needsNwtstyMigration(tt.args.left, tt.args.right))
		})
	}
}

func Test_moveToNwtsty(t *testing.T) {
	type args struct {
		langs map[int]MergeSide
		left  []*model.Location
		right []*model.Location
	}
	tests := []struct {
		name string
		args args
		want args
	}{
		{
			args: args{
				langs: map[int]MergeSide{
					0: LeftSide,
					1: RightSide,
				},
				left: []*model.Location{
					{KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{KeySymbol: sql.NullString{"other", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{KeySymbol: sql.NullString{"other", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
				},
				right: []*model.Location{
					{KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
				},
			},
			want: args{
				left: []*model.Location{
					{KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{KeySymbol: sql.NullString{"other", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{KeySymbol: sql.NullString{"other", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
				},
				right: []*model.Location{
					{KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
				},
			},
		},
		{
			name: "Skip locations with DocID",
			args: args{
				langs: map[int]MergeSide{
					1: RightSide,
				},
				left: []*model.Location{
					{KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{KeySymbol: sql.NullString{"other", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{KeySymbol: sql.NullString{"other", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{DocumentID: sql.NullInt32{1, true}, KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{Track: sql.NullInt32{1, true}, KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
				},
				right: []*model.Location{
					{KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{DocumentID: sql.NullInt32{1, true}, KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{Track: sql.NullInt32{1, true}, KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
				},
			},
			want: args{
				left: []*model.Location{
					{KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{KeySymbol: sql.NullString{"other", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{KeySymbol: sql.NullString{"other", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{DocumentID: sql.NullInt32{1, true}, KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{Track: sql.NullInt32{1, true}, KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
				},
				right: []*model.Location{
					{KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{KeySymbol: sql.NullString{"nwtsty", true}, MepsLanguage: sql.NullInt32{Int32: 0, Valid: true}},
					{DocumentID: sql.NullInt32{1, true}, KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
					{Track: sql.NullInt32{1, true}, KeySymbol: sql.NullString{"nwt", true}, MepsLanguage: sql.NullInt32{Int32: 1, Valid: true}},
				},
			},
		},
	}
	for _, tt := range tests {
		moveToNwtsty(tt.args.langs, tt.args.left, tt.args.right)
		assert.Equal(t, tt.want.left, tt.args.left, tt.args.left)
		assert.Equal(t, tt.want.right, tt.args.right, tt.args.right)
	}
}

func Test_cleanupDuplicateLocations(t *testing.T) {
	type args struct {
		entries []*model.Location
	}
	tests := []struct {
		name          string
		args          args
		wantLocations []*model.Location
		wantChanges   map[int]int
	}{
		{
			name: "No duplicates",
			args: args{
				entries: []*model.Location{
					nil,
					{
						LocationID:     1,
						BookNumber:     sql.NullInt32{Int32: 1, Valid: true},
						ChapterNumber:  sql.NullInt32{Int32: 1, Valid: true},
						DocumentID:     sql.NullInt32{},
						Track:          sql.NullInt32{},
						IssueTagNumber: 0,
						KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
						MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
						LocationType:   0,
						Title:          sql.NullString{String: "", Valid: true},
					},
					nil,
					{
						LocationID:     3,
						BookNumber:     sql.NullInt32{Int32: 2, Valid: true},
						ChapterNumber:  sql.NullInt32{Int32: 4, Valid: true},
						DocumentID:     sql.NullInt32{},
						Track:          sql.NullInt32{},
						IssueTagNumber: 0,
						KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
						MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
						LocationType:   0,
						Title:          sql.NullString{String: "A", Valid: true},
					},
					{
						LocationID:     4,
						BookNumber:     sql.NullInt32{Int32: 5, Valid: true},
						ChapterNumber:  sql.NullInt32{Int32: 10, Valid: true},
						DocumentID:     sql.NullInt32{},
						Track:          sql.NullInt32{},
						IssueTagNumber: 0,
						KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
						MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
						LocationType:   0,
						Title:          sql.NullString{String: "B", Valid: true},
					},
				},
			},
			wantLocations: []*model.Location{
				nil,
				{
					LocationID:     1,
					BookNumber:     sql.NullInt32{Int32: 1, Valid: true},
					ChapterNumber:  sql.NullInt32{Int32: 1, Valid: true},
					DocumentID:     sql.NullInt32{},
					Track:          sql.NullInt32{},
					IssueTagNumber: 0,
					KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
					MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
					LocationType:   0,
					Title:          sql.NullString{String: "", Valid: true},
				},
				{
					LocationID:     2,
					BookNumber:     sql.NullInt32{Int32: 2, Valid: true},
					ChapterNumber:  sql.NullInt32{Int32: 4, Valid: true},
					DocumentID:     sql.NullInt32{},
					Track:          sql.NullInt32{},
					IssueTagNumber: 0,
					KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
					MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
					LocationType:   0,
					Title:          sql.NullString{String: "A", Valid: true},
				},
				{
					LocationID:     3,
					BookNumber:     sql.NullInt32{Int32: 5, Valid: true},
					ChapterNumber:  sql.NullInt32{Int32: 10, Valid: true},
					DocumentID:     sql.NullInt32{},
					Track:          sql.NullInt32{},
					IssueTagNumber: 0,
					KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
					MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
					LocationType:   0,
					Title:          sql.NullString{String: "B", Valid: true},
				},
			},
			wantChanges: map[int]int{
				3: 2,
				4: 3,
			},
		},
		{
			name: "Duplicate, both have no title",
			args: args{
				entries: []*model.Location{
					nil,
					{
						LocationID:     1,
						BookNumber:     sql.NullInt32{Int32: 1, Valid: true},
						ChapterNumber:  sql.NullInt32{Int32: 1, Valid: true},
						DocumentID:     sql.NullInt32{},
						Track:          sql.NullInt32{},
						IssueTagNumber: 0,
						KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
						MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
						LocationType:   0,
						Title:          sql.NullString{String: "", Valid: true},
					},
					{
						LocationID:     2,
						BookNumber:     sql.NullInt32{Int32: 2, Valid: true},
						ChapterNumber:  sql.NullInt32{Int32: 4, Valid: true},
						DocumentID:     sql.NullInt32{},
						Track:          sql.NullInt32{},
						IssueTagNumber: 0,
						KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
						MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
						LocationType:   0,
						Title:          sql.NullString{String: "A", Valid: true},
					},
					{
						LocationID:     3,
						BookNumber:     sql.NullInt32{Int32: 1, Valid: true},
						ChapterNumber:  sql.NullInt32{Int32: 1, Valid: true},
						DocumentID:     sql.NullInt32{},
						Track:          sql.NullInt32{},
						IssueTagNumber: 0,
						KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
						MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
						LocationType:   0,
						Title:          sql.NullString{String: "", Valid: true},
					},
				},
			},
			wantLocations: []*model.Location{
				nil,
				{
					LocationID:     1,
					BookNumber:     sql.NullInt32{Int32: 1, Valid: true},
					ChapterNumber:  sql.NullInt32{Int32: 1, Valid: true},
					DocumentID:     sql.NullInt32{},
					Track:          sql.NullInt32{},
					IssueTagNumber: 0,
					KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
					MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
					LocationType:   0,
					Title:          sql.NullString{String: "", Valid: true},
				},
				{
					LocationID:     2,
					BookNumber:     sql.NullInt32{Int32: 2, Valid: true},
					ChapterNumber:  sql.NullInt32{Int32: 4, Valid: true},
					DocumentID:     sql.NullInt32{},
					Track:          sql.NullInt32{},
					IssueTagNumber: 0,
					KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
					MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
					LocationType:   0,
					Title:          sql.NullString{String: "A", Valid: true},
				},
			},
			wantChanges: map[int]int{
				3: 1,
			},
		},
		{
			name: "Duplicate, keep title",
			args: args{
				entries: []*model.Location{
					nil,
					{
						LocationID:     1,
						BookNumber:     sql.NullInt32{Int32: 1, Valid: true},
						ChapterNumber:  sql.NullInt32{Int32: 1, Valid: true},
						DocumentID:     sql.NullInt32{},
						Track:          sql.NullInt32{},
						IssueTagNumber: 0,
						KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
						MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
						LocationType:   0,
						Title:          sql.NullString{String: "PleaseKeepMe", Valid: true},
					},
					{
						LocationID:     2,
						BookNumber:     sql.NullInt32{Int32: 2, Valid: true},
						ChapterNumber:  sql.NullInt32{Int32: 4, Valid: true},
						DocumentID:     sql.NullInt32{},
						Track:          sql.NullInt32{},
						IssueTagNumber: 0,
						KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
						MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
						LocationType:   0,
						Title:          sql.NullString{String: "A", Valid: true},
					},
					{
						LocationID:     3,
						BookNumber:     sql.NullInt32{Int32: 1, Valid: true},
						ChapterNumber:  sql.NullInt32{Int32: 1, Valid: true},
						DocumentID:     sql.NullInt32{},
						Track:          sql.NullInt32{},
						IssueTagNumber: 0,
						KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
						MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
						LocationType:   0,
						Title:          sql.NullString{String: "", Valid: true},
					},
					{
						LocationID:     4,
						BookNumber:     sql.NullInt32{Int32: 1, Valid: true},
						ChapterNumber:  sql.NullInt32{Int32: 1, Valid: true},
						DocumentID:     sql.NullInt32{},
						Track:          sql.NullInt32{},
						IssueTagNumber: 0,
						KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
						MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
						LocationType:   0,
						Title:          sql.NullString{String: "PleaseKeepMe", Valid: true},
					},
				},
			},
			wantLocations: []*model.Location{
				nil,
				{
					LocationID:     1,
					BookNumber:     sql.NullInt32{Int32: 1, Valid: true},
					ChapterNumber:  sql.NullInt32{Int32: 1, Valid: true},
					DocumentID:     sql.NullInt32{},
					Track:          sql.NullInt32{},
					IssueTagNumber: 0,
					KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
					MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
					LocationType:   0,
					Title:          sql.NullString{String: "PleaseKeepMe", Valid: true},
				},
				{
					LocationID:     2,
					BookNumber:     sql.NullInt32{Int32: 2, Valid: true},
					ChapterNumber:  sql.NullInt32{Int32: 4, Valid: true},
					DocumentID:     sql.NullInt32{},
					Track:          sql.NullInt32{},
					IssueTagNumber: 0,
					KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
					MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
					LocationType:   0,
					Title:          sql.NullString{String: "A", Valid: true},
				},
			},
			wantChanges: map[int]int{
				3: 1,
				4: 1,
			},
		},
		{
			name: "nil",
			args: args{
				entries: nil,
			},
			wantLocations: nil,
			wantChanges:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locations, changes := cleanupDuplicateLocations(tt.args.entries)
			assert.Equal(t, tt.wantLocations, locations)
			assert.Equal(t, tt.wantChanges, changes)
		})
	}
}

func Test_detectDuplicateUserMarks(t *testing.T) {
	type args struct {
		userMarks []*model.UserMark
	}
	tests := []struct {
		name string
		args args
		want map[string][]*model.UserMark
	}{
		{
			name: "No duplicates",
			args: args{
				userMarks: []*model.UserMark{
					nil,
					{
						UserMarkGUID: "A",
					},
					nil,
					nil,
					{
						UserMarkGUID: "B",
					},
					{
						UserMarkGUID: "C",
					},
					nil,
				},
			},
			want: map[string][]*model.UserMark{},
		},
		{
			name: "Duplicates",
			args: args{
				userMarks: []*model.UserMark{
					nil,
					{
						UserMarkID:   1,
						UserMarkGUID: "A",
					},
					nil,
					{
						UserMarkID:   3,
						UserMarkGUID: "A",
					},
					{
						UserMarkID:   4,
						UserMarkGUID: "B",
					},
					{
						UserMarkID:   5,
						UserMarkGUID: "A",
					},
					{
						UserMarkID:   6,
						UserMarkGUID: "B",
					},
					{
						UserMarkID:   7,
						UserMarkGUID: "C",
					},
				},
			},
			want: map[string][]*model.UserMark{
				"A": {
					&model.UserMark{
						UserMarkID:   1,
						UserMarkGUID: "A",
					},
					&model.UserMark{
						UserMarkID:   3,
						UserMarkGUID: "A",
					},
					&model.UserMark{
						UserMarkID:   5,
						UserMarkGUID: "A",
					},
				},
				"B": {
					&model.UserMark{
						UserMarkID:   4,
						UserMarkGUID: "B",
					},
					&model.UserMark{
						UserMarkID:   6,
						UserMarkGUID: "B",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, detectDuplicateUserMarks(tt.args.userMarks))
	}
}

func Test_deleteUserMark(t *testing.T) {
	type args struct {
		db *model.Database
		um *model.UserMark
	}
	tests := []struct {
		name   string
		args   args
		result *model.Database
	}{
		{
			args: args{
				db: &model.Database{
					BlockRange: []*model.BlockRange{
						nil,
						{
							BlockRangeID: 1,
							UserMarkID:   1,
						},
						nil,
						{
							BlockRangeID: 3,
							UserMarkID:   1,
						},
						{
							BlockRangeID: 4,
							UserMarkID:   2,
						},
						{
							BlockRangeID: 5,
							UserMarkID:   10,
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID: 1,
						},
						{
							UserMarkID: 2,
						},
						{
							UserMarkID: 10,
						},
					},
				},
				um: &model.UserMark{UserMarkID: 1},
			},
			result: &model.Database{
				BlockRange: []*model.BlockRange{
					nil,
					nil,
					nil,
					nil,
					{
						BlockRangeID: 4,
						UserMarkID:   2,
					},
					{
						BlockRangeID: 5,
						UserMarkID:   10,
					},
				},
				UserMark: []*model.UserMark{
					nil,
					nil,
					{
						UserMarkID: 2,
					},
					{
						UserMarkID: 10,
					},
				},
			},
		},
		{
			name: "nil",
			args: args{
				db: &model.Database{},
				um: nil,
			},
			result: &model.Database{},
		},
	}
	for _, tt := range tests {
		deleteUserMark(tt.args.db, tt.args.um)
		assert.True(t, tt.result.Equals(tt.args.db))
	}
}

func Test_tryDuplicateUserMarkCleanup(t *testing.T) {
	type args struct {
		db         *model.Database
		duplicates map[string][]*model.UserMark
	}
	tests := []struct {
		name        string
		args        args
		errContains string
		wantResult  *model.Database
	}{
		{
			name: "Too many duplicates",
			args: args{
				duplicates: map[string][]*model.UserMark{
					"A": {nil, nil, nil},
				},
			},
			errContains: "there are more than two 2 userMarks with the same GUID",
		},
		{
			name: "Too few duplicates",
			args: args{
				duplicates: map[string][]*model.UserMark{
					"A": {nil},
				},
			},
			errContains: "there are more than two 2 userMarks with the same GUID",
		},
		{
			name: "Location #1 missing",
			args: args{
				db: &model.Database{
					Location: []*model.Location{},
				},
				duplicates: map[string][]*model.UserMark{
					"A": {
						{
							LocationID: 1,
						},
						{
							LocationID: 2,
						},
					},
				},
			},
			errContains: "could not fetch location for duplicate userMark #1",
		},
		{
			name: "Location #2 missing",
			args: args{
				db: &model.Database{
					Location: []*model.Location{
						nil,
						{
							LocationID: 1,
						},
					},
				},
				duplicates: map[string][]*model.UserMark{
					"A": {
						{
							LocationID: 1,
						},
						{
							LocationID: 2,
						},
					},
				},
			},
			errContains: "could not fetch location for duplicate userMark #2",
		},
		{
			name: "Success",
			args: args{
				db: &model.Database{
					BlockRange: []*model.BlockRange{
						nil,
						{
							BlockRangeID: 1,
							Identifier:   1,
							UserMarkID:   1,
						},
						{
							BlockRangeID: 2,
							Identifier:   1,
							UserMarkID:   2,
						},
						nil,
						{
							BlockRangeID: 4,
							Identifier:   2,
							UserMarkID:   1,
						},
						{
							BlockRangeID: 5,
							Identifier:   3,
							UserMarkID:   1,
						},
						{
							BlockRangeID: 6,
							Identifier:   1,
							UserMarkID:   3,
						},
					},
					Location: []*model.Location{
						nil,
						{
							LocationID: 1,
							KeySymbol:  sql.NullString{String: "nwt", Valid: true},
						},
						nil,
						{
							LocationID: 3,
							KeySymbol:  sql.NullString{String: "something", Valid: true},
						},
						{
							LocationID: 4,
							KeySymbol:  sql.NullString{String: "nwtsty", Valid: true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "Duplicate",
						},
						nil,
						{
							UserMarkID:   3,
							LocationID:   3,
							UserMarkGUID: "Something",
						},
						{
							UserMarkID:   4,
							LocationID:   4,
							UserMarkGUID: "Duplicate",
						},
					},
				},
				duplicates: map[string][]*model.UserMark{
					"Duplicate": {
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "Duplicate",
						},
						{
							UserMarkID:   4,
							LocationID:   4,
							UserMarkGUID: "Duplicate",
						},
					},
				},
			},
			wantResult: &model.Database{
				BlockRange: []*model.BlockRange{
					nil,
					nil,
					{
						BlockRangeID: 2,
						Identifier:   1,
						UserMarkID:   2,
					},
					nil,
					nil,
					nil,
					{
						BlockRangeID: 6,
						Identifier:   1,
						UserMarkID:   3,
					},
				},
				Location: []*model.Location{
					nil,
					{
						LocationID: 1,
						KeySymbol:  sql.NullString{String: "nwt", Valid: true},
					},
					nil,
					{
						LocationID: 3,
						KeySymbol:  sql.NullString{String: "something", Valid: true},
					},
					{
						LocationID: 4,
						KeySymbol:  sql.NullString{String: "nwtsty", Valid: true},
					},
				},
				UserMark: []*model.UserMark{
					nil,
					nil,
					nil,
					{
						UserMarkID:   3,
						LocationID:   3,
						UserMarkGUID: "Something",
					},
					{
						UserMarkID:   4,
						LocationID:   4,
						UserMarkGUID: "Duplicate",
					},
				},
			},
		},
		{
			name: "Fail because both are nwtsty",
			args: args{
				db: &model.Database{
					Location: []*model.Location{
						nil,
						{
							LocationID: 1,
							KeySymbol:  sql.NullString{String: "nwtsty", Valid: true},
						},
						nil,
						{
							LocationID: 3,
							KeySymbol:  sql.NullString{String: "something", Valid: true},
						},
						{
							LocationID: 4,
							KeySymbol:  sql.NullString{String: "nwtsty", Valid: true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "Duplicate",
						},
						nil,
						{
							UserMarkID:   3,
							LocationID:   3,
							UserMarkGUID: "Something",
						},
						{
							UserMarkID:   4,
							LocationID:   4,
							UserMarkGUID: "Duplicate",
						},
					},
				},
				duplicates: map[string][]*model.UserMark{
					"Duplicate": {
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "Duplicate",
						},
						{
							UserMarkID:   4,
							LocationID:   4,
							UserMarkGUID: "Duplicate",
						},
					},
				},
			},
			errContains: "there are two userMarks with the same GUID that were not caused by migrating from nwt to nwtsty",
		},
		{
			name: "Fail because both are nwt",
			args: args{
				db: &model.Database{
					Location: []*model.Location{
						nil,
						{
							LocationID: 1,
							KeySymbol:  sql.NullString{String: "nwt", Valid: true},
						},
						nil,
						{
							LocationID: 3,
							KeySymbol:  sql.NullString{String: "something", Valid: true},
						},
						{
							LocationID: 4,
							KeySymbol:  sql.NullString{String: "nwt", Valid: true},
						},
					},
					UserMark: []*model.UserMark{
						nil,
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "Duplicate",
						},
						nil,
						{
							UserMarkID:   3,
							LocationID:   3,
							UserMarkGUID: "Something",
						},
						{
							UserMarkID:   4,
							LocationID:   4,
							UserMarkGUID: "Duplicate",
						},
					},
				},
				duplicates: map[string][]*model.UserMark{
					"Duplicate": {
						{
							UserMarkID:   1,
							LocationID:   1,
							UserMarkGUID: "Duplicate",
						},
						{
							UserMarkID:   4,
							LocationID:   4,
							UserMarkGUID: "Duplicate",
						},
					},
				},
			},
			errContains: "there are two userMarks with the same GUID that were not caused by migrating from nwt to nwtsty",
		},
	}
	for _, tt := range tests {
		err := tryDuplicateUserMarkCleanup(tt.args.db, tt.args.duplicates)
		if tt.errContains != "" {
			assert.Error(t, err, tt.name)
			assert.True(t, strings.Contains(err.Error(), tt.errContains), tt.name)
			continue
		}
		assert.NoError(t, err, tt.name)
		fmt.Println(tt.args.db.BlockRange)
		assert.True(t, tt.wantResult.Equals(tt.args.db), tt.name)
	}
}
