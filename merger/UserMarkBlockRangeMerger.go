package merger

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AndreasSko/go-library-merger/model"
)

// brFrom indicates from which mergeSide a *BlockRange
// is coming from.
type brFrom struct {
	side MergeSide
	br   *model.BlockRange
}

// umbrFrom indicates from which mergeSide a []*UserMarkBlockRange
// is coming from.
type umbrFrom struct {
	side MergeSide
	umbr []*model.UserMarkBlockRange
}

// MergeUserMarkAndBlockRange joins UserMarks and BlockRanges from both sides and
// tries to merge them. Afterwards it will update the IDs of UserMark and BlockRange
// and returns them separately againg. If there is a collision, it will try to solve
// it using duplicate detection and - if that fails - returns an error asking for
// specification how it should handle it. MergeConflicts will be returned as a joined
// UserMarkBlockRange struct to make it easier representing conflicts.
// The returned IDChanges indicate if a UserMarkID has changed in the merge process.
func MergeUserMarkAndBlockRange(leftUM []*model.UserMark, leftBR []*model.BlockRange,
	rightUM []*model.UserMark, rightBR []*model.BlockRange,
	conflictSolution map[string]MergeSolution) ([]*model.UserMark, []*model.BlockRange, IDChanges, error) {
	if conflictSolution == nil {
		conflictSolution = map[string]MergeSolution{}
	}

	left := joinToUserMarkBlockRange(leftUM, leftBR)
	right := joinToUserMarkBlockRange(rightUM, rightBR)

	var merged []*model.UserMarkBlockRange
	var changes IDChanges
	var err error

	for {
		merged, changes, err = mergeUMBR(left, right, conflictSolution)
		if err == nil {
			um, br := splitUserMarkBlockRange(merged)
			return um, br, changes, nil
		}

		// If merge failed, try to solve conflicts using solveEqualityMergeConflict
		switch err := err.(type) {
		case MergeConflictError:
			autoConflictSolution, sErr := solveEqualityMergeConflict(err.Conflicts)
			for key, autoSol := range autoConflictSolution {
				conflictSolution[key] = autoSol
			}
			if sErr == nil {
				continue
			}
			// If no more conflicts could be solved, fail and return error
			if reflect.DeepEqual(err.Conflicts, sErr.(MergeConflictError).Conflicts) {
				return nil, nil, IDChanges{}, sErr
			}
		default:
			return nil, nil, IDChanges{}, err
		}
	}
}

