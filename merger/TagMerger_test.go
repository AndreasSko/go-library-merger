package merger

import (
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/assert"
)

func TestMergeTags(t *testing.T) {
	// Successful merge
	left := []*model.Tag{
		{
			TagID:   1,
			TagType: 1,
			Name:    "A tag on the left",
		},
		{
			TagID:   2,
			TagType: 1,
			Name:    "Another tag on the left",
		},
		{
			TagID:   3,
			TagType: 1,
			Name:    "A duplicate tag",
		},
		{
			TagID:   4,
			TagType: 1,
			Name:    "A duplicate tag with imageFilename",
		},
		nil,
	}
	right := []*model.Tag{
		{
			TagID:   1,
			TagType: 1,
			Name:    "A tag on the right",
		},
		{
			TagID:   2,
			TagType: 1,
			Name:    "A duplicate tag",
		},
		{
			TagID:   3,
			TagType: 1,
			Name:    "A duplicate tag with imageFilename",
		},
		{
			TagID:   4,
			TagType: 1,
			Name:    "One more tag only on the right",
		},
	}

	expectedResult := []*model.Tag{
		nil,
		{
			TagID:   1,
			TagType: 1,
			Name:    "A tag on the left",
		},
		{
			TagID:   2,
			TagType: 1,
			Name:    "A tag on the right",
		},
		{
			TagID:   3,
			TagType: 1,
			Name:    "Another tag on the left",
		},
		{
			TagID:   4,
			TagType: 1,
			Name:    "A duplicate tag",
		},
		{
			TagID:   5,
			TagType: 1,
			Name:    "A duplicate tag with imageFilename",
		},
		{
			TagID:   6,
			TagType: 1,
			Name:    "One more tag only on the right",
		},
	}

	expectedChanges := IDChanges{
		Left: map[int]int{
			2: 3,
			3: 4,
			4: 5,
		},
		Right: map[int]int{
			1: 2,
			2: 4,
			3: 5,
			4: 6,
		},
	}

	result, changes, err := MergeTags(left, right, nil)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, expectedChanges, changes)
	// Check if original has not been tweaked
	assert.Equal(t, 1, left[0].TagID)
	assert.Equal(t, 1, right[0].TagID)
}
