package merger

import (
	"github.com/AndreasSko/go-library-merger/model"
)

// MergeBookmarks tries to merge the left and right slices of Bookmarks. If there is a
// collision, it returns an error asking for specification how it should handle it.
func MergeBookmarks(left []*model.Bookmark, right []*model.Bookmark, conflictSolution map[string]MergeSolution) ([]*model.Bookmark, IDChanges, error) {
	result, changes, err := tryMergeWithConflictSolver(left, right, conflictSolution, solveEqualityMergeConflict)

	return model.Bookmark{}.MakeSlice(result), changes, err
}
