package merger

import "github.com/AndreasSko/go-library-merger/model"

// MergePlaylistItemIndependentMediaMap tries to merge the left and right slice of PlaylistItemIndependentMediaMap. If there is a
// collision, it returns an error asking for specification how it should handle it.
func MergePlaylistItemIndependentMediaMap(left []*model.PlaylistItemIndependentMediaMap, right []*model.PlaylistItemIndependentMediaMap,
	conflictSolution map[string]MergeSolution) ([]*model.PlaylistItemIndependentMediaMap, IDChanges, error) {
	result, changes, err := tryMergeWithConflictSolver(left, right, conflictSolution, solveEqualityMergeConflict)

	return model.PlaylistItemIndependentMediaMap{}.MakeSlice(result), changes, err
}
