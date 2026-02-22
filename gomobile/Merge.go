package gomobile

import (
	"github.com/AndreasSko/go-library-merger/merger"
	"github.com/pkg/errors"
	_ "golang.org/x/mobile/bind"
)

// MergeLocations merges locations
func (dbw *DatabaseWrapper) MergeLocations() error {
	mergedLocations, locationIDChanges, err := merger.MergeLocations(dbw.leftTmp.Location, dbw.rightTmp.Location)
	if err != nil {
		return errors.Wrap(err, "Could not merge locations")
	}
	dbw.merged.Location = mergedLocations
	merger.UpdateLRIDs(dbw.leftTmp.Bookmark, dbw.rightTmp.Bookmark, "LocationID", locationIDChanges)
	merger.UpdateLRIDs(dbw.leftTmp.Bookmark, dbw.rightTmp.Bookmark, "PublicationLocationID", locationIDChanges)
	merger.UpdateLRIDs(dbw.leftTmp.InputField, dbw.rightTmp.InputField, "LocationID", locationIDChanges)
	merger.UpdateLRIDs(dbw.leftTmp.Note, dbw.rightTmp.Note, "LocationID", locationIDChanges)
	merger.UpdateLRIDs(dbw.leftTmp.TagMap, dbw.rightTmp.TagMap, "LocationID", locationIDChanges)
	merger.UpdateLRIDs(dbw.leftTmp.UserMark, dbw.rightTmp.UserMark, "LocationID", locationIDChanges)
	merger.UpdateLRIDs(dbw.leftTmp.PlaylistItemLocationMap, dbw.rightTmp.PlaylistItemLocationMap, "LocationID", locationIDChanges)

	return nil
}

// MergeBookmarks merges bookmarks
func (dbw *DatabaseWrapper) MergeBookmarks(conflictSolver string, mcw *MergeConflictsWrapper) error {
	var conflictSolution = mcw.solutions
	if conflictSolution == nil {
		conflictSolution = map[string]merger.MergeSolution{}
	}
	for {
		merged, _, err := merger.MergeBookmarks(dbw.leftTmp.Bookmark, dbw.rightTmp.Bookmark, conflictSolution)
		if err == nil {
			dbw.merged.Bookmark = merged
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			if conflictSolver == "" {
				mcw.addConflicts(err.Conflicts)
				return MergeConflictError{}
			}
			var resErr error
			newSolutions, resErr := merger.AutoResolveConflicts(err.Conflicts, conflictSolver)
			if resErr != nil {
				return errors.Wrap(resErr, "Could not automatically solve conflicts for bookmarks")
			}
			addToSolutions(conflictSolution, newSolutions)
		default:
			return errors.Wrap(err, "Could not merge bookmarks")
		}
	}

	return nil
}

// MergeInputField merges inputFields
func (dbw *DatabaseWrapper) MergeInputFields(conflictSolver string, mcw *MergeConflictsWrapper) error {
	var conflictSolution = mcw.solutions
	if conflictSolution == nil {
		conflictSolution = map[string]merger.MergeSolution{}
	}
	for {
		merged, _, err := merger.MergeInputFields(dbw.leftTmp.InputField, dbw.rightTmp.InputField, conflictSolution)
		if err == nil {
			dbw.merged.InputField = merged
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			if conflictSolver == "" {
				mcw.addConflicts(err.Conflicts)
				return MergeConflictError{}
			}
			var resErr error
			newSolutions, resErr := merger.AutoResolveConflicts(err.Conflicts, conflictSolver)
			if resErr != nil {
				return errors.Wrap(resErr, "Could not automatically solve conflicts for inputFields")
			}
			addToSolutions(conflictSolution, newSolutions)
		default:
			return errors.Wrap(err, "Could not merge inputFields")
		}
	}

	return nil
}

// MergeTags merges tags
func (dbw *DatabaseWrapper) MergeTags() error {
	var conflictSolution map[string]merger.MergeSolution
	for {
		merged, idChanges, err := merger.MergeTags(dbw.leftTmp.Tag, dbw.rightTmp.Tag, conflictSolution)
		if err == nil {
			dbw.merged.Tag = merged
			merger.UpdateLRIDs(dbw.leftTmp.TagMap, dbw.rightTmp.TagMap, "TagID", idChanges)
			break
		}
		return errors.Wrap(err, "Could not merge tags")
	}

	return nil
}

