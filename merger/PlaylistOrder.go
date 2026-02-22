package merger

import (
	"fmt"
	"sort"

	"github.com/AndreasSko/go-library-merger/model"
)

// playlistOrder combines relative ordering constraints from both inputs.
// Absolute positions are local to each backup and cannot be compared after an
// insertion. Contradictory reorders are rejected instead of silently reordered.
func playlistOrder(left, right []*model.TagMap) (map[string]int, error) {
	nodes := map[string]*model.TagMap{}
	edges := map[string]map[string]bool{}
	indegree := map[string]int{}
	for _, side := range [][]*model.TagMap{left, right} {
		groups := map[int][]*model.TagMap{}
		for _, tm := range side {
			if tm == nil || !tm.PlaylistItemID.Valid {
				continue
			}
			key := tm.UniqueKey()
			nodes[key] = tm
			groups[tm.TagID] = append(groups[tm.TagID], tm)
		}
		for _, seq := range groups {
			sort.Slice(seq, func(i, j int) bool {
				if seq[i].Position != seq[j].Position {
					return seq[i].Position < seq[j].Position
				}
				return seq[i].TagMapID < seq[j].TagMapID
			})
			for i := 1; i < len(seq); i++ {
				a, b := seq[i-1].UniqueKey(), seq[i].UniqueKey()
				if a == b {
					continue
				}
				if edges[a] == nil {
					edges[a] = map[string]bool{}
				}
				if !edges[a][b] {
					edges[a][b] = true
					indegree[b]++
				}
			}
		}
	}
	ready := []string{}
	for key := range nodes {
		if indegree[key] == 0 {
			ready = append(ready, key)
		}
	}
	result := map[string]int{}
	positions := map[int]int{}
	for len(ready) > 0 {
		sort.Strings(ready)
		key := ready[0]
		ready = ready[1:]
		tag := nodes[key].TagID
		result[key] = positions[tag]
		positions[tag]++
		for next := range edges[key] {
			indegree[next]--
			if indegree[next] == 0 {
				ready = append(ready, next)
			}
		}
	}
	if len(result) != len(nodes) {
		return nil, fmt.Errorf("conflicting playlist order: the backups contain contradictory reorders")
	}
	return result, nil
}
