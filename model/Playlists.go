package model

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

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
	if db.WorkingDir == "" {
		return nil, fmt.Errorf("missing playlist thumbnail %q", item.ThumbnailFilePath.String)
	}
	return os.ReadFile(filepath.Join(db.WorkingDir, item.ThumbnailFilePath.String))
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

// LinkPlaylistItems validates playlist relationships and links each item to
// its playlist and media. Repeats of the same media in one playlist get a
// repeatIndex by position, so merging does not remove intentional repeats.
func (db *Database) LinkPlaylistItems() error {
	for _, item := range db.PlaylistItem {
		if item != nil {
			item.tag = nil
			item.location = nil
			item.independentMedia = nil
			item.repeatIndex = 0
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
	mediaPaths := map[string]bool{}
	for _, media := range db.IndependentMedia {
		if media != nil {
			mediaPaths[media.FilePath] = true
		}
	}
	for _, item := range db.PlaylistItem {
		if item == nil {
			continue
		}
		if item.ThumbnailFilePath.Valid && !mediaPaths[item.ThumbnailFilePath.String] {
			return fmt.Errorf("invalid playlist thumbnail %q for item %d", item.ThumbnailFilePath.String, item.ID())
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
	repeats := map[string]int{}
	for _, tm := range ordered {
		item := db.PlaylistItem[tm.PlaylistItemID.Int32]
		key := item.mediaKey()
		item.repeatIndex = repeats[key]
		repeats[key]++
	}
	return nil
}

// relinkCopiedPlaylistRelations points the items of a copied database to the
// copied rows, because merging changes those rows in place.
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
	}
}
