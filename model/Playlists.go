package model

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type playlistOccurrenceSettings struct {
	label                             string
	start, end, baseDuration          sql.NullInt64
	accuracy, endAction, mediaType    int
	mediaDuration                     int64
	hasLocationMap, hasIndependentMap bool
}

func playlistSettings(db *Database) map[int]playlistOccurrenceSettings {
	result := make(map[int]playlistOccurrenceSettings)
	for _, item := range db.PlaylistItem {
		if item == nil {
			continue
		}
		result[item.PlaylistItemID] = playlistOccurrenceSettings{
			label: item.Label,
			start: item.StartTrimOffsetTicks, end: item.EndTrimOffsetTicks,
			accuracy: item.Accuracy, endAction: item.EndAction,
		}
	}
	for _, row := range db.PlaylistItemLocationMap {
		if row == nil {
			continue
		}
		v := result[row.PlaylistItemID]
		v.hasLocationMap = true
		v.mediaType = row.MajorMultimediaType
		v.baseDuration = row.BaseDurationTicks
		result[row.PlaylistItemID] = v
	}
	for _, row := range db.PlaylistItemIndependentMediaMap {
		if row == nil {
			continue
		}
		v := result[row.PlaylistItemID]
		v.hasIndependentMap = true
		v.mediaDuration = row.DurationTicks
		result[row.PlaylistItemID] = v
	}
	return result
}

// orderedPlaylistSequences includes other media between repetitions when choosing
// anchors. Looking only at one content group would pair AB with the final A in
// ABA and manufacture a contradictory order for an ordinary append.
func orderedPlaylistSequences(db *Database) map[string][]*PlaylistItem {
	type position struct{ order, id int }
	positions := map[int]position{}
	for _, tm := range db.TagMap {
		if tm != nil && tm.PlaylistItemID.Valid {
			positions[int(tm.PlaylistItemID.Int32)] = position{tm.Position, tm.TagMapID}
		}
	}
	sequences := map[string][]*PlaylistItem{}
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
		sequences[tag] = append(sequences[tag], item)
	}
	for _, sequence := range sequences {
		sort.Slice(sequence, func(i, j int) bool {
			a, b := positions[sequence[i].ID()], positions[sequence[j].ID()]
			if a.order != b.order {
				return a.order < b.order
			}
			return a.id < b.id
		})
	}
	return sequences
}

func playlistOccurrenceAnchors(left, right []*PlaylistItem, ls, rs map[int]playlistOccurrenceSettings) (map[*PlaylistItem]*PlaylistItem, error) {
	leftKeys, rightKeys := make([]string, len(left)), make([]string, len(right))
	for i, item := range left {
		leftKeys[i] = item.playlistContentKey()
	}
	for i, item := range right {
		rightKeys[i] = item.playlistContentKey()
	}
	equal := func(i, j int) bool {
		return leftKeys[i] == rightKeys[j] && ls[left[i].ID()] == rs[right[j].ID()]
	}
	// Ordinary additive backups are exact subsequences. Match them in linear
	// time so large repetitive playlists need no quadratic allocation.
	anchors := map[*PlaylistItem]*PlaylistItem{}
	if len(left) <= len(right) {
		i := 0
		for j := range right {
			if i < len(left) && equal(i, j) {
				anchors[left[i]] = right[j]
				i++
			}
		}
		if i == len(left) {
			return anchors, nil
		}
	} else {
		j := 0
		for i := range left {
			if j < len(right) && equal(i, j) {
				anchors[left[i]] = right[j]
				j++
			}
		}
		if j == len(right) {
			return anchors, nil
		}
	}
	// Bound the harder insertion-and-edit case before allocating its matrix.
	// This caps the matrix at about 32 MiB on 64-bit platforms.
	const maxCells = 4_000_000
	if len(left)+1 > maxCells/(len(right)+1) {
		return nil, fmt.Errorf("playlist occurrence matching exceeds supported complexity")
	}
	dp := make([][]int, len(left)+1)
	for i := range dp {
		dp[i] = make([]int, len(right)+1)
	}
	for i := 1; i <= len(left); i++ {
		for j := 1; j <= len(right); j++ {
			dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			if equal(i-1, j-1) {
				dp[i][j] = max(dp[i][j], dp[i-1][j-1]+1)
			}
		}
	}
	anchors = map[*PlaylistItem]*PlaylistItem{}
	for i, j := len(left), len(right); i > 0 && j > 0; {
		if equal(i-1, j-1) && dp[i][j] == dp[i-1][j-1]+1 {
			anchors[left[i-1]] = right[j-1]
			i--
			j--
		} else if dp[i-1][j] >= dp[i][j-1] {
			i--
		} else {
			j--
		}
	}
	return anchors, nil
}

