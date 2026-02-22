package merger

import (
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/AndreasSko/go-library-merger/model"
	"github.com/stretchr/testify/assert"
)

func TestAutoResolveConflicts(t *testing.T) {
	conflicts := map[string]MergeConflict{
		"bookmarkCollision": {
			Left: &model.Bookmark{
				Title: "LeftBookmark",
			},
			Right: &model.Bookmark{
				Title: "RightBookmark",
			},
		},
		"noteCollision": {
			Left: &model.Note{
				GUID: "LeftNote",
			},
			Right: &model.Note{
				GUID: "RightNote",
			},
		},
	}

	// Choose left
	expectedResult := map[string]MergeSolution{
		"bookmarkCollision": {
			Side: LeftSide,
			Solution: &model.Bookmark{
				Title: "LeftBookmark",
			},
			Discarded: &model.Bookmark{
				Title: "RightBookmark",
			},
		},
		"noteCollision": {
			Side: LeftSide,
			Solution: &model.Note{
				GUID: "LeftNote",
			},
			Discarded: &model.Note{
				GUID: "RightNote",
			},
		},
	}

	result, err := AutoResolveConflicts(conflicts, "chooseNewest")
	assert.Error(t, err)
	assert.Nil(t, result)

	result, err = AutoResolveConflicts(conflicts, "wrongResolver")
	assert.Error(t, err)
	assert.Nil(t, result)

	result, err = AutoResolveConflicts(conflicts, "")
	assert.NoError(t, err)
	assert.Nil(t, result)

	result, err = AutoResolveConflicts(conflicts, "chooseLeft")
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestSolveConflictByChoosingX(t *testing.T) {
	conflicts := map[string]MergeConflict{
		"bookmarkCollision": {
			Left: &model.Bookmark{
				Title: "LeftBookmark",
			},
			Right: &model.Bookmark{
				Title: "RightBookmark",
			},
		},
		"noteCollision": {
			Left: &model.Note{
				GUID: "LeftNote",
			},
			Right: &model.Note{
				GUID: "RightNote",
			},
		},
	}

	// Choose left
	expectedResult := map[string]MergeSolution{
		"bookmarkCollision": {
			Side: LeftSide,
			Solution: &model.Bookmark{
				Title: "LeftBookmark",
			},
			Discarded: &model.Bookmark{
				Title: "RightBookmark",
			},
		},
		"noteCollision": {
			Side: LeftSide,
			Solution: &model.Note{
				GUID: "LeftNote",
			},
			Discarded: &model.Note{
				GUID: "RightNote",
			},
		},
	}

	result, err := SolveConflictByChoosingLeft(conflicts)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)

	// Choose right
	expectedResult = map[string]MergeSolution{
		"bookmarkCollision": {
			Side: RightSide,
			Solution: &model.Bookmark{
				Title: "RightBookmark",
			},
			Discarded: &model.Bookmark{
				Title: "LeftBookmark",
			},
		},
		"noteCollision": {
			Side: RightSide,
			Solution: &model.Note{
				GUID: "RightNote",
			},
			Discarded: &model.Note{
				GUID: "LeftNote",
			},
		},
	}

	result, err = SolveConflictByChoosingRight(conflicts)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestSolveConflictByChoosingNewest(t *testing.T) {
	conflicts := map[string]MergeConflict{
		"leftNewer": {
			Left: &model.Note{
				GUID:         "Left",
				LastModified: "2020-09-15T13:00:00Z",
				Created:      "2020-09-15T12:00:00+00:00",
			},
			Right: &model.Note{
				GUID:         "Right",
				LastModified: "2020-09-15T12:00:00+0000",
				Created:      "2020-09-15T12:00:00+00:00",
			},
		},
		"rightNewer": {
			Left: &model.Note{
				GUID:         "Left",
				LastModified: "2020-09-15T13:00:00+00:00",
				Created:      "2020-09-15T13:00:00+00:00",
			},
			Right: &model.Note{
				GUID:         "Right",
				LastModified: "2020-09-15T13:01:00+0000",
				Created:      "2020-09-15T13:00:00+00:00",
			},
		},
	}

	expectedResult := map[string]MergeSolution{
		"leftNewer": {
			Side: LeftSide,
			Solution: &model.Note{
				GUID:         "Left",
				LastModified: "2020-09-15T13:00:00Z",
				Created:      "2020-09-15T12:00:00+00:00",
			},
			Discarded: &model.Note{
				GUID:         "Right",
				LastModified: "2020-09-15T12:00:00+0000",
				Created:      "2020-09-15T12:00:00+00:00",
			},
		},
		"rightNewer": {
			Side: RightSide,
			Solution: &model.Note{
				GUID:         "Right",
				LastModified: "2020-09-15T13:01:00+0000",
				Created:      "2020-09-15T13:00:00+00:00",
			},
			Discarded: &model.Note{
				GUID:         "Left",
				LastModified: "2020-09-15T13:00:00+00:00",
				Created:      "2020-09-15T13:00:00+00:00",
			},
		},
	}

	result, err := SolveConflictByChoosingNewest(conflicts)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)

	conflicts = map[string]MergeConflict{
		"bookmarkCollision": {
			Left: &model.Bookmark{
				Title: "LeftBookmark",
			},
			Right: &model.Bookmark{
				Title: "RightBookmark",
			},
		},
	}
	_, err = SolveConflictByChoosingNewest(conflicts)
	assert.Error(t, err)
}

