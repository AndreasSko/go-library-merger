package merger

import (
	"testing"

	"github.com/AndreasSko/go-library-merger/model"
)

func TestMigrationMissingLocationsDoesNotPanic(t *testing.T) {
	for _, tc := range []struct {
		name string
		id   int
	}{
		{name: "negative", id: -1},
		{name: "zero", id: 0},
		{name: "absent", id: 1},
		{name: "out of range", id: 1000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			left := &model.Database{Location: []*model.Location{nil}, UserMark: []*model.UserMark{nil, {UserMarkID: 1, LocationID: tc.id, UserMarkGUID: "fictional"}}}
			right := model.MakeDatabaseCopy(left)
			PrepareDatabasesPreMerge(left, right)
		})
	}
}