// AlignPlaylistItemOccurrences gives repeated uses of the same media a common
// occurrence identity. Unchanged settings anchor matches in the whole playlist;
// edits in equal-sized spans still become merge conflicts. An insertion beside
// an edited occurrence is rejected when its identity cannot be established.
func AlignPlaylistItemOccurrences(left, right *Database) error {
	ls, rs := playlistSettings(left), playlistSettings(right)
	leftSequences, rightSequences := orderedPlaylistSequences(left), orderedPlaylistSequences(right)
	for tag, rightSequence := range rightSequences {
		leftSequence := leftSequences[tag]
		leftGroups, rightGroups := map[string][]*PlaylistItem{}, map[string][]*PlaylistItem{}
		for _, item := range leftSequence {
			leftGroups[item.playlistContentKey()] = append(leftGroups[item.playlistContentKey()], item)
		}
		for _, item := range rightSequence {
			rightGroups[item.playlistContentKey()] = append(rightGroups[item.playlistContentKey()], item)
		}
		needsAnchors := false
		for key, rg := range rightGroups {
			if lg := leftGroups[key]; len(lg) > 0 && len(lg) != len(rg) {
				needsAnchors = true
			}
		}
		matched := map[*PlaylistItem]*PlaylistItem{}
		if needsAnchors {
			var err error
			matched, err = playlistOccurrenceAnchors(leftSequence, rightSequence, ls, rs)
			if err != nil {
				return err
			}
		}
		for key, rg := range rightGroups {
			lg := leftGroups[key]
			if len(lg) == 0 {
				continue
			}
			if len(lg) == len(rg) {
				for i := range rg {
					rg[i].occurrence = lg[i].occurrence
				}
				continue
			}
			type pair struct{ left, right int }
			rightIndices := map[*PlaylistItem]int{}
			for i, item := range rg {
				rightIndices[item] = i
			}
			anchors := []pair{{-1, -1}}
			for i, item := range lg {
				if rightItem := matched[item]; rightItem != nil {
					anchors = append(anchors, pair{i, rightIndices[rightItem]})
				}
			}
			anchors = append(anchors, pair{len(lg), len(rg)})
			next := 0
			for _, item := range lg {
				next = max(next, item.occurrence+1)
			}
			paired := len(anchors) - 2
			for a := 1; a < len(anchors); a++ {
				prev, curr := anchors[a-1], anchors[a]
				ln, rn := curr.left-prev.left-1, curr.right-prev.right-1
				if ln > 0 && rn > 0 && ln != rn {
					return fmt.Errorf("ambiguous repeated playlist occurrence insertion and edit")
				}
				paired += min(ln, rn)
				for offset := 1; offset <= rn; offset++ {
					rightIndex := prev.right + offset
					if offset <= ln {
						rg[rightIndex].occurrence = lg[prev.left+offset].occurrence
					} else {
						rg[rightIndex].occurrence = next
						next++
					}
				}
				if curr.left < len(lg) {
					rg[curr.right].occurrence = lg[curr.left].occurrence
				}
			}
			// Leaving unmatched occurrences on both sides can disguise an edit as two
			// additions, even if an arbitrary LCS tie made each individual gap empty on
			// one side. Reject rather than silently growing the playlist.
			if paired != min(len(lg), len(rg)) {
				return fmt.Errorf("ambiguous repeated playlist occurrence insertion and edit")
			}
		}
	}
	return nil
}

