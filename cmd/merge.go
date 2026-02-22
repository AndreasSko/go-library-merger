package cmd

import (
	"fmt"
	"os"
	"reflect"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/AndreasSko/go-library-merger/merger"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/buger/goterm"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Use:   "merge <left-backup> <right-backup> <dest-filename>",
	Short: "Merge two JW Library backup files",
	Long: `merge imports the left and right .jwlibrary backup file, merges them and 
exports it to the destination file. If a collision between the left and 
the right backup is detected, the user is asked to choose which side should
be included in the merged backup. You are able to let the merger 
automatically solve conflicts using the 'chooseLeft', 'chooseRight', and 
'chooseNewest' resolvers (see Flags).`,
	Example: `library-merger merge left.jwlibrary right.jwlibrary merged.jwlibrary
library-merger merge left.jwlibrary right.jwlibrary merged.jwlibrary --bookmarks chooseLeft --markings chooseRight --notes chooseNewest --inputFields chooseRight`,
	RunE: func(cmd *cobra.Command, args []string) error {
		leftFilename := args[0]
		rightFilename := args[1]
		mergedFilename := args[2]
		return merge(leftFilename, rightFilename, mergedFilename, terminal.Stdio{In: os.Stdin, Out: os.Stdout, Err: os.Stderr})
	},
	Args: cobra.ExactArgs(3),
}

// BookmarkResolver represents a resolver that should be used for conflicting Bookmarks
var BookmarkResolver string

// MarkingResolver represents a resolver that should be used for conflicting UserMarkBlockRanges
var MarkingResolver string

// NoteResolver represents a resolver that should be used for conflicting Notes
var NoteResolver string

// InputFieldResolver represents a resolver that should be used for conflicting InputFields
var InputFieldResolver string

// SkipPlaylists indicates if playlists should be skipped when importing backups.
// It is meant as a temporary workaround until merging of playlists is implemented.
var SkipPlaylists bool

