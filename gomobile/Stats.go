package gomobile

import (
	"reflect"

	"github.com/AndreasSko/go-library-merger/model"
)

// DatabaseStats represents the rough number of entries
// within a Database{} by defining it as the length
// of the slices.
type DatabaseStats struct {
	BlockRange int
	Bookmark   int
	InputField int
	Location   int
	Note       int
	Tag        int
	TagMap     int
	UserMark   int
}

// Stats generates a DatabaseStats for the given mergeSide
func (dbw *DatabaseWrapper) Stats(side string) *DatabaseStats {
	var db *model.Database

	switch side {
	case "leftSide":
		db = dbw.left
	case "rightSide":
		db = dbw.right
	case "mergeSide":
		db = dbw.merged
	default:
		db = nil
	}

	if db == nil {
		return &DatabaseStats{}
	}

	return &DatabaseStats{
		BlockRange: countSliceEntries(db.BlockRange),
		Bookmark:   countSliceEntries(db.Bookmark),
		InputField: countSliceEntries(db.InputField),
		Location:   countSliceEntries(db.Location),
		Note:       countSliceEntries(db.Note),
		Tag:        countSliceEntries(db.Tag),
		TagMap:     countSliceEntries(db.TagMap),
		UserMark:   countSliceEntries(db.UserMark),
	}
}

func countSliceEntries(entries interface{}) int {
	count := 0

	slice := reflect.ValueOf(entries)
	tp := slice.Kind()
	switch tp {
	case reflect.Slice:
		for j := 0; j < slice.Len(); j++ {
			if slice.Index(j).IsNil() {
				continue
			}
			count++
		}
	default:
		return 0
	}

	return count
}