// CleanupMedia releases the media directory owned by this imported database.
// Database copies share the directory; clean up only after all copies are done.
func (db *Database) CleanupMedia() error {
	if db.IndependentMediaDir == "" {
		return nil
	}
	err := os.RemoveAll(db.IndependentMediaDir)
	if err == nil {
		db.IndependentMediaDir = ""
	}
	return err
}

// PlaylistThumbnail reads a thumbnail from the imported database that owns it.
// Item IDs refer to that input, before merge-time media filename remapping.
func (db *Database) PlaylistThumbnail(playlistItemID int) ([]byte, error) {
	if playlistItemID < 0 || playlistItemID >= len(db.PlaylistItem) || db.PlaylistItem[playlistItemID] == nil {
		return nil, fmt.Errorf("invalid PlaylistItemID %d", playlistItemID)
	}
	item := db.PlaylistItem[playlistItemID]
	if !item.ThumbnailFilePath.Valid {
		return append([]byte(nil), defaultThumbnailFile...), nil
	}
	if err := validateBackupFilename(item.ThumbnailFilePath.String); err != nil {
		return nil, err
	}
	if item.thumbnail == nil || db.IndependentMediaDir == "" {
		return nil, fmt.Errorf("missing playlist thumbnail %q", item.ThumbnailFilePath.String)
	}
	return os.ReadFile(filepath.Join(db.IndependentMediaDir, item.ThumbnailFilePath.String))
}

func (db *Database) removePlaylists() {
	for i, tag := range db.Tag {
		if tag != nil && tag.TagType == 2 {
			db.Tag[i] = nil
			db.ContainsPlaylists = true
		}
	}
	for i, tm := range db.TagMap {
		if tm != nil && tm.PlaylistItemID.Valid {
			db.TagMap[i] = nil
			db.ContainsPlaylists = true
		}
	}
	db.PlaylistItem = nil
	db.PlaylistItemLocationMap = nil
	db.PlaylistItemIndependentMediaMap = nil
	db.IndependentMedia = nil
}

