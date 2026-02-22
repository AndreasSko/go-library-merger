package merger

import (
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/assert"
)

func TestMergeInputFields_WithoutConflict(t *testing.T) {
	left := []*model.InputField{
		nil,
		{
			LocationID: 1,
			TextTag:    "a1",
			Value:      "a1",
		},
		{
			LocationID: 1,
			TextTag:    "b1",
			Value:      "b1",
		},
		{
			LocationID: 3,
			TextTag:    "c1",
			Value:      "c1",
		},
	}
	right := []*model.InputField{
		nil,
		{
			LocationID: 2,
			TextTag:    "a1",
			Value:      "a1",
		},
		{
			LocationID: 1,
			TextTag:    "a1",
			Value:      "a1",
		},
		{
			LocationID: 2,
			TextTag:    "d1",
			Value:      "d1",
		},
	}
	expectedResult := []*model.InputField{
		nil,
		{
			LocationID: 1,
			TextTag:    "a1",
			Value:      "a1",
		},
		{
			LocationID: 1,
			TextTag:    "b1",
			Value:      "b1",
		},
		{
			LocationID: 2,
			TextTag:    "a1",
			Value:      "a1",
		},
		{
			LocationID: 2,
			TextTag:    "d1",
			Value:      "d1",
		},
		{
			LocationID: 3,
			TextTag:    "c1",
			Value:      "c1",
		},
	}

	result, _, err := MergeInputFields(left, right, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestMergeInputFields_WithConflict(t *testing.T) {
	left := []*model.InputField{
		nil,
		{
			LocationID: 1,
			TextTag:    "a1",
			Value:      "a1",
		},
		{
			LocationID: 1,
			TextTag:    "b1",
			Value:      "b1",
		},
		{
			LocationID: 3,
			TextTag:    "c1",
			Value:      "c1",
		},
	}
	right := []*model.InputField{
		nil,
		{
			LocationID: 2,
			TextTag:    "d2",
			Value:      "d2",
		},
		{
			LocationID: 1,
			TextTag:    "a1",
			Value:      "different",
		},
		{
			LocationID: 2,
			TextTag:    "d1",
			Value:      "d1",
		},
		{
			LocationID: 3,
			TextTag:    "c1",
			Value:      "alsodifferent",
		},
		{
			LocationID: 1,
			TextTag:    "b1",
			Value:      "b1",
		},
	}
	expectedConflicts := map[string]MergeConflict{
		"1_a1": {
			Left: &model.InputField{
				LocationID: 1,
				TextTag:    "a1",
				Value:      "a1",
			},
			Right: &model.InputField{
				LocationID: 1,
				TextTag:    "a1",
				Value:      "different",
			},
		},
		"3_c1": {
			Left: &model.InputField{
				LocationID: 3,
				TextTag:    "c1",
				Value:      "c1",
			},
			Right: &model.InputField{
				LocationID: 3,
				TextTag:    "c1",
				Value:      "alsodifferent",
			},
		},
	}

	result, _, err := MergeInputFields(left, right, nil)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Equal(t, expectedConflicts, err.(MergeConflictError).Conflicts)

	conflictSolution := map[string]MergeSolution{
		"1_a1": {
			Side: RightSide,
			Solution: &model.InputField{
				LocationID: 1,
				TextTag:    "a1",
				Value:      "different",
			},
			Discarded: &model.InputField{
				LocationID: 1,
				TextTag:    "a1",
				Value:      "a1",
			},
		},
		"3_c1": {
			Side: LeftSide,
			Solution: &model.InputField{
				LocationID: 3,
				TextTag:    "c1",
				Value:      "c1",
			},
			Discarded: &model.InputField{
				LocationID: 3,
				TextTag:    "c1",
				Value:      "alsodifferent",
			},
		},
	}
	expectedResult := []*model.InputField{
		nil,
		{
			LocationID: 1,
			TextTag:    "a1",
			Value:      "different",
		},
		{
			LocationID: 1,
			TextTag:    "b1",
			Value:      "b1",
		},
		{
			LocationID: 2,
			TextTag:    "d1",
			Value:      "d1",
		},
		{
			LocationID: 2,
			TextTag:    "d2",
			Value:      "d2",
		},
		{
			LocationID: 3,
			TextTag:    "c1",
			Value:      "c1",
		},
	}

	result, _, err = MergeInputFields(left, right, conflictSolution)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}
