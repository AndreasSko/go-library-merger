package merger

import "github.com/AndreasSko/go-library-merger/model"

// MergePlaylistItems tries to merge the left and right slice of PlaylistItem. If there is a
// collision, it returns an error asking for specification how it should handle it.
func MergePlaylistItems(left []*model.PlaylistItem, right []*model.PlaylistItem, conflictSolution map[string]MergeSolution) ([]*model.PlaylistItem, IDChanges, error) {
	// Work on copies: original inputs remain available across conflict retries.
	copyItems := func(items []*model.PlaylistItem) []*model.PlaylistItem {
		result := make([]*model.PlaylistItem, len(items))
		for i, item := range items {
			if item != nil {
				result[i] = model.MakeModelCopy(item).(*model.PlaylistItem)
			}
		}
		return result
	}
	left, right = copyItems(left), copyItems(right)
	byKey := map[string]*model.PlaylistItem{}
	for _, item := range left {
		if item != nil {
			byKey[item.UniqueKey()] = item
		}
	}
	for _, item := range right {
		if item != nil {
			if other := byKey[item.UniqueKey()]; other != nil {
				model.NormalizePlaylistThumbnails(other, item)
			}
		}
	}
	// Explicit callers may supply the original chosen model rather than the
	// normalized conflict model. Keep their meaningful settings while applying
	// the same cosmetic policy, without mutating the supplied solution.
	resolved := make(map[string]MergeSolution, len(conflictSolution))
	for key, solution := range conflictSolution {
		if item, ok := solution.Solution.(*model.PlaylistItem); ok && byKey[key] != nil {
			copy := model.MakeModelCopy(item).(*model.PlaylistItem)
			model.CopyPlaylistThumbnail(copy, byKey[key])
			solution.Solution = copy
		}
		resolved[key] = solution
	}
	result, changes, err := tryMergeWithConflictSolver(left, right, resolved, solveEqualityMergeConflict)

	return model.PlaylistItem{}.MakeSlice(result), changes, err
}
