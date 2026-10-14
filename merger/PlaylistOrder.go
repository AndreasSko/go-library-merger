package merger

import (
	"sort"

	"github.com/AndreasSko/go-library-merger/model"
)

// playlistOrder combines the order of the playlist items of both sides.
// Absolute positions are local to each backup and cannot be compared after an
// insertion, so it uses which item follows which. If the sides contradict each
// other, for example because a playlist was reordered on one device, the
// order of the left side wins. Items of the right side still keep their place
// relative to their other neighbors.
func playlistOrder(left, right []*model.TagMap) map[string]int {
	nodes := map[string]*model.TagMap{}
	edges := map[string]map[string]bool{}
	indegree := map[string]int{}
	var reaches func(from, to string, seen map[string]bool) bool
	reaches = func(from, to string, seen map[string]bool) bool {
		if from == to {
			return true
		}
		seen[from] = true
		for next := range edges[from] {
			if !seen[next] && reaches(next, to, seen) {
				return true
			}
		}
		return false
	}
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
				// Skip an order that contradicts the one already known.
				if edges[a][b] || reaches(b, a, map[string]bool{}) {
					continue
				}
				if edges[a] == nil {
					edges[a] = map[string]bool{}
				}
				edges[a][b] = true
				indegree[b]++
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
	return result
}
