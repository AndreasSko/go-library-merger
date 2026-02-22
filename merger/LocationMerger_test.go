package merger

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/assert"
)

func Test_MergeLocationsNonTableTest(t *testing.T) {
	left := []*model.Location{
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
			Title:          sql.NullString{String: "ThisTitleShouldStay", Valid: true},
		},
		{
			LocationID:     3,
			BookNumber:     sql.NullInt32{Int32: 5, Valid: true},
			ChapterNumber:  sql.NullInt32{Int32: 8, Valid: true},
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
			BookNumber:     sql.NullInt32{Int32: 10, Valid: true},
			ChapterNumber:  sql.NullInt32{Int32: 8, Valid: true},
			DocumentID:     sql.NullInt32{},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "SomeTitle", Valid: true},
		},
		{
			LocationID:     5,
			BookNumber:     sql.NullInt32{},
			ChapterNumber:  sql.NullInt32{},
			DocumentID:     sql.NullInt32{Int32: 2020401, Valid: true},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "w", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "AWTTitle", Valid: true},
		},
		{
			LocationID:     6,
			BookNumber:     sql.NullInt32{},
			ChapterNumber:  sql.NullInt32{},
			DocumentID:     sql.NullInt32{Int32: 2020501, Valid: true},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "w", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "SomethingDuplicate", Valid: true},
		},
	}

	right := []*model.Location{
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
			Title:          sql.NullString{String: "ThisTitleShouldStay", Valid: true},
		},
		{
			LocationID:     3,
			BookNumber:     sql.NullInt32{Int32: 5, Valid: true},
			ChapterNumber:  sql.NullInt32{Int32: 8, Valid: true},
			DocumentID:     sql.NullInt32{},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "ThisTitleShouldStay", Valid: true},
		},
		{
			LocationID:     4,
			BookNumber:     sql.NullInt32{Int32: 6, Valid: true},
			ChapterNumber:  sql.NullInt32{Int32: 5, Valid: true},
			DocumentID:     sql.NullInt32{},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "SomeOtherTitle", Valid: true},
		},
		{
			LocationID:     5,
			BookNumber:     sql.NullInt32{},
			ChapterNumber:  sql.NullInt32{},
			DocumentID:     sql.NullInt32{Int32: 123456789, Valid: true},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "CO-pgm20", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "ATitle", Valid: true},
		},
		{
			LocationID:     6,
			BookNumber:     sql.NullInt32{},
			ChapterNumber:  sql.NullInt32{},
			DocumentID:     sql.NullInt32{Int32: 123456790, Valid: true},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "CO-pgm20", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "AATitle", Valid: true},
		},
		{
			LocationID:     7,
			BookNumber:     sql.NullInt32{},
			ChapterNumber:  sql.NullInt32{},
			DocumentID:     sql.NullInt32{Int32: 2020501, Valid: true},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "w", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "SomethingDuplicate", Valid: true},
		},
	}

	expectedResult := []*model.Location{
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
			Title:          sql.NullString{String: "ThisTitleShouldStay", Valid: true},
		},
		{
			LocationID:     3,
			BookNumber:     sql.NullInt32{Int32: 5, Valid: true},
			ChapterNumber:  sql.NullInt32{Int32: 8, Valid: true},
			DocumentID:     sql.NullInt32{},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "ThisTitleShouldStay", Valid: true},
		},
		{
			LocationID:     4,
			BookNumber:     sql.NullInt32{Int32: 10, Valid: true},
			ChapterNumber:  sql.NullInt32{Int32: 8, Valid: true},
			DocumentID:     sql.NullInt32{},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "SomeTitle", Valid: true},
		},
		{
			LocationID:     5,
			BookNumber:     sql.NullInt32{Int32: 6, Valid: true},
			ChapterNumber:  sql.NullInt32{Int32: 5, Valid: true},
			DocumentID:     sql.NullInt32{},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "SomeOtherTitle", Valid: true},
		},
		{
			LocationID:     6,
			BookNumber:     sql.NullInt32{},
			ChapterNumber:  sql.NullInt32{},
			DocumentID:     sql.NullInt32{Int32: 2020401, Valid: true},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "w", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "AWTTitle", Valid: true},
		},
		{
			LocationID:     7,
			BookNumber:     sql.NullInt32{},
			ChapterNumber:  sql.NullInt32{},
			DocumentID:     sql.NullInt32{Int32: 123456789, Valid: true},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "CO-pgm20", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "ATitle", Valid: true},
		},
		{
			LocationID:     8,
			BookNumber:     sql.NullInt32{},
			ChapterNumber:  sql.NullInt32{},
			DocumentID:     sql.NullInt32{Int32: 2020501, Valid: true},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "w", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "SomethingDuplicate", Valid: true},
		},
		{
			LocationID:     9,
			BookNumber:     sql.NullInt32{},
			ChapterNumber:  sql.NullInt32{},
			DocumentID:     sql.NullInt32{Int32: 123456790, Valid: true},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "CO-pgm20", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: "AATitle", Valid: true},
		},
	}

	expectedChanges := IDChanges{
		Left: map[int]int{
			5: 6,
			6: 8,
		},
		Right: map[int]int{
			4: 5,
			5: 7,
			6: 9,
			7: 8,
		},
	}

	result, changes, err := MergeLocations(left, right)

	assert.Equal(t, expectedResult, result)
	assert.Equal(t, expectedChanges, changes)
	assert.NoError(t, err)
	// Check if original has not been tweaked
	assert.Equal(t, 6, left[6].LocationID)
	assert.Equal(t, 7, right[7].LocationID)
}

