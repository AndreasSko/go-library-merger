package model

import (
	"database/sql"
	"slices"
	"strconv"
	"strings"
)

type playlistSetting struct {
	field, name, value string
}

// playlistSettings returns the settings of a playlist entry that a user can
// change, in display order. Thumbnails are left out: the chosen side brings
// its own thumbnail.
func playlistSettings(entry Model) []playlistSetting {
	switch m := entry.(type) {
	case *PlaylistItem:
		return []playlistSetting{
			{"label", "Label", m.Label},
			{"startTrimOffsetTicks", "StartTrimOffsetTicks", nullInt64Text(m.StartTrimOffsetTicks)},
			{"endTrimOffsetTicks", "EndTrimOffsetTicks", nullInt64Text(m.EndTrimOffsetTicks)},
			{"accuracy", "Accuracy", strconv.Itoa(m.Accuracy)},
			{"endAction", "EndAction", strconv.Itoa(m.EndAction)},
		}
	case *PlaylistItemLocationMap:
		return []playlistSetting{
			{"baseDurationTicks", "BaseDurationTicks", nullInt64Text(m.BaseDurationTicks)},
			{"majorMultimediaType", "MajorMultimediaType", strconv.Itoa(m.MajorMultimediaType)},
		}
	case *PlaylistItemIndependentMediaMap:
		return []playlistSetting{
			{"durationTicks", "DurationTicks", strconv.FormatInt(m.DurationTicks, 10)},
		}
	}
	return nil
}

func nullInt64Text(value sql.NullInt64) string {
	if !value.Valid {
		return "Not set"
	}
	return strconv.FormatInt(value.Int64, 10)
}

// PlaylistChangedFields lists the playlist settings that differ, in display
// order. It is empty for entries that are not part of a playlist.
func PlaylistChangedFields(left, right Model) []string {
	l, r := playlistSettings(left), playlistSettings(right)
	fields := []string{}
	if len(l) != len(r) {
		return fields
	}
	for i := range l {
		if l[i].value != r[i].value {
			fields = append(fields, l[i].field)
		}
	}
	return fields
}

// PlaylistChangedFieldText formats the given settings of a playlist entry.
func PlaylistChangedFieldText(entry Model, fields []string) string {
	lines := []string{}
	for _, setting := range playlistSettings(entry) {
		if slices.Contains(fields, setting.field) {
			lines = append(lines, setting.name+": "+setting.value)
		}
	}
	return strings.Join(lines, "\n")
}
