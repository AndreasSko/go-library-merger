// +build !windows

package gomobile

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/AndreasSko/go-library-merger/merger"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/assert"
)

func TestMergeConflictsWrapper_InitDBWrapper(t *testing.T) {
	dbw := &DatabaseWrapper{
		left: &model.Database{},
	}

	mcw := MergeConflictsWrapper{}
	mcw.InitDBWrapper(dbw)

	assert.Same(t, dbw, mcw.DBWrapper)
}

func TestMergeConflictsWrapper_addConflicts(t *testing.T) {
	conflicts := map[string]merger.MergeConflict{
		"1": {
			Left: &model.Bookmark{
				Title: "1Left",
			},
			Right: &model.Bookmark{
				Title: "1Right",
			},
		},
		"2": {
			Left: &model.Tag{
				Name: "2Left",
			},
			Right: &model.Tag{
				Name: "2Right",
			},
		},
	}

	mcw := MergeConflictsWrapper{}
	mcw.addConflicts(conflicts)
	assert.Equal(t, conflicts, mcw.conflicts)
	assert.Equal(t, map[string]bool{"1": true, "2": true}, mcw.unsolvedConflicts)

	conflicts["3"] = merger.MergeConflict{
		Left: &model.Tag{
			Name: "2Left",
		},
		Right: &model.Tag{
			Name: "2Right",
		},
	}

	mcw.addConflicts(conflicts)
	assert.Equal(t, conflicts, mcw.conflicts)
	assert.Equal(t, map[string]bool{"1": true, "2": true, "3": true}, mcw.unsolvedConflicts)
}

func TestMergeConflictsWrapper_NextConflict(t *testing.T) {
	db := &model.Database{
		Location: []*model.Location{
			nil,
			{
				LocationID: 1,
				Title:      sql.NullString{"Location-Title", true},
			},
		},
	}

	mcw := MergeConflictsWrapper{
		DBWrapper: &DatabaseWrapper{
			merged: db,
		},
		conflicts: map[string]merger.MergeConflict{
			"1": {
				Left: &model.Bookmark{
					LocationID: 1,
					Title:      "1Left",
				},
				Right: &model.Bookmark{
					LocationID: 1,
					Title:      "1Right",
				},
			},
			"2": {
				Left: &model.Tag{
					Name: "2Left",
				},
				Right: &model.Tag{
					Name: "2Right",
				},
			},
		},
		unsolvedConflicts: map[string]bool{"1": true, "2": true},
	}

	expectedConflicts := map[string]*MergeConflict{
		"1": {
			Key: "1",
			Left: jsonMarhshalIgnoreErr(modelRelatedTuple{
				Model:   mcw.conflicts["1"].Left,
				Related: model.Related{Location: db.Location[1], PublicationLocation: db.Location[1]},
			}),
			Right: jsonMarhshalIgnoreErr(modelRelatedTuple{
				Model:   mcw.conflicts["1"].Right,
				Related: model.Related{Location: db.Location[1], PublicationLocation: db.Location[1]},
			}),
		},
		"2": {
			Key: "2",
			Left: jsonMarhshalIgnoreErr(modelRelatedTuple{
				Model:   mcw.conflicts["2"].Left,
				Related: model.Related{},
			}),
			Right: jsonMarhshalIgnoreErr(modelRelatedTuple{
				Model:   mcw.conflicts["2"].Right,
				Related: model.Related{},
			}),
		},
	}

	conflicts := map[string]*MergeConflict{}
	for {
		conflict, err := mcw.NextConflict()
		if err != nil {
			assert.EqualError(t, err, "There are no unsolved conflicts")
			break
		}
		conflicts[conflict.Key] = conflict
		delete(mcw.unsolvedConflicts, conflict.Key)
	}

	assert.Equal(t, expectedConflicts, conflicts)

	mcw.DBWrapper = nil
	mcw.unsolvedConflicts["2"] = true
	conflict, err := mcw.NextConflict()
	assert.NoError(t, err)
	assert.Equal(t,
		jsonMarhshalIgnoreErr(modelRelatedTuple{
			Model:   mcw.conflicts["2"].Right,
			Related: model.Related{},
		}),
		conflict.Right)
	delete(mcw.unsolvedConflicts, conflict.Key)

	_, err = mcw.NextConflict()
	assert.EqualError(t, err, "There are no unsolved conflicts")
}

func jsonMarhshalIgnoreErr(m interface{}) string {
	result, _ := json.Marshal(m)
	return string(result)
}

func TestMergeConflictsWrapper_SolveConflict(t *testing.T) {
	mcw := MergeConflictsWrapper{}
	assert.EqualError(t, mcw.SolveConflict("bla", "leftSide"), "There are no unsolved conflicts")

	mcw = MergeConflictsWrapper{
		conflicts: map[string]merger.MergeConflict{
			"1": {
				Left: &model.Bookmark{
					Title: "1Left",
				},
				Right: &model.Bookmark{
					Title: "1Right",
				},
			},
			"2": {
				Left: &model.Tag{
					Name: "2Left",
				},
				Right: &model.Tag{
					Name: "2Right",
				},
			},
		},
		unsolvedConflicts: map[string]bool{"1": true, "2": true},
	}
	assert.EqualError(t, mcw.SolveConflict("5", "leftSide"), "Unsolved conflict with key 5 does not exist")
	assert.EqualError(t, mcw.SolveConflict("1", "wrongSide"), "Side wrongSide is not valid")

	assert.NoError(t, mcw.SolveConflict("1", "rightSide"))
	assert.Equal(t,
		merger.MergeSolution{
			Side:      merger.RightSide,
			Solution:  mcw.conflicts["1"].Right,
			Discarded: mcw.conflicts["1"].Left,
		},
		mcw.solutions["1"])
	assert.EqualError(t, mcw.SolveConflict("1", "rightSide"), "Unsolved conflict with key 1 does not exist")

	assert.NoError(t, mcw.SolveConflict("2", "leftSide"))
	assert.Equal(t,
		merger.MergeSolution{
			Side:      merger.LeftSide,
			Solution:  mcw.conflicts["2"].Left,
			Discarded: mcw.conflicts["2"].Right,
		},
		mcw.solutions["2"])

	assert.Empty(t, mcw.unsolvedConflicts)
	_, err := mcw.NextConflict()
	assert.EqualError(t, err, "There are no unsolved conflicts")
}
