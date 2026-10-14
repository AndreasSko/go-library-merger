package merger

import "github.com/AndreasSko/go-library-merger/model"

// MergePlaylistItems tries to merge the left and right slice of PlaylistItem. If there is a
// collision, it returns an error asking for specification how it should handle it.
func MergePlaylistItems(left []*model.PlaylistItem, right []*model.PlaylistItem, conflictSolution map[string]MergeSolution) ([]*model.PlaylistItem, IDChanges, error) {
	result, changes, err := tryMergeWithConflictSolver(left, right, conflictSolution, solveEqualityMergeConflict)

	return model.PlaylistItem{}.MakeSlice(result), changes, err
}
