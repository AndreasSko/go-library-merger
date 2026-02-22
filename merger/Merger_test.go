package merger

import (
	"database/sql"
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/assert"
)

func Test_sortMergeSolution(t *testing.T) {
	solution := []MergeSolution{
		{
			Side: RightSide,
			Solution: &model.Location{
				LocationID:     3,
				BookNumber:     sql.NullInt32{Int32: 1, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 1, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "ChooseRight", Valid: true},
			},
		},
		{
			Side: LeftSide,
			Solution: &model.Location{
				LocationID:     3,
				BookNumber:     sql.NullInt32{Int32: 2, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 2, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "ChooseLeft", Valid: true},
			},
		},
		{
			Side: LeftSide,
			Solution: &model.Location{
				LocationID:     2,
				BookNumber:     sql.NullInt32{Int32: 3, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 3, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "ChooseLeft", Valid: true},
			},
		},
		{
			Side: RightSide,
			Solution: &model.Location{
				LocationID:     1,
				BookNumber:     sql.NullInt32{Int32: 4, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 4, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "ChooseRight", Valid: true},
			},
		},
	}

	expectedResult := []MergeSolution{
		{
			Side: RightSide,
			Solution: &model.Location{
				LocationID:     1,
				BookNumber:     sql.NullInt32{Int32: 4, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 4, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "ChooseRight", Valid: true},
			},
		},
		{
			Side: LeftSide,
			Solution: &model.Location{
				LocationID:     2,
				BookNumber:     sql.NullInt32{Int32: 3, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 3, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "ChooseLeft", Valid: true},
			},
		},
		{
			Side: LeftSide,
			Solution: &model.Location{
				LocationID:     3,
				BookNumber:     sql.NullInt32{Int32: 2, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 2, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "ChooseLeft", Valid: true},
			},
		},
		{
			Side: RightSide,
			Solution: &model.Location{
				LocationID:     3,
				BookNumber:     sql.NullInt32{Int32: 1, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 1, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "ChooseRight", Valid: true},
			},
		},
	}

	sortMergeSolution(&solution)

	assert.Equal(t, expectedResult, solution)
}
