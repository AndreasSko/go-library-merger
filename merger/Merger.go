package merger

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/AndreasSko/go-library-merger/model"
)

// MergeSolution indicates wheter a entry came from the left or right
// side of a to-be-merged model slice pair.
type MergeSolution struct {
	Side      MergeSide
	Solution  model.Model
	Discarded model.Model
}

// MergeSide indicates the side of a merge
type MergeSide string

const (
	// LeftSide is the left side of a merge
	LeftSide MergeSide = "leftSide"
	// RightSide is the right side of a merge
	RightSide MergeSide = "rightSide"
)

// MergeConflict represents two Models that collide because of the same
// UniqueKey or other similarities.
type MergeConflict struct {
	Left  model.Model
	Right model.Model
}

// MergeConflictError indicates that a conflict happened while trying to merge
// two slices of Model. It contains the conflicts in order for the caller to solve them.
type MergeConflictError struct {
	Err       string
	Conflicts map[string]MergeConflict
}

func (e MergeConflictError) Error() string {
	return fmt.Sprintf("There were conflicts while trying to merge: %s", e.Conflicts)
}

// merge merges a left and a right slice of structs implementing the Model interface.
// If there is a collision in the process, it returns an error asking for specification how it should handle it.
func merge(left interface{}, right interface{}, conflictSolution map[string]MergeSolution) (map[string]MergeSolution, error) {
	maxLen := 0
	if reflect.ValueOf(left).Len() > reflect.ValueOf(right).Len() {
		maxLen = reflect.ValueOf(left).Len()
	} else {
		maxLen = reflect.ValueOf(right).Len()
	}

	duplicateCheck := make(map[string]MergeSolution, reflect.ValueOf(left).Len()+reflect.ValueOf(right).Len())
	collisions := make(map[string]MergeConflict, maxLen)

	// First add all entries of the left slice
	switch reflect.TypeOf(left).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(left)
		for i := 0; i < s.Len(); i++ {
			// Make sure we don't have a nil-pointer
			if s.Index(i).IsNil() {
				continue
			}

			l := s.Index(i).Interface().(model.Model)
			duplicateCheck[l.UniqueKey()] = MergeSolution{Side: LeftSide, Solution: l}
		}
	}

	// Try to add entries of right side, if they don't conflict with existing ones
	switch reflect.TypeOf(right).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(right)
		for i := 0; i < s.Len(); i++ {
			// Make sure we don't have a nil-pointer
			if s.Index(i).IsNil() {
				continue
			}

			r := s.Index(i).Interface().(model.Model)
			if conflict, exists := duplicateCheck[r.UniqueKey()]; exists {
				if solution, ok := conflictSolution[r.UniqueKey()]; ok {
					duplicateCheck[r.UniqueKey()] = solution
				} else {
					collisions[r.UniqueKey()] = MergeConflict{
						Left:  conflict.Solution,
						Right: r,
					}
				}
			} else {
				duplicateCheck[r.UniqueKey()] = MergeSolution{Side: RightSide, Solution: r}
			}
		}
	}

	if len(collisions) != 0 {
		return duplicateCheck, MergeConflictError{
			Err:       "There were conflicts while trying to merge",
			Conflicts: collisions,
		}
	}

	return duplicateCheck, nil
}

// tryMergeWithConflictSolver is a generalized method for merging a left and a right
// slice of structs implementing the Model interface. It tries to solve possible
// conflicts using the given mergeConflictSolver and will return a mergeConflictError
// if it wasn't able to solve all conflicts on its own.
func tryMergeWithConflictSolver(left interface{}, right interface{}, conflictSolution map[string]MergeSolution, conflictSolver MergeConflictSolver) ([]model.Model, IDChanges, error) {
	var solutionMap map[string]MergeSolution
	var err error

	if conflictSolution == nil {
		conflictSolution = map[string]MergeSolution{}
	}

	// Try to merge with automatic conflic resolution until the number of conflicts
	// doesn't shrink anymore
	prevConflicts := 0
Loop:
	for {
		solutionMap, err = merge(left, right, conflictSolution)
		if err == nil {
			break
		}

		switch err := err.(type) {
		case MergeConflictError:
			// If we couldn't shrink number of conflicts in last iteration, break
			if prevConflicts == len(err.Conflicts) {
				break Loop
			}

			// merge automatic conflict solution with existing/given solution
			autoConflictSolution, _ := conflictSolver(err.Conflicts)
			for key, autoSol := range autoConflictSolution {
				if sol, exists := conflictSolution[key]; exists {
					return []model.Model{}, IDChanges{}, fmt.Errorf("One of the given conflictSolution is conflicting with the one generated automatically: given %s, automatic: %s", sol, autoSol)
				}
				conflictSolution[key] = autoSol
			}

			prevConflicts = len(err.Conflicts)
		default:
			return []model.Model{}, IDChanges{}, err
		}
	}

	if err != nil {
		return []model.Model{}, IDChanges{}, err
	}

	result, changes := prepareMergeSolution(&solutionMap)

	return result, changes, err
}

// prepareMergeSolution creates a sorted slice of the solutions given in the solutionMap
// and updates the IDs of the entries if necessary. IDChanges will track changed IDs.
func prepareMergeSolution(solutionMap *map[string]MergeSolution) ([]model.Model, IDChanges) {
	// Convert map to slice and sort it so we have a deterministic output
	solutionSlice := make([]MergeSolution, len(*solutionMap))
	i := 0
	for _, sol := range *solutionMap {
		solutionSlice[i] = sol
		i++
	}
	sortMergeSolution(&solutionSlice)

	result := make([]model.Model, len(*solutionMap)+1)
	changes := IDChanges{
		Left:  map[int]int{},
		Right: map[int]int{},
	}

	i = 1
	for _, sol := range solutionSlice {
		result[i] = model.MakeModelCopy(sol.Solution)
		// Update ID if needed
		if sol.Solution.ID() != i {
			if sol.Side == LeftSide {
				changes.Left[sol.Solution.ID()] = i
			} else {
				changes.Right[sol.Solution.ID()] = i
			}
			result[i].SetID(i)
		}

		// If we merged a duplicate, we also need to cope with
		// changing the ID of the other side
		if sol.Discarded != nil && sol.Discarded.ID() != i {
			if sol.Side == LeftSide {
				changes.Right[sol.Discarded.ID()] = i
			} else {
				changes.Left[sol.Discarded.ID()] = i
			}
		}
		i++
	}

	return result, changes
}

// sortMergeSolution sorts a slice of mergeSolution according to the ID of the model.
// If both IDs are equal, a solution being on the left side is considered greater
// than one on the right.
func sortMergeSolution(solution *[]MergeSolution) {
	sort.Slice(*solution, func(i, j int) bool {
		// If ID are the same, then leftSide > rightSide
		if (*solution)[i].Solution.ID() == (*solution)[j].Solution.ID() {
			return (*solution)[i].Side == LeftSide
		}
		return (*solution)[i].Solution.ID() < (*solution)[j].Solution.ID()
	})
}
