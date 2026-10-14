package merger

import (
	"errors"
	"slices"

	"github.com/AndreasSko/go-library-merger/model"
)

// MergePlaylists merges all playlist tables after locations and tags have been
// merged. It changes the TagMaps of left and right only on success, so it can
// be called again with more conflict solutions. The caller must Close merged.
func MergePlaylists(left, right, merged *model.Database, solutions map[string]MergeSolution) error {
	l, r := model.MakeDatabaseCopy(left), model.MakeDatabaseCopy(right)
	model.AlignPlaylistRepeats(l, r)
	media, mediaIDs, err := MergeIndependentMedia(l.IndependentMedia, r.IndependentMedia)
	if err != nil {
		return err
	}
	UpdateLRIDs(l.PlaylistItemIndependentMediaMap, r.PlaylistItemIndependentMediaMap, "IndependentMediaID", mediaIDs)
	mediaPaths, err := remapIndependentMediaPaths(media)
	if err != nil {
		return err
	}
	// Thumbnails refer to media by filename, which may have changed above.
	// Media that exists in both backups is kept once, under one filename.
	pathByHash := map[string]string{}
	for _, m := range media {
		if m != nil {
			pathByHash[m.UniqueKey()] = m.FilePath
		}
	}
	for _, db := range []*model.Database{l, r} {
		paths := map[string]string{}
		for _, m := range db.IndependentMedia {
			if m != nil {
				path := pathByHash[m.UniqueKey()]
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
	// A solution can be the original item, with a thumbnail filename from
	// before the renaming. Use the renamed item of the chosen side instead.
	itemsByKey := map[MergeSide]map[string]*model.PlaylistItem{LeftSide: {}, RightSide: {}}
	for side, db := range map[MergeSide]*model.Database{LeftSide: l, RightSide: r} {
		for _, item := range db.PlaylistItem {
			if item != nil {
				itemsByKey[side][item.UniqueKey()] = item
			}
		}
	}
	itemSolutions := make(map[string]MergeSolution, len(solutions))
	for key, solution := range solutions {
		if renamed := itemsByKey[solution.Side][key]; renamed != nil {
			if _, ok := solution.Solution.(*model.PlaylistItem); ok {
				solution.Solution = renamed
			}
		}
		itemSolutions[key] = solution
	}
	items, itemIDs, err := MergePlaylistItems(l.PlaylistItem, r.PlaylistItem, itemSolutions)
	if err != nil {
		return err
	}
	UpdateLRIDs(l.TagMap, r.TagMap, "PlaylistItemID", itemIDs)
	UpdateLRIDs(l.PlaylistItemLocationMap, r.PlaylistItemLocationMap, "PlaylistItemID", itemIDs)
	UpdateLRIDs(l.PlaylistItemIndependentMediaMap, r.PlaylistItemIndependentMediaMap, "PlaylistItemID", itemIDs)
	// Conflicts in the maps below show the related item to the user.
	merged.PlaylistItem = items
	locations, _, err := MergePlaylistItemLocationMap(l.PlaylistItemLocationMap, r.PlaylistItemLocationMap, solutions)
	if err != nil {
		return err
	}
	mediaMaps, _, err := MergePlaylistItemIndependentMediaMap(l.PlaylistItemIndependentMediaMap, r.PlaylistItemIndependentMediaMap, solutions)
	if err != nil {
		return err
	}
	// Copy the media into a new directory, so a failed copy keeps the result
	// of an earlier call.
	output := model.MakeDatabaseCopy(merged)
	output.WorkingDir = ""
	if slices.ContainsFunc(media, func(m *model.IndependentMedia) bool { return m != nil }) {
		if err := output.CreateWorkingDirectory(); err != nil {
			return err
		}
		if err := CopyMergedIndependentMedia(media, left, right, output.WorkingDir); err != nil {
			return errors.Join(err, output.Close())
		}
	}
	if err := merged.Close(); err != nil {
		return errors.Join(err, output.Close())
	}
	*merged = *output
	merged.IndependentMedia = media
	merged.PlaylistItem = items
	merged.PlaylistItemLocationMap = locations
	merged.PlaylistItemIndependentMediaMap = mediaMaps
	left.TagMap, right.TagMap = l.TagMap, r.TagMap
	return nil
}
