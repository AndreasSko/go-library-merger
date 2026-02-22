// +build !windows

package gomobile

import (
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseWrapper_Stats(t *testing.T) {
	dbw := &DatabaseWrapper{
		left:   leftMultiCollision,
		right:  emptyDB,
		merged: mergedAllLeftDB,
	}

	left := &DatabaseStats{
		BlockRange: 4,
		Bookmark:   0,
		InputField: 2,
		Location:   1,
		Note:       0,
		Tag:        0,
		TagMap:     0,
		UserMark:   4,
	}
	right := &DatabaseStats{
		BlockRange: 0,
		Bookmark:   0,
		Location:   0,
		Note:       0,
		Tag:        0,
		TagMap:     0,
		UserMark:   0,
	}
	merged := &DatabaseStats{
		BlockRange: 3,
		Bookmark:   1,
		InputField: 3,
		Location:   4,
		Note:       3,
		Tag:        4,
		TagMap:     3,
		UserMark:   3,
	}

	assert.Equal(t, left, dbw.Stats("leftSide"))
	assert.Equal(t, right, dbw.Stats("rightSide"))
	assert.Equal(t, merged, dbw.Stats("mergeSide"))
	assert.Equal(t, &DatabaseStats{}, dbw.Stats("wrongSide"))
}

func Test_countSliceEntries(t *testing.T) {
	assert.Equal(t, 3, countSliceEntries([]*model.BlockRange{nil, {}, {}, {}}))
	assert.NotPanics(t, func() {
		assert.Equal(t, 0, countSliceEntries(nil))
		assert.Equal(t, 0, countSliceEntries([]string{}))
		assert.Equal(t, 0, countSliceEntries(0))
		assert.Equal(t, 0, countSliceEntries("A"))
	})
}