func TestMergeLocations(t *testing.T) {
	tests := []struct {
		name        string
		left        []*model.Location
		right       []*model.Location
		wantMerged  []*model.Location
		wantChanges IDChanges
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name: "MepsLanguage 0 != NULL",
			left: []*model.Location{
				nil,
				{
					LocationID:     1,
					DocumentID:     sql.NullInt32{Int32: int32(123), Valid: true},
					IssueTagNumber: 0,
					KeySymbol:      sql.NullString{String: "bla", Valid: true},
					MepsLanguage:   sql.NullInt32{Int32: 0, Valid: true},
					LocationType:   0,
					Title:          sql.NullString{String: "Something in english", Valid: true},
				},
			},
			right: []*model.Location{
				nil,
				{
					LocationID:     1,
					DocumentID:     sql.NullInt32{Int32: int32(123), Valid: true},
					IssueTagNumber: 0,
					KeySymbol:      sql.NullString{String: "bla", Valid: true},
					MepsLanguage:   sql.NullInt32{},
					LocationType:   0,
					Title:          sql.NullString{String: "Something else", Valid: true},
				},
			},
			wantMerged: []*model.Location{
				nil,
				{
					LocationID:     1,
					DocumentID:     sql.NullInt32{Int32: int32(123), Valid: true},
					IssueTagNumber: 0,
					KeySymbol:      sql.NullString{String: "bla", Valid: true},
					MepsLanguage:   sql.NullInt32{Int32: 0, Valid: true},
					LocationType:   0,
					Title:          sql.NullString{String: "Something in english", Valid: true},
				},
				{
					LocationID:     2,
					DocumentID:     sql.NullInt32{Int32: int32(123), Valid: true},
					IssueTagNumber: 0,
					KeySymbol:      sql.NullString{String: "bla", Valid: true},
					MepsLanguage:   sql.NullInt32{},
					LocationType:   0,
					Title:          sql.NullString{String: "Something else", Valid: true},
				},
			},
			wantChanges: IDChanges{
				Left:  map[int]int{},
				Right: map[int]int{1: 2},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged, changes, err := MergeLocations(tt.left, tt.right)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.wantMerged, merged)
			assert.Equal(t, tt.wantChanges, changes)
		})
	}
}

