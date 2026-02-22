package merger

import "github.com/AndreasSko/go-library-merger/model"

// MergePlaylistItemLocationMap tries to merge the left and right slice of PlaylistItemLocationMap. If there is a
// collision, it returns an error asking for specification how it should handle it.
func MergePlaylistItemLocationMap(left []*model.PlaylistItemLocationMap, right []*model.PlaylistItemLocationMap,
	conflictSolution map[string]MergeSolution) ([]*model.PlaylistItemLocationMap, IDChanges, error) {
	result, changes, err := tryMergeWithConflictSolver(left, right, conflictSolution, solveEqualityMergeConflict)

	return model.PlaylistItemLocationMap{}.MakeSlice(result), changes, err
}
