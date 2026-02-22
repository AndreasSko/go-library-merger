package merger

import (
	"database/sql"
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/assert"
)

func TestMergeBookmarks(t *testing.T) {
	// Successfully merge on first try
	left := []*model.Bookmark{
		{
			BookmarkID:            1,
			LocationID:            1,
			PublicationLocationID: 1,
			Slot:                  1,
			Title:                 "CollisionThatShouldBeSolvedAutomatically",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            2,
			LocationID:            2,
			PublicationLocationID: 2,
			Slot:                  2,
			Title:                 "AnotherCollisionThatShouldBeSolvedAutomatically",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            3,
			LocationID:            4,
			PublicationLocationID: 4,
			Slot:                  4,
			Title:                 "ABookmark",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
	}
	right := []*model.Bookmark{
		{
			BookmarkID:            1,
			LocationID:            2,
			PublicationLocationID: 2,
			Slot:                  2,
			Title:                 "AnotherCollisionThatShouldBeSolvedAutomatically",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            2,
			LocationID:            1,
			PublicationLocationID: 1,
			Slot:                  1,
			Title:                 "CollisionThatShouldBeSolvedAutomatically",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            3,
			LocationID:            5,
			PublicationLocationID: 5,
			Slot:                  5,
			Title:                 "AnotherBookmark",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
	}

	expectedResult := []*model.Bookmark{
		nil,
		{
			BookmarkID:            1,
			LocationID:            1,
			PublicationLocationID: 1,
			Slot:                  1,
			Title:                 "CollisionThatShouldBeSolvedAutomatically",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            2,
			LocationID:            2,
			PublicationLocationID: 2,
			Slot:                  2,
			Title:                 "AnotherCollisionThatShouldBeSolvedAutomatically",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            3,
			LocationID:            4,
			PublicationLocationID: 4,
			Slot:                  4,
			Title:                 "ABookmark",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            4,
			LocationID:            5,
			PublicationLocationID: 5,
			Slot:                  5,
			Title:                 "AnotherBookmark",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
	}
	expectedChanges := IDChanges{
		Left: map[int]int{},
		Right: map[int]int{
			1: 2,
			2: 1,
			3: 4,
		},
	}

	result, changes, err := MergeBookmarks(left, right, nil)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, expectedChanges, changes)
	// Check if original has not been tweaked
	assert.Equal(t, 3, right[2].BookmarkID)

	// Fail with mergeConflict
	left = []*model.Bookmark{
		{
			BookmarkID:            1,
			LocationID:            1,
			PublicationLocationID: 1,
			Slot:                  1,
			Title:                 "CollisionOnLeft",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            2,
			LocationID:            2,
			PublicationLocationID: 2,
			Slot:                  2,
			Title:                 "ACollisionThatShouldBeSolvedAutomatically",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            3,
			LocationID:            4,
			PublicationLocationID: 4,
			Slot:                  4,
			Title:                 "ABookmark",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            4,
			LocationID:            6,
			PublicationLocationID: 6,
			Slot:                  6,
			Title:                 "DifferentTitleCollision",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
	}
	right = []*model.Bookmark{
		{
			BookmarkID:            1,
			LocationID:            2,
			PublicationLocationID: 2,
			Slot:                  2,
			Title:                 "ACollisionThatShouldBeSolvedAutomatically",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            2,
			LocationID:            10,
			PublicationLocationID: 1,
			Slot:                  1,
			Title:                 "CollisionOnRight",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            3,
			LocationID:            5,
			PublicationLocationID: 5,
			Slot:                  5,
			Title:                 "AnotherBookmark",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            4,
			LocationID:            6,
			PublicationLocationID: 6,
			Slot:                  6,
			Title:                 "DifferentTitleToLeftCollision",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
	}

	expectedConflicts := map[string]MergeConflict{
		"1_1": {
			Left: &model.Bookmark{
				BookmarkID:            1,
				LocationID:            1,
				PublicationLocationID: 1,
				Slot:                  1,
				Title:                 "CollisionOnLeft",
				Snippet:               sql.NullString{},
				BlockType:             0,
				BlockIdentifier:       sql.NullInt32{},
			},
			Right: &model.Bookmark{
				BookmarkID:            2,
				LocationID:            10,
				PublicationLocationID: 1,
				Slot:                  1,
				Title:                 "CollisionOnRight",
				Snippet:               sql.NullString{},
				BlockType:             0,
				BlockIdentifier:       sql.NullInt32{},
			},
		},
		"6_6": {
			Left: &model.Bookmark{
				BookmarkID:            4,
				LocationID:            6,
				PublicationLocationID: 6,
				Slot:                  6,
				Title:                 "DifferentTitleCollision",
				Snippet:               sql.NullString{},
				BlockType:             0,
				BlockIdentifier:       sql.NullInt32{},
			},
			Right: &model.Bookmark{
				BookmarkID:            4,
				LocationID:            6,
				PublicationLocationID: 6,
				Slot:                  6,
				Title:                 "DifferentTitleToLeftCollision",
				Snippet:               sql.NullString{},
				BlockType:             0,
				BlockIdentifier:       sql.NullInt32{},
			},
		},
	}

	_, _, err = MergeBookmarks(left, right, nil)
	assert.Error(t, err)
	assert.Equal(t, expectedConflicts, err.(MergeConflictError).Conflicts)

	// Succeed using mergeSolution
	conflictSolution := map[string]MergeSolution{
		"1_1": {
			Side: RightSide,
			Solution: &model.Bookmark{
				BookmarkID:            2,
				LocationID:            10,
				PublicationLocationID: 1,
				Slot:                  1,
				Title:                 "CollisionOnRight",
				Snippet:               sql.NullString{},
				BlockType:             0,
				BlockIdentifier:       sql.NullInt32{},
			},
			Discarded: &model.Bookmark{
				BookmarkID:            1,
				LocationID:            1,
				PublicationLocationID: 1,
				Slot:                  1,
				Title:                 "CollisionOnLeft",
				Snippet:               sql.NullString{},
				BlockType:             0,
				BlockIdentifier:       sql.NullInt32{},
			},
		},
		"6_6": {
			Side: LeftSide,
			Solution: &model.Bookmark{
				BookmarkID:            4,
				LocationID:            6,
				PublicationLocationID: 6,
				Slot:                  6,
				Title:                 "DifferentTitleCollision",
				Snippet:               sql.NullString{},
				BlockType:             0,
				BlockIdentifier:       sql.NullInt32{},
			},
			Discarded: &model.Bookmark{
				BookmarkID:            4,
				LocationID:            6,
				PublicationLocationID: 6,
				Slot:                  6,
				Title:                 "DifferentTitleToLeftCollision",
				Snippet:               sql.NullString{},
				BlockType:             0,
				BlockIdentifier:       sql.NullInt32{},
			},
		},
	}

	expectedResult = []*model.Bookmark{
		nil,
		{
			BookmarkID:            1,
			LocationID:            2,
			PublicationLocationID: 2,
			Slot:                  2,
			Title:                 "ACollisionThatShouldBeSolvedAutomatically",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            2,
			LocationID:            10,
			PublicationLocationID: 1,
			Slot:                  1,
			Title:                 "CollisionOnRight",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            3,
			LocationID:            4,
			PublicationLocationID: 4,
			Slot:                  4,
			Title:                 "ABookmark",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            4,
			LocationID:            5,
			PublicationLocationID: 5,
			Slot:                  5,
			Title:                 "AnotherBookmark",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
		{
			BookmarkID:            5,
			LocationID:            6,
			PublicationLocationID: 6,
			Slot:                  6,
			Title:                 "DifferentTitleCollision",
			Snippet:               sql.NullString{},
			BlockType:             0,
			BlockIdentifier:       sql.NullInt32{},
		},
	}

	expectedChanges = IDChanges{
		Left: map[int]int{
			1: 2,
			2: 1,
			4: 5,
		},
		Right: map[int]int{
			3: 4,
			4: 5,
		},
	}

	result, changes, err = MergeBookmarks(left, right, conflictSolution)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, expectedChanges, changes)
}
