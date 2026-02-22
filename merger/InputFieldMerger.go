package merger

import "github.com/AndreasSko/go-library-merger/model"

// MergeInputFields tries to merge the left and right slice of InputField. If there is a
// collision, it returns an error asking for specification how it should handle it.
func MergeInputFields(left []*model.InputField, right []*model.InputField, conflictSolution map[string]MergeSolution) ([]*model.InputField, IDChanges, error) {
	result, changes, err := tryMergeWithConflictSolver(left, right, conflictSolution, solveEqualityMergeConflict)
	// As InputField does not have a proper ID to sort by, we additionally
	// sort it by UniqueKey to have a consistent result
	model.SortByUniqueKey(&result)

	return model.InputField{}.MakeSlice(result), changes, err
}