func Test_parseResolver(t *testing.T) {
	resolver, err := parseResolver("")
	assert.NoError(t, err)
	assert.Nil(t, resolver)

	// https://github.com/stretchr/testify/issues/182#issuecomment-495359313
	resolver, err = parseResolver("chooseLeft")
	assert.NoError(t, err)
	assert.Equal(t,
		"github.com/AndreasSko/go-library-merger/merger.SolveConflictByChoosingLeft",
		runtime.FuncForPC(reflect.ValueOf(resolver).Pointer()).Name())

	resolver, err = parseResolver("chooseRight")
	assert.NoError(t, err)
	assert.Equal(t,
		"github.com/AndreasSko/go-library-merger/merger.SolveConflictByChoosingRight",
		runtime.FuncForPC(reflect.ValueOf(resolver).Pointer()).Name())

	resolver, err = parseResolver("chooseNewest")
	assert.NoError(t, err)
	assert.Equal(t,
		"github.com/AndreasSko/go-library-merger/merger.SolveConflictByChoosingNewest",
		runtime.FuncForPC(reflect.ValueOf(resolver).Pointer()).Name())

	resolver, err = parseResolver("nonexistent")
	assert.EqualError(t, err, "nonexistent is not a valid conflict resolver. Can be 'chooseNewest', 'chooseLeft', or 'chooseRight'")
	assert.Nil(t, resolver)
}

func Test_parseDateTimeString(t *testing.T) {
	tests := []struct {
		name     string
		dateTime string
		want     func(t *testing.T) time.Time
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name:     "Timezone without :",
			dateTime: "2023-07-15T09:50:47+0200",
			want: func(*testing.T) time.Time {
				want, err := time.Parse("2006-01-02T15:04:05-0700", "2023-07-15T09:50:47+0200")
				assert.NoError(t, err)
				return want
			},
			wantErr: assert.NoError,
		},
		{
			name:     "Timezone with :",
			dateTime: "2023-07-13T05:36:36+00:00",
			want: func(*testing.T) time.Time {
				want, err := time.Parse("2006-01-02T15:04:05-07:00", "2023-07-13T05:36:36+00:00")
				assert.NoError(t, err)
				return want
			},
			wantErr: assert.NoError,
		},
		{
			name:     "RFC3339",
			dateTime: "2023-07-17T04:45:03Z",
			want: func(*testing.T) time.Time {
				want, err := time.Parse(time.RFC3339, "2023-07-17T04:45:03Z")
				assert.NoError(t, err)
				return want
			},
			wantErr: assert.NoError,
		},
		{
			name:     "Wrong format",
			dateTime: "2023-07-13T05:36:36Z+00:00",
			want: func(t *testing.T) time.Time {
				return time.Time{}
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "failed to parse dateTime 2023-07-13T05:36:36Z+00:00")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDateTimeString(tt.dateTime)
			tt.wantErr(t, err)
			want := tt.want(t)
			assert.Equal(t, want, got)
		})
	}
}
