package model

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type occurrenceSpec struct {
	action   int
	start    int64
	duration int64
}

func occurrenceDB(specs ...occurrenceSpec) *Database {
	db := &Database{}
	tag := &Tag{TagID: 1, TagType: 2, Name: "Repeated"}
	media := &IndependentMedia{IndependentMediaID: 1, Hash: "same-media"}
	for i, spec := range specs {
		id := i + 1
		item := &PlaylistItem{
			PlaylistItemID: id, Label: "clip", Accuracy: 1,
			EndAction: spec.action, repeatIndex: i, tag: tag, independentMedia: media,
		}
		if spec.start >= 0 {
			item.StartTrimOffsetTicks = sql.NullInt64{Int64: spec.start, Valid: true}
		}
		db.PlaylistItem = append(db.PlaylistItem, item)
		db.PlaylistItemIndependentMediaMap = append(db.PlaylistItemIndependentMediaMap,
			&PlaylistItemIndependentMediaMap{PlaylistItemID: id, IndependentMediaID: 1, DurationTicks: spec.duration})
	}
	return db
}

func occurrences(db *Database) []int {
	result := make([]int, len(db.PlaylistItem))
	for i, item := range db.PlaylistItem {
		result[i] = item.repeatIndex
	}
	return result
}

func TestAlignPlaylistRepeats(t *testing.T) {
	first := occurrenceSpec{action: 0, start: 0, duration: 9}
	second := occurrenceSpec{action: 2, start: 10, duration: 9}
	inserted := occurrenceSpec{action: 3, start: 20, duration: 9}
	edited := occurrenceSpec{action: 1, start: 11, duration: 10}
	duplicateEdit := occurrenceSpec{action: 1, start: 10, duration: 9}
	tests := []struct {
		name        string
		left, right *Database
		wantRight   []int
	}{
		{"front insertion", occurrenceDB(first, second), occurrenceDB(inserted, first, second), []int{2, 0, 1}},
		{"middle insertion", occurrenceDB(first, second), occurrenceDB(first, inserted, second), []int{0, 2, 1}},
		{"end insertion", occurrenceDB(first, second), occurrenceDB(first, second, inserted), []int{0, 1, 2}},
		{"two insertions in one gap", occurrenceDB(first, second), occurrenceDB(first, inserted, occurrenceSpec{action: 1, start: 30, duration: 9}, second), []int{0, 2, 3, 1}},
		{"reverse insertion", occurrenceDB(inserted, first, second), occurrenceDB(first, second), []int{1, 2}},
		{"settings edits keep ordinal identity", occurrenceDB(first, second), occurrenceDB(occurrenceSpec{action: 1, start: 0, duration: 9}, occurrenceSpec{action: 3, start: 10, duration: 10}), []int{0, 1}},
		{"equal counts ignore settings reorder", occurrenceDB(first, second), occurrenceDB(second, first), []int{0, 1}},
		{"mapping duration anchors insertion", occurrenceDB(occurrenceSpec{duration: 10}, occurrenceSpec{duration: 20}), occurrenceDB(occurrenceSpec{duration: 30}, occurrenceSpec{duration: 10}, occurrenceSpec{duration: 20}), []int{2, 0, 1}},
		// When the counts between two matches differ, the right items count as new.
		{"insertion beside an edit keeps both versions", occurrenceDB(first, second), occurrenceDB(first, inserted, edited), []int{0, 2, 3}},
		{"duplicate with edits keeps both versions", occurrenceDB(first, first), occurrenceDB(first, duplicateEdit, duplicateEdit), []int{0, 2, 3}},
		{"reverse duplicate with edits keeps both versions", occurrenceDB(first, duplicateEdit, duplicateEdit), occurrenceDB(first, first), []int{0, 3}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			AlignPlaylistRepeats(tc.left, tc.right)
			if got := occurrences(tc.right); !reflect.DeepEqual(got, tc.wantRight) {
				t.Fatalf("occurrences = %v, want %v", got, tc.wantRight)
			}
			AlignPlaylistRepeats(tc.left, tc.right)
			if got := occurrences(tc.right); !reflect.DeepEqual(got, tc.wantRight) {
				t.Fatalf("repeat alignment = %v, want %v", got, tc.wantRight)
			}
		})
	}
}

func TestPlaylistChangedFields(t *testing.T) {
	left := &PlaylistItem{Label: "before"}
	right := &PlaylistItem{Label: "after", StartTrimOffsetTicks: sql.NullInt64{Int64: 10, Valid: true}, EndTrimOffsetTicks: sql.NullInt64{Int64: 20, Valid: true}, Accuracy: 1, EndAction: 2}
	require.Equal(t, []string{"label", "startTrimOffsetTicks", "endTrimOffsetTicks", "accuracy", "endAction"}, PlaylistChangedFields(left, right))
	require.Equal(t, "StartTrimOffsetTicks: Not set", PlaylistChangedFieldText(left, []string{"startTrimOffsetTicks"}))
	require.Equal(t, "Label: after\nEndAction: 2", PlaylistChangedFieldText(right, []string{"label", "endAction"}))
	require.Empty(t, PlaylistChangedFields(&Note{}, &Note{}))
}
