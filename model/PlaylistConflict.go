package model

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// PlaylistChangedFields lists meaningful settings that differ, in display order.
// IDs identify relationships, and thumbnails are cosmetic merge metadata.
func PlaylistChangedFields(left, right Model) []string {
	fields := []string{}
	add := func(name string, different bool) {
		if different {
			fields = append(fields, name)
		}
	}
	switch l := left.(type) {
	case *PlaylistItem:
		if r, ok := right.(*PlaylistItem); ok {
			add("label", l.Label != r.Label)
			add("startTrimOffsetTicks", l.StartTrimOffsetTicks != r.StartTrimOffsetTicks)
			add("endTrimOffsetTicks", l.EndTrimOffsetTicks != r.EndTrimOffsetTicks)
			add("accuracy", l.Accuracy != r.Accuracy)
			add("endAction", l.EndAction != r.EndAction)
		}
	case *PlaylistItemLocationMap:
		if r, ok := right.(*PlaylistItemLocationMap); ok {
			add("baseDurationTicks", l.BaseDurationTicks != r.BaseDurationTicks)
			add("majorMultimediaType", l.MajorMultimediaType != r.MajorMultimediaType)
		}
	case *PlaylistItemIndependentMediaMap:
		if r, ok := right.(*PlaylistItemIndependentMediaMap); ok {
			add("durationTicks", l.DurationTicks != r.DurationTicks)
		}
	}
	return fields
}

// PlaylistChangedFieldText formats only the settings selected by the shared
// comparison. Callers retain the full original model as conflict context.
func PlaylistChangedFieldText(entry Model, fields []string) string {
	names := map[string]string{
		"label": "Label", "startTrimOffsetTicks": "StartTrimOffsetTicks",
		"endTrimOffsetTicks": "EndTrimOffsetTicks", "accuracy": "Accuracy",
		"endAction": "EndAction", "baseDurationTicks": "BaseDurationTicks",
		"majorMultimediaType": "MajorMultimediaType", "durationTicks": "DurationTicks",
	}
	lines := make([]string, 0, len(fields))
	for _, field := range fields {
		name, known := names[field]
		if !known {
			lines = append(lines, "Unrecognized setting: "+field)
			continue
		}
		modelValue := reflect.ValueOf(entry)
		if modelValue.Kind() != reflect.Pointer || modelValue.IsNil() || !modelValue.Elem().FieldByName(name).IsValid() {
			lines = append(lines, "Unrecognized setting: "+field)
			continue
		}
		value := modelValue.Elem().FieldByName(name).Interface()
		text := fmt.Sprint(value)
		if nullable, ok := value.(sql.NullInt64); ok {
			text = "Not set"
			if nullable.Valid {
				text = fmt.Sprint(nullable.Int64)
			}
		}
		lines = append(lines, name+": "+text)
	}
	return strings.Join(lines, "\n")
}

// NormalizePlaylistThumbnails makes a cosmetic choice independent of the
// meaningful version selected later. Imported present thumbnails have validated
// payloads; prefer one over absence, then the lowest canonical content identity.
// A path tie-breaker also makes equal-content aliases deterministic.
func NormalizePlaylistThumbnails(left, right *PlaylistItem) {
	usable := func(item *PlaylistItem) bool {
		return item.ThumbnailFilePath.Valid && item.ThumbnailFilePath.String != ""
	}
	if !usable(left) && !usable(right) {
		left.ThumbnailFilePath, right.ThumbnailFilePath = sql.NullString{}, sql.NullString{}
		left.thumbnail, right.thumbnail = nil, nil
		return
	}
	chosen := left
	if (!usable(left) && usable(right)) || (usable(left) && usable(right) &&
		(right.thumbnailContentKey() < left.thumbnailContentKey() ||
			(right.thumbnailContentKey() == left.thumbnailContentKey() && right.ThumbnailFilePath.String < left.ThumbnailFilePath.String))) {
		chosen = right
	}
	left.ThumbnailFilePath, right.ThumbnailFilePath = chosen.ThumbnailFilePath, chosen.ThumbnailFilePath
	left.thumbnail, right.thumbnail = chosen.thumbnail, chosen.thumbnail
}

// CopyPlaylistThumbnail preserves a thumbnail already selected and remapped by
// the merge. Reapplying the policy to an original alias could select a filename
// that no longer exists after identical media have been deduplicated.
func CopyPlaylistThumbnail(target, source *PlaylistItem) {
	target.ThumbnailFilePath = source.ThumbnailFilePath
	target.thumbnail = source.thumbnail
}