func Benchmark_MergeLocations(b *testing.B) {
	const locationCount = 1000000
	left := make([]*model.Location, locationCount+1)
	right := make([]*model.Location, locationCount+1)

	// Duplicates
	for i := 1; i < locationCount+1; i++ {
		left[i] = &model.Location{
			LocationID:     i,
			BookNumber:     sql.NullInt32{Int32: int32(i), Valid: true},
			ChapterNumber:  sql.NullInt32{Int32: int32(i), Valid: true},
			DocumentID:     sql.NullInt32{},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: fmt.Sprint(i), Valid: true},
		}
		right[i] = &model.Location{
			LocationID:     i,
			BookNumber:     sql.NullInt32{Int32: int32(i), Valid: true},
			ChapterNumber:  sql.NullInt32{Int32: int32(i), Valid: true},
			DocumentID:     sql.NullInt32{},
			Track:          sql.NullInt32{},
			IssueTagNumber: 0,
			KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
			MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
			LocationType:   0,
			Title:          sql.NullString{String: fmt.Sprint(i), Valid: true},
		}
	}

	MergeLocations(left, right)
}

func Test_solveLocationMergeConflict(t *testing.T) {
	conflicts := map[string]MergeConflict{
		"ChooseLeftConflict": {
			Left: &model.Location{
				LocationID:     1,
				BookNumber:     sql.NullInt32{Int32: 1, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 1, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "ChooseLeft", Valid: true},
			},
			Right: &model.Location{
				LocationID:     5,
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
		"ChooseRightConflict": {
			Left: &model.Location{
				LocationID:     2,
				BookNumber:     sql.NullInt32{Int32: 2, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 2, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "", Valid: true},
			},
			Right: &model.Location{
				LocationID:     6,
				BookNumber:     sql.NullInt32{Int32: 2, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 2, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "ChooseRight", Valid: true},
			},
		},
		"ChooseRightBecauseBothEmpty": {
			Left: &model.Location{
				LocationID:     3,
				BookNumber:     sql.NullInt32{Int32: 3, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 3, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "", Valid: true},
			},
			Right: &model.Location{
				LocationID:     7,
				BookNumber:     sql.NullInt32{Int32: 3, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 3, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "", Valid: true},
			},
		},
	}

	expectedResult := map[string]MergeSolution{
		"ChooseLeftConflict": {
			Side: LeftSide,
			Solution: &model.Location{
				LocationID:     1,
				BookNumber:     sql.NullInt32{Int32: 1, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 1, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "ChooseLeft", Valid: true},
			},
			Discarded: &model.Location{
				LocationID:     5,
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
		"ChooseRightConflict": {
			Side: RightSide,
			Solution: &model.Location{
				LocationID:     6,
				BookNumber:     sql.NullInt32{Int32: 2, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 2, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "ChooseRight", Valid: true},
			},
			Discarded: &model.Location{
				LocationID:     2,
				BookNumber:     sql.NullInt32{Int32: 2, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 2, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "", Valid: true},
			},
		},
		"ChooseRightBecauseBothEmpty": {
			Side: RightSide,
			Solution: &model.Location{
				LocationID:     7,
				BookNumber:     sql.NullInt32{Int32: 3, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 3, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "", Valid: true},
			},
			Discarded: &model.Location{
				LocationID:     3,
				BookNumber:     sql.NullInt32{Int32: 3, Valid: true},
				ChapterNumber:  sql.NullInt32{Int32: 3, Valid: true},
				DocumentID:     sql.NullInt32{},
				Track:          sql.NullInt32{},
				IssueTagNumber: 0,
				KeySymbol:      sql.NullString{String: "nwtsty", Valid: true},
				MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
				LocationType:   0,
				Title:          sql.NullString{String: "", Valid: true},
			},
		},
	}

	result, _ := solveLocationMergeConflict(conflicts)

	assert.Equal(t, expectedResult, result)

	assert.PanicsWithValue(t, "No other type than *model.Location is supported! Given: *model.Bookmark", func() {
		panicConflict := map[string]MergeConflict{
			"WrongType": {
				Left:  &model.Bookmark{},
				Right: &model.Bookmark{},
			},
		}
		solveLocationMergeConflict(panicConflict)
	})
}
