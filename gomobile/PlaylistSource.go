package gomobile

import "github.com/AndreasSko/go-library-merger/model"

// Item conflicts retain the input item IDs, whereas mapping conflicts use the
// merged item IDs. Read the original mapping only for an item conflict; a
// mapping conflict already contains its multimedia type. This avoids deriving
// a video link from an audio item or from an unrelated merged row.
func playlistConflictSourceURL(mdl model.Model, related model.Related, source *model.Database) string {
	switch item := mdl.(type) {
	case *model.PlaylistItem:
		if source == nil {
			return ""
		}
		var mapping *model.PlaylistItemLocationMap
		for _, row := range source.PlaylistItemLocationMap {
			if row == nil || row.PlaylistItemID != item.PlaylistItemID {
				continue
			}
			if mapping != nil {
				return ""
			}
			mapping = row
		}
		if mapping == nil {
			return ""
		}
		return model.PlaylistSourceURL(related.Location, mapping.MajorMultimediaType)
	case *model.PlaylistItemLocationMap:
		return model.PlaylistSourceURL(related.Location, item.MajorMultimediaType)
	default:
		return ""
	}
}
