package merger

import "github.com/AndreasSko/go-library-merger/model"

// MergeNotes tries to merge the left and right slice of Note. If there is a
// collision, it returns an error asking for specification how it should handle it.
func MergeNotes(left []*model.Note, right []*model.Note, conflictSolution map[string]MergeSolution) ([]*model.Note, IDChanges, error) {
	result, changes, err := tryMergeWithConflictSolver(left, right, conflictSolution, solveEqualityMergeConflict)

	return model.Note{}.MakeSlice(result), changes, err
}
