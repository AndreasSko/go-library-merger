package model

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlaylistThumbnailPolicyCanonicalIdentityAndAbsence(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(map[bool]string{false: "forward", true: "reverse"}[reverse], func(t *testing.T) {
			left := &PlaylistItem{ThumbnailFilePath: sql.NullString{String: "z.jpg", Valid: true}, thumbnail: &IndependentMedia{Hash: "native", validatedHash: "canonical", validatedSourceHash: "native"}}
			right := &PlaylistItem{ThumbnailFilePath: sql.NullString{String: "a.jpg", Valid: true}, thumbnail: &IndependentMedia{Hash: "canonical"}}
			if reverse {
				left, right = right, left
			}
			NormalizePlaylistThumbnails(left, right)
			require.Equal(t, "a.jpg", left.ThumbnailFilePath.String)
			require.Equal(t, left.ThumbnailFilePath, right.ThumbnailFilePath)
			require.True(t, left.Equals(right))
			// Unlinked callers can carry stale data in a nullable absence.
			left.ThumbnailFilePath = sql.NullString{String: "stale.jpg"}
			right.ThumbnailFilePath = sql.NullString{Valid: true}
			NormalizePlaylistThumbnails(left, right)
			require.Equal(t, sql.NullString{}, left.ThumbnailFilePath)
			require.Equal(t, sql.NullString{}, right.ThumbnailFilePath)
			require.Nil(t, left.thumbnail)
			require.Nil(t, right.thumbnail)
		})
	}
}

func TestPlaylistChangedFieldsOrderAndUnknownFormatterRemainVisible(t *testing.T) {
	left := &PlaylistItem{Label: "before"}
	right := &PlaylistItem{Label: "after", StartTrimOffsetTicks: sql.NullInt64{Int64: 10, Valid: true}, EndTrimOffsetTicks: sql.NullInt64{Int64: 20, Valid: true}, Accuracy: 1, EndAction: 2}
	require.Equal(t, []string{"label", "startTrimOffsetTicks", "endTrimOffsetTicks", "accuracy", "endAction"}, PlaylistChangedFields(left, right))
	require.Equal(t, "StartTrimOffsetTicks: Not set", PlaylistChangedFieldText(left, []string{"startTrimOffsetTicks"}))
	require.Equal(t, "Unrecognized setting: unknown", PlaylistChangedFieldText(left, []string{"unknown"}))
	require.Equal(t, "Unrecognized setting: durationTicks", PlaylistChangedFieldText(left, []string{"durationTicks"}))
	require.Equal(t, "Unrecognized setting: accuracy", PlaylistChangedFieldText((*PlaylistItem)(nil), []string{"accuracy"}))
}