func merge(leftFilename string, rightFilename string, mergedFilename string, stdio terminal.Stdio) error {
	fmt.Fprintln(stdio.Out, "Importing left backup")
	left := model.Database{
		SkipPlaylists: SkipPlaylists,
	}
	err := left.ImportJWLBackup(leftFilename)
	if err != nil {
		return fmt.Errorf("failed to import left backup: %w", err)
	}

	fmt.Fprintln(stdio.Out, "Importing right backup")
	right := model.Database{
		SkipPlaylists: SkipPlaylists,
	}
	err = right.ImportJWLBackup(rightFilename)
	if err != nil {
		return fmt.Errorf("failed to import right backup: %w", err)
	}

	fmt.Fprintln(stdio.Out, "⌛ Preparing Databases")
	merger.PrepareDatabasesPreMerge(&left, &right)

	merged := model.Database{}

	fmt.Fprintln(stdio.Out, "🧭 Merging Locations")
	mergedLocations, locationIDChanges, err := merger.MergeLocations(left.Location, right.Location)
	if err != nil {
		return fmt.Errorf("failed to merge locations: %w", err)
	}
	merged.Location = mergedLocations
	merger.UpdateLRIDs(left.Bookmark, right.Bookmark, "LocationID", locationIDChanges)
	merger.UpdateLRIDs(left.Bookmark, right.Bookmark, "PublicationLocationID", locationIDChanges)
	merger.UpdateLRIDs(left.InputField, right.InputField, "LocationID", locationIDChanges)
	merger.UpdateLRIDs(left.Note, right.Note, "LocationID", locationIDChanges)
	merger.UpdateLRIDs(left.TagMap, right.TagMap, "LocationID", locationIDChanges)
	merger.UpdateLRIDs(left.UserMark, right.UserMark, "LocationID", locationIDChanges)
	fmt.Fprintln(stdio.Out, "Done.")

	fmt.Fprintln(stdio.Out, "📑 Merging Bookmarks")
	bookmarksConflictSolution := map[string]merger.MergeSolution{}
	for {
		mergedBookmarks, _, err := merger.MergeBookmarks(left.Bookmark, right.Bookmark, bookmarksConflictSolution)
		if err == nil {
			merged.Bookmark = mergedBookmarks
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			if BookmarkResolver != "" {
				var resErr error
				newSolutions, resErr := merger.AutoResolveConflicts(err.Conflicts, BookmarkResolver)
				if resErr != nil {
					log.Fatal(resErr)
				}
				addToSolutions(bookmarksConflictSolution, newSolutions)
			} else {
				newSolutions := handleMergeConflict(err.Conflicts, &merged, stdio)
				addToSolutions(bookmarksConflictSolution, newSolutions)
			}
		default:
			return fmt.Errorf("failed to merge bookmarks: %w", err)
		}
	}
	fmt.Fprintln(stdio.Out, "Done.")

	fmt.Fprintln(stdio.Out, "✍️  Merging InputFields")
	inputFieldConflictSolution := map[string]merger.MergeSolution{}
	for {
		mergedInputFields, _, err := merger.MergeInputFields(left.InputField, right.InputField, inputFieldConflictSolution)
		if err == nil {
			merged.InputField = mergedInputFields
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			if InputFieldResolver != "" {
				var resErr error
				newSolutions, resErr := merger.AutoResolveConflicts(err.Conflicts, InputFieldResolver)
				if resErr != nil {
					log.Fatal(resErr)
				}
				addToSolutions(inputFieldConflictSolution, newSolutions)
			} else {
				newSolutions := handleMergeConflict(err.Conflicts, &merged, stdio)
				addToSolutions(inputFieldConflictSolution, newSolutions)
			}
		default:
			return fmt.Errorf("failed to merge inputFields: %w", err)
		}
	}
	fmt.Fprintln(stdio.Out, "Done.")

	fmt.Fprintln(stdio.Out, "🏷  Merging Tags")
	var tagsConflictSolution map[string]merger.MergeSolution
	for {
		mergedTags, tagIDChanges, err := merger.MergeTags(left.Tag, right.Tag, tagsConflictSolution)
		if err == nil {
			merged.Tag = mergedTags
			merger.UpdateLRIDs(left.TagMap, right.TagMap, "TagID", tagIDChanges)
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			tagsConflictSolution = handleMergeConflict(err.Conflicts, nil, stdio) // TODO
		default:
			return fmt.Errorf("failed to merge tags: %w", err)
		}
	}
	fmt.Fprintln(stdio.Out, "Done.")

	fmt.Fprintln(stdio.Out, "🖍  Merging Markings")
	UMBRConflictSolution := map[string]merger.MergeSolution{}
	for {
		mergedUserMarks, mergedBlockRanges, userMarkIDChanges, err := merger.MergeUserMarkAndBlockRange(left.UserMark, left.BlockRange, right.UserMark, right.BlockRange, UMBRConflictSolution)
		if err == nil {
			merged.UserMark = mergedUserMarks
			merged.BlockRange = mergedBlockRanges
			merger.UpdateLRIDs(left.Note, right.Note, "UserMarkID", userMarkIDChanges)
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			if MarkingResolver != "" {
				var resErr error
				newSolutions, resErr := merger.AutoResolveConflicts(err.Conflicts, MarkingResolver)
				if resErr != nil {
					log.Fatal(resErr)
				}
				addToSolutions(UMBRConflictSolution, newSolutions)
			} else {
				newSolutions := handleMergeConflict(err.Conflicts, &merged, stdio)
				addToSolutions(UMBRConflictSolution, newSolutions)
			}
		default:
			return fmt.Errorf("failed to merge markings: %w", err)
		}
	}
	fmt.Fprintln(stdio.Out, "Done.")

	fmt.Fprintln(stdio.Out, "📝 Merging Notes")
	notesConflictSolution := map[string]merger.MergeSolution{}
	for {
		mergedNotes, notesIDChanges, err := merger.MergeNotes(left.Note, right.Note, notesConflictSolution)
		if err == nil {
			merged.Note = mergedNotes
			merger.UpdateLRIDs(left.TagMap, right.TagMap, "NoteID", notesIDChanges)
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			if NoteResolver != "" {
				var resErr error
				newSolutions, resErr := merger.AutoResolveConflicts(err.Conflicts, NoteResolver)
				if resErr != nil {
					log.Fatal(resErr)
				}
				addToSolutions(notesConflictSolution, newSolutions)
			} else {
				newSolutions := handleMergeConflict(err.Conflicts, &merged, stdio)
				addToSolutions(notesConflictSolution, newSolutions)
			}
		default:
			return fmt.Errorf("failed to merge notes: %w", err)
		}
	}
	fmt.Fprintln(stdio.Out, "Done.")

	fmt.Fprintln(stdio.Out, "🏷  Merging TagMaps")
	var tagMapsConflictSolution map[string]merger.MergeSolution
	for {
		mergedTagMaps, _, err := merger.MergeTagMaps(left.TagMap, right.TagMap, tagMapsConflictSolution)
		if err == nil {
			merged.TagMap = mergedTagMaps
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			tagMapsConflictSolution = handleMergeConflict(err.Conflicts, nil, stdio)
		default:
			return fmt.Errorf("failed to merge tagMaps: %w", err)
		}
	}
	fmt.Fprintln(stdio.Out, "Done.")

	fmt.Fprintln(stdio.Out, "🎉 Finished merging!")

	fmt.Fprintln(stdio.Out, "⌛ Preparing merged database for exporting")
	if err := merger.PrepareDatabasesPostMerge(&merged); err != nil {
		return fmt.Errorf("failed to prepare database after merging: %w", err)
	}

	fmt.Fprintln(stdio.Out, "Exporting merged database")
	if err = merged.ExportJWLBackup(mergedFilename); err != nil {
		return fmt.Errorf("failed to export backup: %w", err)
	}

	return nil
}

