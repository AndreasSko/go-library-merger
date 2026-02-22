package merger

import (
	"sort"

	"github.com/AndreasSko/go-library-merger/model"
)

// MergeTagMaps merges a left and right slice of TagMap. It automatically
// removes redundant entries and also makes sure that the position-order
// stays similar.
func MergeTagMaps(left []*model.TagMap, right []*model.TagMap, conflictSolution map[string]MergeSolution) ([]*model.TagMap, IDChanges, error) {
	if len(left)+len(right) == 0 {
		return []*model.TagMap{nil}, IDChanges{}, nil
	}

	// map[TagID]map[UniqueKey]TagMap
	tags := make(map[int]map[string]*model.TagMap, len(left)+len(right))

	// Per TagID add TagMap entries to map with UniqueKey as the key,
	// automatically filtering duplicate entries
	for _, side := range [][]*model.TagMap{left, right} {
		for _, tm := range side {
			if tm == nil {
				continue
			}

			if _, ok := tags[tm.TagID]; !ok {
				tags[tm.TagID] = map[string]*model.TagMap{}
			}

			tags[tm.TagID][tm.UniqueKey()] = tm
		}
	}

	result := make([]*model.TagMap, len(left)+len(right))

	// Go through map in sorted order so we have deterministic results
	sortedTagIDs := make([]int, len(tags))
	i := 0
	for key := range tags {
		sortedTagIDs[i] = key
		i++
	}
	sort.Ints(sortedTagIDs)

	// For each TagID add all connected TagMaps to result
	i = 1
	for _, id := range sortedTagIDs {
		tagMapSet := tags[id]
		sortedTagMap := make([]*model.TagMap, len(tagMapSet))
		j := 0
		for _, tm := range tagMapSet {
			sortedTagMap[j] = tm
			j++
		}

		sort.SliceStable(sortedTagMap, func(i, j int) bool {
			// If equal position, sort by TagID
			if sortedTagMap[i].Position == sortedTagMap[j].Position {
				return sortedTagMap[i].TagMapID < sortedTagMap[j].TagMapID
			}
			return sortedTagMap[i].Position < sortedTagMap[j].Position
		})

		j = 0
		for j, tm := range sortedTagMap {
			result[i] = model.MakeModelCopy(tm).(*model.TagMap)
			result[i].SetID(i)
			// Position is defined per Tag(!), not PlaylistItemID/LocationID/NoteID
			result[i].Position = j
			j++
			i++
		}
	}

	return result[:i], IDChanges{}, nil
}