// mergeUMBR merges a left and a right side of *[]UserMarkBlockRange. It will check
// for overlapping (i.e. conflicting) BlockRanges and returns an mergeConflictError
// if it finds some, asking the caller for specification how it should handle it.
// IDChanges indicate if a UserMarkID has changed in the merge process.
func mergeUMBR(left []*model.UserMarkBlockRange, right []*model.UserMarkBlockRange,
	conflictSolution map[string]MergeSolution) ([]*model.UserMarkBlockRange, IDChanges, error) {
	// First, replace conflictSolution entries with the conflicting ones on the left
	// and right side, so we don't detect them again.
	changes, invertedChanges := replaceUMBRConflictsWithSolution(&left, &right, conflictSolution)

	conflicts := map[string]MergeConflict{}

	// Ingest UserMarks & BlockRanges in a Map[LocationID]map[Identifier][]*model.BlockRange
	blRanges := ingestUMBR(left, right)

	// For each BlockRange slice per identifier, sort BlockRanges by StartToken
	for _, locationBlock := range blRanges {
		for _, identifierBlock := range locationBlock {
			// Filter out duplicates and immediatelly add them to conflicts
			identifierBlock, moreConflicts := detectAndFilterDuplicateBRs(identifierBlock, left, right)
			for key, value := range moreConflicts {
				conflicts[key] = value
			}

			// Check for overlapping intervals
			for i := 0; i < len(identifierBlock); i++ {
				br := identifierBlock[i]

			Loop:
				for j := i + 1; j < len(identifierBlock) && br.br.EndToken.Int32 >= identifierBlock[j].br.StartToken.Int32; j++ {
					// We found a collision!

					// If collision is on the same side, then ignore it
					// (it's probably not our fault and we hope its okay...)
					if br.side == identifierBlock[j].side {
						continue
					}

					// If its on different sides, add it to conflicts & make sure
					// that entries are on correct side of mergeConflict{}
					var first, second *model.UserMarkBlockRange
					if br.side == LeftSide {
						first = left[br.br.UserMarkID]
						second = right[identifierBlock[j].br.UserMarkID]
					} else {
						first = left[identifierBlock[j].br.UserMarkID]
						second = right[br.br.UserMarkID]
					}
					var conflictKey strings.Builder
					conflictKey.Grow(160)
					// Use UnixNano as a monotonically increasing number, so we
					// are able to apply conflict solutions in the right order later
					conflictKey.WriteString(fmt.Sprint(time.Now().UnixNano()))
					conflictKey.WriteString("_")
					conflictKey.WriteString(first.UniqueKey())
					conflictKey.WriteString("_")
					conflictKey.WriteString(second.UniqueKey())
					conflicts[conflictKey.String()] = MergeConflict{first, second}

					// Skip further possible collisions of this interval
					// by continuing at the next BlockRange that starts after the
					// EndToken of the current found collision, as we don't
					// know, how the user might resolve the current collision.
					for ; i < len(identifierBlock); i++ {
						if identifierBlock[i].br.StartToken.Int32 > identifierBlock[j].br.EndToken.Int32 {
							break Loop
						}
					}
				}
			}
		}
	}

	if len(conflicts) > 0 {
		return []*model.UserMarkBlockRange{}, IDChanges{}, MergeConflictError{Conflicts: conflicts}
	}

	// Add left and right to result & update (UserMark-)ID
	result := make([]*model.UserMarkBlockRange, len(left)+len(right)+2)
	i := 1
	for _, mergeSide := range []MergeSide{LeftSide, RightSide} {
		var side []*model.UserMarkBlockRange
		if mergeSide == LeftSide {
			side = left
		} else {
			side = right
		}

		for _, entry := range side {
			if entry == nil {
				continue
			}
			entry = model.MakeModelCopy(entry).(*model.UserMarkBlockRange)
			// Note IDChanges if necessary
			if entry.ID() != i {
				if mergeSide == LeftSide {
					// Check if on the other side an ID has changed to entry.ID
					// after its entry was discareded. If so, we again need update
					// the entry, as the current ID it is pointing at will be changed too.
					if val, ok := invertedChanges.Right[entry.ID()]; ok {
						changes.Right[val] = i
					}
					changes.Left[entry.ID()] = i
				} else {
					if val, ok := invertedChanges.Left[entry.ID()]; ok {
						changes.Left[val] = i
					}
					changes.Right[entry.ID()] = i
				}
			}

			result[i] = entry
			result[i].SetID(i)
			i++
		}
	}

	return result[:i], changes, nil
}

// detectAndFilterDuplicateBRs removes block Range entries that exists on both
// sides (duplicates) and only leaves the one on the left side.
// It returns a slice of brFroms sorted by StartToken
func detectAndFilterDuplicateBRs(idBlock []brFrom, left []*model.UserMarkBlockRange,
	right []*model.UserMarkBlockRange) ([]brFrom, map[string]MergeConflict) {
	conflicts := map[string]MergeConflict{}

	idBlock = sortBRFroms(idBlock)

	for i := 0; i < len(idBlock); i++ {
		if idBlock[i] == (brFrom{}) {
			continue
		}
		for j := i + 1; j < len(idBlock); j++ {
			if idBlock[j] == (brFrom{}) {
				continue
			}
			// We only need to look for entries that are "conflicting"
			if idBlock[j].br.StartToken.Int32 > idBlock[i].br.EndToken.Int32 {
				break
			}

			// Check if they equal in all except userMarkID
			// (BlockRange.Equals checks userMarkID, which we don't want here,
			// as they obviously can differ between backups..)
			if idBlock[i].br.BlockType == idBlock[j].br.BlockType &&
				idBlock[i].br.Identifier == idBlock[j].br.Identifier &&
				idBlock[i].br.StartToken.Int32 == idBlock[j].br.StartToken.Int32 &&
				idBlock[i].br.EndToken.Int32 == idBlock[j].br.EndToken.Int32 {
				// If collision is on the same side, then ignore it
				// (it's probably not our fault and we hope its okay...)
				if idBlock[i].side == idBlock[j].side {
					continue
				}

				var first, second *model.UserMarkBlockRange
				if idBlock[i].side == LeftSide {
					first = left[idBlock[i].br.UserMarkID]
					second = right[idBlock[j].br.UserMarkID]
				} else {
					first = left[idBlock[j].br.UserMarkID]
					second = right[idBlock[i].br.UserMarkID]
				}

				// Last check before calling it a duplicate: Are UserMark equal?
				if !first.UserMark.Equals(second.UserMark) {
					continue
				}

				var conflictKey strings.Builder
				// Use UnixNano as a monotonically increasing number, so we
				// are able to apply conflict solutions in the right order later
				conflictKey.WriteString(fmt.Sprint(time.Now().UnixNano()))
				conflictKey.WriteString("_")
				conflictKey.WriteString(first.UniqueKey())
				conflictKey.WriteString("_")
				conflictKey.WriteString(second.UniqueKey())
				conflicts[conflictKey.String()] = MergeConflict{first, second}
				idBlock[j] = brFrom{}
			}
		}
	}

	idBlock = sortBRFroms(idBlock)
	return idBlock, conflicts
}