// addToSolutions adds new mergeSolutions to the existing map of mergeSolutions
func addToSolutions(solutions map[string]merger.MergeSolution, new map[string]merger.MergeSolution) {
	for key, value := range new {
		solutions[key] = value
	}
}

func handleMergeConflict(conflicts map[string]merger.MergeConflict, mergedDB *model.Database, stdio terminal.Stdio) map[string]merger.MergeSolution {
	helpText := ""
	for _, val := range conflicts {
		helpText = mergeConflictHelp(reflect.TypeOf(val.Left).String())
		break
	}

	prompt := &survey.Select{
		Message: "Select which side should be chosen:",
		Options: []string{"Left", "Right"},
		Help:    helpText,
	}

	result := make(map[string]merger.MergeSolution, len(conflicts))
	for key, conflict := range conflicts {
		t := table.NewWriter()
		t.SetStyle(table.StyleRounded)
		t.Style().Options = table.Options{
			DrawBorder:      true,
			SeparateColumns: true,
			SeparateFooter:  true,
			SeparateHeader:  true,
			SeparateRows:    true,
		}

		t.SetOutputMirror(os.Stdout)
		if goterm.Width() >= 190 {
			t.AppendHeader(table.Row{"Left", "Right"})
			t.AppendRow([]interface{}{conflict.Left.PrettyPrint(mergedDB), conflict.Right.PrettyPrint(mergedDB)})
		} else {
			t.AppendRows([]table.Row{{"Left"}, {conflict.Left.PrettyPrint(mergedDB)}, {"Right"}, {conflict.Right.PrettyPrint(mergedDB)}})
		}

		t.Render()

		fmt.Fprint(stdio.Out, "\n\n")

		var selected string
		err := survey.AskOne(prompt, &selected, survey.WithStdio(stdio.In, stdio.Out, stdio.Err))
		if err == terminal.InterruptErr {
			fmt.Fprintln(stdio.Out, "interrupted")
			os.Exit(0)
		} else if err != nil {
			panic(err)
		}

		if selected == "Left" {
			result[key] = merger.MergeSolution{
				Side:      merger.LeftSide,
				Solution:  conflict.Left,
				Discarded: conflict.Right,
			}
		} else {
			result[key] = merger.MergeSolution{
				Side:      merger.RightSide,
				Solution:  conflict.Right,
				Discarded: conflict.Left,
			}
		}
	}

	return result
}

func init() {
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().StringVar(&BookmarkResolver, "bookmarks", "", "Resolve conflicting bookmarks with resolver (can be 'chooseLeft' or 'chooseRight')")
	mergeCmd.Flags().StringVar(&MarkingResolver, "markings", "", "Resolve conflicting markings with resolver (can be 'chooseLeft' or 'chooseRight')")
	mergeCmd.Flags().StringVar(&NoteResolver, "notes", "", "Resolve conflicting notes with resolver (can be 'chooseNewest', 'chooseLeft', or 'chooseRight')")
	mergeCmd.Flags().StringVar(&InputFieldResolver, "inputFields", "", "Resolve conflicting inputFields with resolver (can be 'chooseLeft', or 'chooseRight')")
	mergeCmd.Flags().BoolVar(&SkipPlaylists, "skipPlaylists", false, "Skip playlists when importing backups. It is meant as a temporary workaround until merging of playlists is implemented.")
}
