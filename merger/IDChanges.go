package merger

import (
	"github.com/AndreasSko/go-library-merger/model"
)

// IDChanges represents the changed ids of two slices of a model type
// after a merge has happened, so dependent objects can be updated
// accordingly. So if the ID of an object of the left slice
// changed from id 5 to 20, it will be represented as: {5: 20}.
type IDChanges struct {
	Left  map[int]int
	Right map[int]int
}

// UpdateLRIDs updates a given ID (named by IDName) on the left and right
// slices of *model.Model according to the given IDChanges.
func UpdateLRIDs(left interface{}, right interface{}, IDName string, changes IDChanges) {
	for _, mSide := range []MergeSide{LeftSide, RightSide} {
		var side interface{}
		var chges map[int]int
		if mSide == LeftSide {
			side = left
			chges = changes.Left
		} else {
			side = right
			chges = changes.Right
		}

		if side == nil {
			continue
		}

		model.UpdateIDs(side, IDName, chges)
	}
}