// sortBRFroms returns a sorted slice of brFrom entries according to their
// startToken. If a entry is empty, it gets removed
func sortBRFroms(entries []brFrom) []brFrom {
	sort.SliceStable(entries, func(i, j int) bool {
		// Consider emptry entries as infinity
		if entries[i] == (brFrom{}) {
			return false
		}
		if entries[j] == (brFrom{}) {
			return true
		}

		return entries[i].br.StartToken.Int32 < entries[j].br.StartToken.Int32
	})

	// Remove all empty entries
	emptyEntries := 0
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i] != (brFrom{}) {
			break
		}
		emptyEntries++
	}

	return entries[:len(entries)-emptyEntries]
}

// replaceUMBRConflictsWithSolution removes conflicting entries on the left and right
// and replaces one of them with the given solution in conflictSolution. IDChanges
// and its inverted counterpart indicate possible changes of IDs.
func replaceUMBRConflictsWithSolution(left *[]*model.UserMarkBlockRange, right *[]*model.UserMarkBlockRange, conflictSolution map[string]MergeSolution) (IDChanges, IDChanges) {
	changes := IDChanges{
		Left:  map[int]int{},
		Right: map[int]int{},
	}
	invertedChanges := IDChanges{
		Left:  map[int]int{},
		Right: map[int]int{},
	}

	// We need to go through solutions in the order they were added. Otherwise
	// it can happen that we add entry back, after it had already been removed
	sortedKeys := sortedSolutionKeys(conflictSolution)

	for _, key := range sortedKeys {
		sol := conflictSolution[key]
		var side, other *[]*model.UserMarkBlockRange

		// Ignore non UserMarkBlockRange solutions
		if _, ok := sol.Solution.(*model.UserMarkBlockRange); !ok {
			continue
		}
		if _, ok := sol.Discarded.(*model.UserMarkBlockRange); !ok {
			continue
		}

		if sol.Side == LeftSide {
			side = left
			other = right
			changes.Right[sol.Discarded.ID()] = sol.Solution.ID()
			invertedChanges.Right[sol.Solution.ID()] = sol.Discarded.ID()
		} else {
			side = right
			other = left
			changes.Left[sol.Discarded.ID()] = sol.Solution.ID()
			invertedChanges.Left[sol.Solution.ID()] = sol.Discarded.ID()
		}

		(*side)[sol.Solution.ID()] = (sol.Solution).(*model.UserMarkBlockRange)
		(*other)[sol.Discarded.ID()] = nil
	}

	return changes, invertedChanges
}

var extractKeyNumber = regexp.MustCompile(`^(\d*)`)

// sortedSolutionKeys returns a list of keys from the conflictSolution map,
// where the keys are sorted by the first number-part of a key (i.e.
// 12345_uniqueKey would be sorted according to 12345).
func sortedSolutionKeys(conflictSolution map[string]MergeSolution) []string {
	orderedKeys := make([]struct {
		number int64
		key    string
	}, len(conflictSolution))

	// Extract first part of key, which is a monotonically increasing number,
	// afterwards sort it. Ignore errors and in these cases just set the number
	// to 0.
	i := 0
	for key := range conflictSolution {
		match := extractKeyNumber.FindStringSubmatch(key)
		number, _ := strconv.ParseInt(match[0], 0, 64)
		orderedKeys[i] = struct {
			number int64
			key    string
		}{
			number: number,
			key:    key,
		}
		i++
	}
	sort.Slice(orderedKeys, func(i int, j int) bool {
		return orderedKeys[i].number < orderedKeys[j].number
	})

	result := make([]string, len(orderedKeys))
	for i, ordered := range orderedKeys {
		result[i] = ordered.key
	}

	return result
}

