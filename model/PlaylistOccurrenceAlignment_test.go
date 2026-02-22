package model

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"
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
			EndAction: spec.action, occurrence: i, tag: tag, independentMedia: media,
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
		result[i] = item.occurrence
	}
	return result
}

func TestAlignPlaylistItemOccurrencesInsertions(t *testing.T) {
	old1 := occurrenceSpec{action: 0, start: 0, duration: 9}
	old2 := occurrenceSpec{action: 2, start: 10, duration: 9}
	inserted := occurrenceSpec{action: 3, start: 20, duration: 9}
	tests := []struct {
		name        string
		left, right *Database
		wantRight   []int
	}{
		{"front", occurrenceDB(old1, old2), occurrenceDB(inserted, old1, old2), []int{2, 0, 1}},
		{"middle", occurrenceDB(old1, old2), occurrenceDB(old1, inserted, old2), []int{0, 2, 1}},
		{"end", occurrenceDB(old1, old2), occurrenceDB(old1, old2, inserted), []int{0, 1, 2}},
		{"two in one gap", occurrenceDB(old1, old2), occurrenceDB(old1, inserted, occurrenceSpec{action: 1, start: 30, duration: 9}, old2), []int{0, 2, 3, 1}},
		{"reverse inputs", occurrenceDB(inserted, old1, old2), occurrenceDB(old1, old2), []int{1, 2}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := AlignPlaylistItemOccurrences(tc.left, tc.right); err != nil {
				t.Fatal(err)
			}
			if got := occurrences(tc.right); !reflect.DeepEqual(got, tc.wantRight) {
				t.Fatalf("occurrences = %v, want %v", got, tc.wantRight)
			}
		})
	}
}

func TestAlignPlaylistItemOccurrencesKeepsEditsAsSameOccurrence(t *testing.T) {
	left := occurrenceDB(
		occurrenceSpec{action: 0, start: 0, duration: 9},
		occurrenceSpec{action: 2, start: 10, duration: 9},
	)
	right := occurrenceDB(
		occurrenceSpec{action: 1, start: 0, duration: 9},
		occurrenceSpec{action: 3, start: 10, duration: 10},
	)
	if err := AlignPlaylistItemOccurrences(left, right); err != nil {
		t.Fatal(err)
	}
	if got := occurrences(right); !reflect.DeepEqual(got, []int{0, 1}) {
		t.Fatalf("edited occurrences = %v", got)
	}
}

func TestAlignPlaylistItemOccurrencesDoesNotUseSettingsToReorderEqualCounts(t *testing.T) {
	first := occurrenceSpec{action: 0, start: 0, duration: 9}
	second := occurrenceSpec{action: 2, start: 10, duration: 9}
	left := occurrenceDB(first, second)
	right := occurrenceDB(second, first)
	if err := AlignPlaylistItemOccurrences(left, right); err != nil {
		t.Fatal(err)
	}
	if got := occurrences(right); !reflect.DeepEqual(got, []int{0, 1}) {
		t.Fatalf("equal-count reorder changed ordinal identity: %v", got)
	}
}

func TestAlignPlaylistItemOccurrencesUsesMappingDurationAsAnchor(t *testing.T) {
	left := occurrenceDB(occurrenceSpec{duration: 10}, occurrenceSpec{duration: 20})
	right := occurrenceDB(occurrenceSpec{duration: 30}, occurrenceSpec{duration: 10}, occurrenceSpec{duration: 20})
	if err := AlignPlaylistItemOccurrences(left, right); err != nil {
		t.Fatal(err)
	}
	if got := occurrences(right); !reflect.DeepEqual(got, []int{2, 0, 1}) {
		t.Fatalf("duration-aligned occurrences = %v", got)
	}
	before := occurrences(right)
	if err := AlignPlaylistItemOccurrences(left, right); err != nil {
		t.Fatal(err)
	}
	if got := occurrences(right); !reflect.DeepEqual(got, before) {
		t.Fatalf("repeat alignment changed %v to %v", before, got)
	}
}

func TestAlignPlaylistItemOccurrencesRejectsInsertionAdjacentToUnanchoredEdit(t *testing.T) {
	left := occurrenceDB(
		occurrenceSpec{action: 0, start: 0, duration: 9},
		occurrenceSpec{action: 2, start: 10, duration: 9},
	)
	right := occurrenceDB(
		occurrenceSpec{action: 0, start: 0, duration: 9},
		occurrenceSpec{action: 3, start: 20, duration: 9},
		occurrenceSpec{action: 1, start: 11, duration: 10},
	)
	if err := AlignPlaylistItemOccurrences(left, right); err == nil {
		t.Fatal("expected ambiguous insertion/edit to be rejected")
	}
}

func TestAlignPlaylistItemOccurrencesRejectsDuplicateAnchorWithUnmatchedEdit(t *testing.T) {
	old := occurrenceSpec{action: 0, start: 0, duration: 9}
	edited := occurrenceSpec{action: 1, start: 10, duration: 9}
	for _, reverse := range []bool{false, true} {
		left, right := occurrenceDB(old, old), occurrenceDB(old, edited, edited)
		if reverse {
			left, right = right, left
		}
		if err := AlignPlaylistItemOccurrences(left, right); err == nil {
			t.Fatalf("expected ambiguous duplicate anchor/edit to be rejected (reverse=%t)", reverse)
		}
	}
}

func TestAlignPlaylistItemOccurrencesBoundsEditedMatching(t *testing.T) {
	leftSpecs, rightSpecs := make([]occurrenceSpec, 2000), make([]occurrenceSpec, 2001)
	for i := range leftSpecs {
		leftSpecs[i] = occurrenceSpec{action: 0, start: 0, duration: 9}
	}
	for i := range rightSpecs {
		rightSpecs[i] = occurrenceSpec{action: 1, start: 10, duration: 9}
	}
	if err := AlignPlaylistItemOccurrences(occurrenceDB(leftSpecs...), occurrenceDB(rightSpecs...)); err == nil || !strings.Contains(err.Error(), "complexity") {
		t.Fatalf("expected bounded matching error, got %v", err)
	}
}
