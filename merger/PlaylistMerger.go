package merger

import (
	"errors"
	"os"

	"github.com/AndreasSko/go-library-merger/model"
)

// MergePlaylists merges all playlist tables after locations and tags have been
// merged. It applies item ID changes to input TagMaps only on success, making
// conflict retries safe. The caller must eventually clean up merged media.
func MergePlaylists(left, right, merged *model.Database, solutions map[string]MergeSolution) error {
	l, r := model.MakeDatabaseCopy(left), model.MakeDatabaseCopy(right)
	if err := model.AlignPlaylistItemOccurrences(l, r); err != nil {
		return err
	}
	media, mediaIDs, err := MergeIndependentMedia(l.IndependentMedia, r.IndependentMedia)
	if err != nil {
		return err
	}
	UpdateLRIDs(l.PlaylistItemIndependentMediaMap, r.PlaylistItemIndependentMediaMap, "IndependentMediaID", mediaIDs)
	mediaPaths, err := remapIndependentMediaPaths(media)
	if err != nil {
		return err
	}
	byHash := map[string]string{}
	for _, m := range media {
		if m != nil {
			byHash[m.UniqueKey()] = m.FilePath
		}
	}
	for _, db := range []*model.Database{l, r} {
		paths := map[string]string{}
		for _, m := range db.IndependentMedia {
			if m != nil {
				path := byHash[m.UniqueKey()]
				if remapped, ok := mediaPaths[m.UniqueKey()+"\x00"+m.FilePath]; ok {
					path = remapped
				}
				paths[m.FilePath] = path
			}
		}
		for _, item := range db.PlaylistItem {
			if item != nil && item.ThumbnailFilePath.Valid {
				if path, ok := paths[item.ThumbnailFilePath.String]; ok {
					item.ThumbnailFilePath.String = path
				}
			}
		}
	}
	items, itemIDs, err := MergePlaylistItems(l.PlaylistItem, r.PlaylistItem, solutions)
	if err != nil {
		return err
	}
	UpdateLRIDs(l.TagMap, r.TagMap, "PlaylistItemID", itemIDs)
	UpdateLRIDs(l.PlaylistItemLocationMap, r.PlaylistItemLocationMap, "PlaylistItemID", itemIDs)
	UpdateLRIDs(l.PlaylistItemIndependentMediaMap, r.PlaylistItemIndependentMediaMap, "PlaylistItemID", itemIDs)
	// Keep item context available to mobile clients resolving mapping conflicts.
	// Export remains blocked until all playlist stages succeed.
	merged.PlaylistItem = items
	locations, _, err := MergePlaylistItemLocationMap(l.PlaylistItemLocationMap, r.PlaylistItemLocationMap, solutions)
	if err != nil {
		return err
	}
	mediaMaps, _, err := MergePlaylistItemIndependentMediaMap(l.PlaylistItemIndependentMediaMap, r.PlaylistItemIndependentMediaMap, solutions)
	if err != nil {
		return err
	}
	// Detect contradictory sequences before copying potentially large media.
	if _, err := playlistOrder(l.TagMap, r.TagMap); err != nil {
		return err
	}
	dir := ""
	hasMedia := false
	for _, m := range media {
		hasMedia = hasMedia || m != nil
	}
	if hasMedia {
		dir, err = os.MkdirTemp(merged.TempDir, "library-merger-media")
		if err != nil {
			return err
		}
		if err = CopyMergedIndependentMedia(media, left, right, dir); err != nil {
			return errors.Join(err, os.RemoveAll(dir))
		}
	}
	if err = merged.CleanupMedia(); err != nil {
		if dir != "" {
			return errors.Join(err, os.RemoveAll(dir))
		}
		return err
	}
	merged.IndependentMediaDir = dir
	merged.IndependentMedia = media
	merged.PlaylistItem = items
	merged.PlaylistItemLocationMap = locations
	merged.PlaylistItemIndependentMediaMap = mediaMaps
	left.TagMap, right.TagMap = l.TagMap, r.TagMap
	return nil
}
