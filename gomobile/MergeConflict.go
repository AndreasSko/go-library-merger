package gomobile

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/AndreasSko/go-library-merger/merger"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/AndreasSko/go-library-merger/publication"
)

// MergeConflictError indicates that a conflict happened while merging. It
// is equivalent to merger.MergeConflictError, but does not contain the
// actual conflicts to make it compatible with Gomobile.
type MergeConflictError struct {
	Err string
}

func (e MergeConflictError) Error() string {
	return fmt.Sprintf("There were conflicts while trying to merge %s", e.Err)
}

// MergeConflictsWrapper wraps mergeConflicts and their solutions
type MergeConflictsWrapper struct {
	DBWrapper         *DatabaseWrapper
	conflicts         map[string]merger.MergeConflict
	unsolvedConflicts map[string]bool
	solutions         map[string]merger.MergeSolution
}

// MergeConflict represents two Models that collide. It is equvalent
// to merger.MergeConflict, but represents the Models as strings
// to make it compatible with Gomobile.
type MergeConflict struct {
	Key           string
	Left          string
	Right         string
	ChangedFields string // JSON array of meaningful playlist setting field names.
}

// modelRelatedTuple contains a model and its related entries
type modelRelatedTuple struct {
	Model     model.Model   `json:"model"`
	Related   model.Related `json:"related"`
	SourceURL string        `json:"sourceURL,omitempty"`
}

// InitDBWrapper initializes the DatabaseWrapper for the MergeConflictsWrapper
// so the DB is accessible for pretty printing.
func (mcw *MergeConflictsWrapper) InitDBWrapper(dbw *DatabaseWrapper) {
	mcw.DBWrapper = dbw
}

// addConflicts adds the given conflicts to the MergeConflictWrapper
// and makes sure that the keys are added to conflictsKeys. This
// makes sure, that the order stays the same. If the conflict already
// exists, it is skipped.
func (mcw *MergeConflictsWrapper) addConflicts(conflicts map[string]merger.MergeConflict) {
	if mcw.conflicts == nil {
		mcw.conflicts = make(map[string]merger.MergeConflict, len(conflicts))
	}

	if mcw.unsolvedConflicts == nil {
		mcw.unsolvedConflicts = make(map[string]bool, len(conflicts))
	}

	for key, value := range conflicts {
		if _, exists := mcw.conflicts[key]; !exists {
			mcw.conflicts[key] = value
			mcw.unsolvedConflicts[key] = true
		}
	}
}

// NextConflict returns the next conflict that should be solved. If there
// are no left, it returns an error
func (mcw *MergeConflictsWrapper) NextConflict() (*MergeConflict, error) {
	if mcw.unsolvedConflicts == nil || len(mcw.unsolvedConflicts) == 0 {
		return nil, errors.New("There are no unsolved conflicts")
	}

	if mcw.DBWrapper == nil {
		mcw.DBWrapper = &DatabaseWrapper{merged: nil}
	}

	// Get any conflict
	var conflictKey string
	for key := range mcw.unsolvedConflicts {
		conflictKey = key
	}
	conflict := mcw.conflicts[conflictKey]

	result := &MergeConflict{
		Key: conflictKey,
	}
	changed, err := json.Marshal(model.PlaylistChangedFields(conflict.Left, conflict.Right))
	if err != nil {
		return nil, err
	}
	result.ChangedFields = string(changed)
	leftRelated := conflict.Left.RelatedEntries(mcw.DBWrapper.merged)
	jsn, err := json.Marshal(modelRelatedTuple{
		Model:     conflict.Left,
		Related:   leftRelated,
		SourceURL: playlistConflictSourceURL(conflict.Left, leftRelated, mcw.DBWrapper.left),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error while marshalling to JSON")
	}
	result.Left = string(jsn)

	rightRelated := conflict.Right.RelatedEntries(mcw.DBWrapper.merged)
	jsn, err = json.Marshal(modelRelatedTuple{
		Model:     conflict.Right,
		Related:   rightRelated,
		SourceURL: playlistConflictSourceURL(conflict.Right, rightRelated, mcw.DBWrapper.right),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error while marshalling to JSON")
	}
	result.Right = string(jsn)

	return result, nil
}

// SolveConflict solves a mergeConflict represented by key and chooses the given side
func (mcw *MergeConflictsWrapper) SolveConflict(key string, side string) error {
	if mcw.unsolvedConflicts == nil || len(mcw.unsolvedConflicts) == 0 {
		return errors.New("There are no unsolved conflicts")
	}
	if _, exists := mcw.unsolvedConflicts[key]; !exists {
		return errors.Errorf("Unsolved conflict with key %s does not exist", key)
	}

	if mcw.solutions == nil {
		mcw.solutions = make(map[string]merger.MergeSolution, len(mcw.conflicts))
	}

	switch side {
	case "leftSide":
		mcw.solutions[key] = merger.MergeSolution{
			Side:      merger.LeftSide,
			Solution:  mcw.conflicts[key].Left,
			Discarded: mcw.conflicts[key].Right,
		}
	case "rightSide":
		mcw.solutions[key] = merger.MergeSolution{
			Side:      merger.RightSide,
			Solution:  mcw.conflicts[key].Right,
			Discarded: mcw.conflicts[key].Left,
		}
	default:
		return fmt.Errorf("Side %s is not valid", side)
	}

	delete(mcw.unsolvedConflicts, key)

	return nil
}

// DownloadListener is told how many bytes of a publication are downloaded,
// out of total, which is -1 if unknown
type DownloadListener interface {
	Progress(downloaded int64, total int64)
}

// MarkingSnippets returns the text of the paragraphs that the marking on the
// given side of a conflict covers, as a JSON array of publication.Snippet. If
// the text of the publication is not cached in cacheDir yet, it is downloaded
// from jw.org, and listener, if not nil, is told the progress.
func (mcw *MergeConflictsWrapper) MarkingSnippets(key string, side string, cacheDir string, listener DownloadListener) (string, error) {
	conflict, exists := mcw.conflicts[key]
	if !exists {
		return "", fmt.Errorf("conflict with key %s does not exist", key)
	}
	var mdl model.Model
	switch side {
	case "leftSide":
		mdl = conflict.Left
	case "rightSide":
		mdl = conflict.Right
	default:
		return "", fmt.Errorf("side %s is not valid", side)
	}

	snippets := []publication.Snippet{}
	if marking, ok := mdl.(*model.UserMarkBlockRange); ok && mcw.DBWrapper != nil {
		location := marking.RelatedEntries(mcw.DBWrapper.merged).Location
		var progress func(int64, int64)
		if listener != nil {
			progress = listener.Progress
		}
		result, err := publication.MarkingSnippets(context.Background(), cacheDir, location, marking.BlockRanges, progress)
		if err != nil {
			return "", err
		}
		snippets = append(snippets, result...)
	}
	jsn, err := json.Marshal(snippets)
	return string(jsn), err
}

// playlistConflictSourceURL returns a jw.org link to the source of a playlist
// conflict. An item conflict uses the item ID of its backup, so the media type
// comes from the location map of that backup. A location map conflict has the
// media type itself.
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