// MergeUserMarkAndBlockRange merges UserMarks and BlockRanges
func (dbw *DatabaseWrapper) MergeUserMarkAndBlockRange(conflictSolver string, mcw *MergeConflictsWrapper) error {
	var conflictSolution = mcw.solutions
	if conflictSolution == nil {
		conflictSolution = map[string]merger.MergeSolution{}
	}
	for {
		mergedUserMarks, mergedBlockRanges, idChanges, err := merger.MergeUserMarkAndBlockRange(dbw.leftTmp.UserMark, dbw.leftTmp.BlockRange, dbw.rightTmp.UserMark, dbw.rightTmp.BlockRange, conflictSolution)
		if err == nil {
			dbw.merged.UserMark = mergedUserMarks
			dbw.merged.BlockRange = mergedBlockRanges
			merger.UpdateLRIDs(dbw.leftTmp.Note, dbw.rightTmp.Note, "UserMarkID", idChanges)
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			if conflictSolver == "" {
				mcw.addConflicts(err.Conflicts)
				return MergeConflictError{}
			}
			var resErr error
			newSolutions, resErr := merger.AutoResolveConflicts(err.Conflicts, conflictSolver)
			if resErr != nil {
				return errors.Wrap(resErr, "Could not automatically solve conflicts for markings")
			}
			addToSolutions(conflictSolution, newSolutions)
		default:
			return errors.Wrap(err, "Could not merge markings")
		}
	}

	return nil
}

// MergeNotes merges notes
func (dbw *DatabaseWrapper) MergeNotes(conflictSolver string, mcw *MergeConflictsWrapper) error {
	var conflictSolution = mcw.solutions
	if conflictSolution == nil {
		conflictSolution = map[string]merger.MergeSolution{}
	}
	for {
		merged, idChanges, err := merger.MergeNotes(dbw.leftTmp.Note, dbw.rightTmp.Note, conflictSolution)
		if err == nil {
			dbw.merged.Note = merged
			merger.UpdateLRIDs(dbw.leftTmp.TagMap, dbw.rightTmp.TagMap, "NoteID", idChanges)
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			if conflictSolver == "" {
				mcw.addConflicts(err.Conflicts)
				return MergeConflictError{}
			}
			var resErr error
			newSolutions, resErr := merger.AutoResolveConflicts(err.Conflicts, conflictSolver)
			if resErr != nil {
				return errors.Wrap(resErr, "Could not automatically solve conflicts for notes")
			}
			addToSolutions(conflictSolution, newSolutions)
		default:
			return errors.Wrap(err, "Could not merge notes")
		}
	}

	return nil
}

// MergeTagMaps merges tagMaps
func (dbw *DatabaseWrapper) MergeTagMaps() error {
	dbw.tagMapsMerged = false
	// Existing callers already finish with MergeTagMaps. Include playlists here
	// so they cannot silently export dangling playlist membership references.
	if !dbw.playlistsMerged {
		if err := dbw.MergePlaylists("", &MergeConflictsWrapper{}); err != nil {
			return err
		}
	}
	var conflictSolution map[string]merger.MergeSolution
	for {
		merged, _, err := merger.MergeTagMaps(dbw.leftTmp.TagMap, dbw.rightTmp.TagMap, conflictSolution)
		if err == nil {
			dbw.merged.TagMap = merged
			dbw.tagMapsMerged = true
			break
		}

		return errors.Wrap(err, "Could not merge tagMaps")
	}

	return nil
}

// addToSolutions adds new mergeSolutions to the existing map of mergeSolutions
func addToSolutions(solutions map[string]merger.MergeSolution, new map[string]merger.MergeSolution) {
	for key, value := range new {
		solutions[key] = value
	}
}

// MergePlaylists merges playlist items and media after MergeLocations and
// MergeTags. Conflicts can be resolved with chooseLeft/chooseRight or mcw.
func (dbw *DatabaseWrapper) MergePlaylists(conflictSolver string, mcw *MergeConflictsWrapper) error {
	if dbw.playlistsMerged {
		return nil
	}
	if mcw == nil {
		mcw = &MergeConflictsWrapper{}
	}
	solutions := mcw.solutions
	if solutions == nil {
		solutions = map[string]merger.MergeSolution{}
	}
	for {
		err := merger.MergePlaylists(dbw.leftTmp, dbw.rightTmp, dbw.merged, solutions)
		if err == nil {
			dbw.playlistsMerged = true
			return nil
		}
		conflict, ok := err.(merger.MergeConflictError)
		if !ok {
			return errors.Wrap(err, "Could not merge playlists")
		}
		if conflictSolver == "" {
			mcw.addConflicts(conflict.Conflicts)
			return MergeConflictError{}
		}
		resolved, err := merger.AutoResolveConflicts(conflict.Conflicts, conflictSolver)
		if err != nil {
			return err
		}
		addToSolutions(solutions, resolved)
	}
}
