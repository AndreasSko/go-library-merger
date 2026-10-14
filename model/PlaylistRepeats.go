package model

import (
	"sort"
)

// A playlist can contain the same media several times, for example one video
// with different trims. Items are matched across backups by their media and
// their repeatIndex, so both backups must number those repeats the same way:
//
//	left:  [video 0:00, video 0:20]
//	right: [video 0:00, video 0:10, video 0:20]
//
// Numbering by position would match 0:20 with 0:10. AlignPlaylistRepeats
// first matches items whose media and settings are unchanged (0:00 and 0:20),
// and only then gives the remaining items their numbers (0:10 is new).

// itemSettingsByID returns the settings of each item and its location or
// media row, joined into one string so they can be compared.
func itemSettingsByID(db *Database) map[int]string {
	result := map[int]string{}
	add := func(id int, entry Model) {
		for _, setting := range playlistSettings(entry) {
			result[id] += setting.value + "\x00"
		}
	}
	for _, item := range db.PlaylistItem {
		if item != nil {
			add(item.PlaylistItemID, item)
		}
	}
	for _, row := range db.PlaylistItemLocationMap {
		if row != nil {
			add(row.PlaylistItemID, row)
		}
	}
	for _, row := range db.PlaylistItemIndependentMediaMap {
		if row != nil {
			add(row.PlaylistItemID, row)
		}
	}
	return result
}

// itemsByPlaylist returns the items of each playlist in playlist order.
func itemsByPlaylist(db *Database) map[string][]*PlaylistItem {
	type position struct{ order, id int }
	positions := map[int]position{}
	for _, tm := range db.TagMap {
		if tm != nil && tm.PlaylistItemID.Valid {
			positions[int(tm.PlaylistItemID.Int32)] = position{tm.Position, tm.TagMapID}
		}
	}
	playlists := map[string][]*PlaylistItem{}
	for i, item := range db.PlaylistItem {
		if item == nil {
			continue
		}
		if _, ok := positions[item.ID()]; !ok {
			positions[item.ID()] = position{i, i}
		}
		tag := ""
		if item.tag != nil {
			tag = item.tag.UniqueKey()
		}
		playlists[tag] = append(playlists[tag], item)
	}
	for _, items := range playlists {
		sort.Slice(items, func(i, j int) bool {
			a, b := positions[items[i].ID()], positions[items[j].ID()]
			if a.order != b.order {
				return a.order < b.order
			}
			return a.id < b.id
		})
	}
	return playlists
}

// matchUnchangedItems pairs items with the same media and settings, keeping
// the playlist order on both sides (a longest common subsequence).
func matchUnchangedItems(left, right []*PlaylistItem, leftSettings, rightSettings map[int]string) map[*PlaylistItem]*PlaylistItem {
	unchanged := func(i, j int) bool {
		return left[i].mediaKey() == right[j].mediaKey() && leftSettings[left[i].ID()] == rightSettings[right[j].ID()]
	}
	// common[i][j] is the number of matches between left[i:] and right[j:].
	common := make([][]int, len(left)+1)
	for i := range common {
		common[i] = make([]int, len(right)+1)
	}
	for i := len(left) - 1; i >= 0; i-- {
		for j := len(right) - 1; j >= 0; j-- {
			if unchanged(i, j) {
				common[i][j] = common[i+1][j+1] + 1
			} else {
				common[i][j] = max(common[i+1][j], common[i][j+1])
			}
		}
	}
	matches := map[*PlaylistItem]*PlaylistItem{}
	for i, j := 0, 0; i < len(left) && j < len(right); {
		switch {
		case unchanged(i, j):
			matches[left[i]] = right[j]
			i++
			j++
		case common[i+1][j] >= common[i][j+1]:
			i++
		default:
			j++
		}
	}
	return matches
}

// AlignPlaylistRepeats renumbers the repeated media in right so they match
// the numbers in left. Between two matched items, the same number of
// unmatched items on both sides are paired by position, so their changed
// settings become a merge conflict. If the numbers differ, the right items
// count as new, so the merged playlist keeps both versions.
func AlignPlaylistRepeats(left, right *Database) {
	leftSettings, rightSettings := itemSettingsByID(left), itemSettingsByID(right)
	leftPlaylists := itemsByPlaylist(left)
	for tag, rightItems := range itemsByPlaylist(right) {
		leftItems := leftPlaylists[tag]
		leftByMedia, rightByMedia := map[string][]*PlaylistItem{}, map[string][]*PlaylistItem{}
		for _, item := range leftItems {
			leftByMedia[item.mediaKey()] = append(leftByMedia[item.mediaKey()], item)
		}
		for _, item := range rightItems {
			rightByMedia[item.mediaKey()] = append(rightByMedia[item.mediaKey()], item)
		}

		var matches map[*PlaylistItem]*PlaylistItem
		for media, rightRepeats := range rightByMedia {
			leftRepeats := leftByMedia[media]
			// With the same count on both sides, numbering by position already matches.
			if len(leftRepeats) == 0 || len(leftRepeats) == len(rightRepeats) {
				continue
			}
			if matches == nil {
				matches = matchUnchangedItems(leftItems, rightItems, leftSettings, rightSettings)
			}
			alignRepeats(leftRepeats, rightRepeats, matches)
		}
	}
}

func alignRepeats(left, right []*PlaylistItem, matches map[*PlaylistItem]*PlaylistItem) {
	rightIndex := map[*PlaylistItem]int{}
	for i, item := range right {
		rightIndex[item] = i
	}
	type pair struct{ left, right int }
	pairs := []pair{{-1, -1}}
	nextRepeatIndex := 0
	for i, item := range left {
		if match := matches[item]; match != nil {
			pairs = append(pairs, pair{i, rightIndex[match]})
		}
		nextRepeatIndex = max(nextRepeatIndex, item.repeatIndex+1)
	}
	pairs = append(pairs, pair{len(left), len(right)})

	for p := 1; p < len(pairs); p++ {
		prev, curr := pairs[p-1], pairs[p]
		leftGap, rightGap := curr.left-prev.left-1, curr.right-prev.right-1
		for offset := 1; offset <= rightGap; offset++ {
			item := right[prev.right+offset]
			if leftGap == rightGap {
				item.repeatIndex = left[prev.left+offset].repeatIndex
			} else {
				item.repeatIndex = nextRepeatIndex
				nextRepeatIndex++
			}
		}
		if curr.left < len(left) {
			right[curr.right].repeatIndex = left[curr.left].repeatIndex
		}
	}
}
