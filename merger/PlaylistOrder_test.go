package merger

import (
	"database/sql"
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/require"
)

func TestPlaylistContradictoryOrder(t *testing.T) {
	sequence := func(ids ...int32) []*model.TagMap {
		out := []*model.TagMap{nil}
		for pos, id := range ids {
			out = append(out, &model.TagMap{TagMapID: pos + 1, TagID: 1, PlaylistItemID: sql.NullInt32{Int32: id, Valid: true}, Position: pos})
		}
		return out
	}
	_, _, err := MergeTagMaps(sequence(1, 2), sequence(2, 1), nil)
	require.ErrorContains(t, err, "conflicting playlist order")
}