// LinkPlaylistItems validates playlist relationships and rebuilds content keys.
// Repeated uses of the same media are distinguished by their occurrence in the
// playlist, so deduplication across backups does not remove intentional repeats.
func (db *Database) LinkPlaylistItems() error {
	for _, item := range db.PlaylistItem {
		if item != nil {
			item.tag = nil
			item.location = nil
			item.independentMedia = nil
			item.thumbnail = nil
			item.occurrence = 0
		}
	}
	itemAt := func(id int) (*PlaylistItem, error) {
		if id < 0 || id >= len(db.PlaylistItem) || db.PlaylistItem[id] == nil {
			return nil, fmt.Errorf("invalid PlaylistItemID %d", id)
		}
		return db.PlaylistItem[id], nil
	}
	ordered := make([]*TagMap, 0)
	for _, tm := range db.TagMap {
		if tm == nil || !tm.PlaylistItemID.Valid {
			continue
		}
		item, err := itemAt(int(tm.PlaylistItemID.Int32))
		if err != nil {
			return err
		}
		if tm.TagID < 0 || tm.TagID >= len(db.Tag) || db.Tag[tm.TagID] == nil || db.Tag[tm.TagID].TagType != 2 {
			return fmt.Errorf("invalid playlist TagID %d", tm.TagID)
		}
		if item.tag != nil {
			return fmt.Errorf("playlist item %d belongs to multiple playlist entries", item.ID())
		}
		item.tag = db.Tag[tm.TagID]
		ordered = append(ordered, tm)
	}
	for _, m := range db.PlaylistItemLocationMap {
		if m == nil {
			continue
		}
		item, err := itemAt(m.PlaylistItemID)
		if err != nil {
			return err
		}
		if m.LocationID < 0 || m.LocationID >= len(db.Location) || db.Location[m.LocationID] == nil {
			return fmt.Errorf("invalid playlist LocationID %d", m.LocationID)
		}
		if item.location != nil {
			return fmt.Errorf("multiple locations for playlist item %d are not supported", item.ID())
		}
		item.location = db.Location[m.LocationID]
	}
	for _, m := range db.PlaylistItemIndependentMediaMap {
		if m == nil {
			continue
		}
		item, err := itemAt(m.PlaylistItemID)
		if err != nil {
			return err
		}
		if m.IndependentMediaID < 0 || m.IndependentMediaID >= len(db.IndependentMedia) || db.IndependentMedia[m.IndependentMediaID] == nil {
			return fmt.Errorf("invalid IndependentMediaID %d", m.IndependentMediaID)
		}
		if item.independentMedia != nil {
			return fmt.Errorf("multiple media for playlist item %d are not supported", item.ID())
		}
		item.independentMedia = db.IndependentMedia[m.IndependentMediaID]
	}
	mediaByPath := map[string]*IndependentMedia{}
	for _, media := range db.IndependentMedia {
		if media != nil {
			mediaByPath[media.FilePath] = media
		}
	}
	for _, item := range db.PlaylistItem {
		if item == nil {
			continue
		}
		if item.ThumbnailFilePath.Valid {
			item.thumbnail = mediaByPath[item.ThumbnailFilePath.String]
			if item.thumbnail == nil {
				return fmt.Errorf("invalid playlist thumbnail %q for item %d", item.ThumbnailFilePath.String, item.ID())
			}
		}
		if item.tag == nil {
			return fmt.Errorf("playlist item %d has no related Tag", item.ID())
		}
		if (item.location == nil) == (item.independentMedia == nil) {
			return fmt.Errorf("playlist item %d must have exactly one media or location", item.ID())
		}
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].TagID != ordered[j].TagID {
			return ordered[i].TagID < ordered[j].TagID
		}
		if ordered[i].Position != ordered[j].Position {
			return ordered[i].Position < ordered[j].Position
		}
		return ordered[i].TagMapID < ordered[j].TagMapID
	})
	occurrences := map[string]int{}
	for _, tm := range ordered {
		item := db.PlaylistItem[tm.PlaylistItemID.Int32]
		key := item.UniqueKey()
		item.occurrence = occurrences[key]
		occurrences[key]++
	}
	return nil
}

func copyTag(v *Tag) *Tag {
	if v == nil {
		return nil
	}
	return MakeModelCopy(v).(*Tag)
}
func copyMedia(v *IndependentMedia) *IndependentMedia {
	if v == nil {
		return nil
	}
	return MakeModelCopy(v).(*IndependentMedia)
}
func copyLocation(v *Location) *Location {
	if v == nil {
		return nil
	}
	return MakeModelCopy(v).(*Location)
}

// Database copies must preserve the sharing between an item and its related
// rows. Pre-merge migrations update those rows in place.
func (db *Database) relinkCopiedPlaylistRelations() {
	locations := map[int]*Location{}
	tags := map[int]*Tag{}
	media := map[int]*IndependentMedia{}
	for _, v := range db.Location {
		if v != nil {
			locations[v.ID()] = v
		}
	}
	for _, v := range db.Tag {
		if v != nil {
			tags[v.ID()] = v
		}
	}
	for _, v := range db.IndependentMedia {
		if v != nil {
			media[v.ID()] = v
		}
	}
	for _, item := range db.PlaylistItem {
		if item == nil {
			continue
		}
		if item.location != nil {
			if v := locations[item.location.ID()]; v != nil {
				item.location = v
			}
		}
		if item.tag != nil {
			if v := tags[item.tag.ID()]; v != nil {
				item.tag = v
			}
		}
		if item.independentMedia != nil {
			if v := media[item.independentMedia.ID()]; v != nil {
				item.independentMedia = v
			}
		}
		if item.thumbnail != nil {
			if v := media[item.thumbnail.ID()]; v != nil {
				item.thumbnail = v
			}
		}
	}
}