// ingestUMBR ingest UserMarks & BlockRanges in a Map[LocationID]map[Identifier][]*model.BlockRange
func ingestUMBR(left []*model.UserMarkBlockRange, right []*model.UserMarkBlockRange) map[int]map[int][]brFrom {
	result := make(map[int]map[int][]brFrom, estimateLocationCount(left, right))
	for _, side := range []*umbrFrom{{LeftSide, left}, {RightSide, right}} {
		for _, umbr := range side.umbr {
			if umbr == nil {
				continue
			}
			if _, ok := result[umbr.UserMark.LocationID]; !ok {
				result[umbr.UserMark.LocationID] = map[int][]brFrom{}
			}
			for _, br := range umbr.BlockRanges {
				if _, ok := result[umbr.UserMark.LocationID][br.Identifier]; !ok {
					result[umbr.UserMark.LocationID][br.Identifier] = []brFrom{}
				}
				result[umbr.UserMark.LocationID][br.Identifier] = append(result[umbr.UserMark.LocationID][br.Identifier], brFrom{side.side, br})
			}
		}
	}

	return result
}

// joinToUserMarkBlockRange joins entries of UserMark and BlockRange together creating
// a slice of UserMarkBlockRange for which the index corresponds to the UserMarkID.
// It expects the IDs of UserMark and BlockRange to correspond to their
// index within in um and br! If a UserMark doesn't have BlockRanges, it will
// be removed from the result.
func joinToUserMarkBlockRange(um []*model.UserMark, br []*model.BlockRange) []*model.UserMarkBlockRange {
	result := make([]*model.UserMarkBlockRange, len(um))
	for i, entry := range um {
		if entry == nil || *entry == (model.UserMark{}) {
			continue
		}
		result[i] = &model.UserMarkBlockRange{UserMark: model.MakeModelCopy(entry).(*model.UserMark), BlockRanges: []*model.BlockRange{}}
	}

	for _, entry := range br {
		if entry == nil || entry.UserMarkID >= len(result) || result[entry.UserMarkID] == nil {
			continue
		}
		result[entry.UserMarkID].BlockRanges = append(result[entry.UserMarkID].BlockRanges, model.MakeModelCopy(entry).(*model.BlockRange))
	}

	// Remove UserMarks that don't have BlockRanges attached to them (e.g. "empty" markings)
	for i, umbr := range result {
		if umbr == nil {
			continue
		}
		if len(umbr.BlockRanges) == 0 {
			result[i] = nil
		}
	}

	return result
}

// splitUserMarkBlockRange splits a UserMarkBlockRange into separate UserMark and BlockRange slices.
// It expects the UserMark IDs of BlockRanges within a UserMarkBlockRange to not be correct yet, so it will
// update them according to their UserMark. It will also update the IDs of all entries to be in sync
// with the index of the slices.
func splitUserMarkBlockRange(umbrs []*model.UserMarkBlockRange) ([]*model.UserMark, []*model.BlockRange) {
	if len(umbrs) == 0 {
		return []*model.UserMark{}, []*model.BlockRange{}
	}

	userMarks := make([]*model.UserMark, len(umbrs))
	blockRanges := make([]*model.BlockRange, 1, len(umbrs))
	blockRanges[0] = nil

	brIndex := 1
	for i, umbr := range umbrs {
		if umbr == nil {
			continue
		}

		userMarks[i] = umbr.UserMark
		for _, br := range umbr.BlockRanges {
			if br == nil || *br == (model.BlockRange{}) {
				continue
			}
			br.UserMarkID = umbr.UserMark.UserMarkID
			br.SetID(brIndex)
			blockRanges = append(blockRanges, br)
			brIndex++
		}
	}

	return userMarks, blockRanges
}

// estimateLocationCount tries to estimate the number of locations of the left
// and right side combined by looking at the highest LocationID.
func estimateLocationCount(left []*model.UserMarkBlockRange, right []*model.UserMarkBlockRange) int {
	maxLocationID := 0
	for _, side := range [][]*model.UserMarkBlockRange{left, right} {
		for i := len(side) - 1; i >= 0; i-- {
			if side[i] == nil || side[i].UserMark.LocationID == 0 {
				continue
			}
			maxLocationID += side[i].UserMark.LocationID
			break
		}
	}
	return maxLocationID
}
