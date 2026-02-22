package merger

import "github.com/AndreasSko/go-library-merger/model"

// MergeTags tries to merge the left and right slice of Tag. If there is a
// collision, it returns an error asking for specification how it should handle it.
func MergeTags(left []*model.Tag, right []*model.Tag, conflictSolution map[string]MergeSolution) ([]*model.Tag, IDChanges, error) {
	result, changes, err := tryMergeWithConflictSolver(left, right, conflictSolution, solveEqualityMergeConflict)

	return model.Tag{}.MakeSlice(result), changes, err
}
