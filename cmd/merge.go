package cmd

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/AndreasSko/go-library-merger/merger"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/buger/goterm"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
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

// PlaylistResolver represents a resolver that should be used for conflicting playlist items
var PlaylistResolver string

// SkipPlaylists indicates if playlists should be skipped when importing and merging backups.
var SkipPlaylists bool

func merge(leftFilename string, rightFilename string, mergedFilename string, stdio terminal.Stdio) (resultErr error) {
	if PlaylistResolver != "" && PlaylistResolver != "chooseLeft" && PlaylistResolver != "chooseRight" {
		return fmt.Errorf("playlist resolver must be chooseLeft or chooseRight, got %q", PlaylistResolver)
	}
	fmt.Fprintln(stdio.Out, "Importing left backup")
	left := model.Database{
		SkipPlaylists: SkipPlaylists,
	}
	err := left.ImportJWLBackup(leftFilename)
	if err != nil {
		return fmt.Errorf("failed to import left backup: %w", err)
	}

	defer func() { resultErr = errors.Join(resultErr, left.Close()) }()
	fmt.Fprintln(stdio.Out, "Importing right backup")
	right := model.Database{
		SkipPlaylists: SkipPlaylists,
	}
	err = right.ImportJWLBackup(rightFilename)
	if err != nil {
		return fmt.Errorf("failed to import right backup: %w", err)
	}

	defer func() { resultErr = errors.Join(resultErr, right.Close()) }()
	fmt.Fprintln(stdio.Out, "⌛ Preparing Databases")
	merger.PrepareDatabasesPreMerge(&left, &right)

	merged := model.Database{}
	defer func() { resultErr = errors.Join(resultErr, merged.Close()) }()

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
	merger.UpdateLRIDs(left.PlaylistItemLocationMap, right.PlaylistItemLocationMap, "LocationID", locationIDChanges)
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
					return resErr
				}
				addToSolutions(bookmarksConflictSolution, newSolutions)
			} else {
				newSolutions, selectionErr := handleMergeConflict(err.Conflicts, &merged, stdio)
				if selectionErr != nil {
					return selectionErr
				}
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
					return resErr
				}
				addToSolutions(inputFieldConflictSolution, newSolutions)
			} else {
				newSolutions, selectionErr := handleMergeConflict(err.Conflicts, &merged, stdio)
				if selectionErr != nil {
					return selectionErr
				}
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
			resolved, selectionErr := handleMergeConflict(err.Conflicts, nil, stdio)
			if selectionErr != nil {
				return selectionErr
			}
			tagsConflictSolution = resolved
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
					return resErr
				}
				addToSolutions(UMBRConflictSolution, newSolutions)
			} else {
				newSolutions, selectionErr := handleMergeConflict(err.Conflicts, &merged, stdio)
				if selectionErr != nil {
					return selectionErr
				}
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
					return resErr
				}
				addToSolutions(notesConflictSolution, newSolutions)
			} else {
				newSolutions, selectionErr := handleMergeConflict(err.Conflicts, &merged, stdio)
				if selectionErr != nil {
					return selectionErr
				}
				addToSolutions(notesConflictSolution, newSolutions)
			}
		default:
			return fmt.Errorf("failed to merge notes: %w", err)
		}
	}
	fmt.Fprintln(stdio.Out, "Done.")

	if _, err := fmt.Fprintln(stdio.Out, "🎵 Merging Playlists"); err != nil {
		return err
	}
	playlistSolutions := map[string]merger.MergeSolution{}
	for {
		err := merger.MergePlaylists(&left, &right, &merged, playlistSolutions)
		if err == nil {
			break
		}
		if conflict, ok := err.(merger.MergeConflictError); ok {
			if PlaylistResolver == "" {
				resolved, selectionErr := handleMergeConflict(conflict.Conflicts, &merged, stdio)
				if selectionErr != nil {
					return selectionErr
				}
				addToSolutions(playlistSolutions, resolved)
			} else {
				resolved, err := merger.AutoResolveConflicts(conflict.Conflicts, PlaylistResolver)
				if err != nil {
					return fmt.Errorf("failed to resolve playlist settings: %w", err)
				}
				addToSolutions(playlistSolutions, resolved)
			}
		} else {
			return fmt.Errorf("failed to merge playlists: %w", err)
		}
	}
	if _, err := fmt.Fprintln(stdio.Out, "Done."); err != nil {
		return err
	}

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
			resolved, selectionErr := handleMergeConflict(err.Conflicts, nil, stdio)
			if selectionErr != nil {
				return selectionErr
			}
			tagMapsConflictSolution = resolved
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

func handleMergeConflict(conflicts map[string]merger.MergeConflict, mergedDB *model.Database, stdio terminal.Stdio) (map[string]merger.MergeSolution, error) {
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
		if context := playlistConflictContext(conflict, mergedDB); context != "" {
			if _, err := fmt.Fprintln(stdio.Out, context); err != nil {
				return nil, err
			}
		}
		t := table.NewWriter()
		t.SetStyle(table.StyleRounded)
		t.Style().Options = table.Options{
			DrawBorder:      true,
			SeparateColumns: true,
			SeparateFooter:  true,
			SeparateHeader:  true,
			SeparateRows:    true,
		}

		t.SetOutputMirror(stdio.Out)
		leftText, rightText := playlistConflictText(conflict, mergedDB)
		if goterm.Width() >= 190 {
			t.AppendHeader(table.Row{"Left", "Right"})
			t.AppendRow([]interface{}{leftText, rightText})
		} else {
			t.AppendRows([]table.Row{{"Left"}, {leftText}, {"Right"}, {rightText}})
		}

		t.Render()

		if _, err := fmt.Fprint(stdio.Out, "\n\n"); err != nil {
			return nil, err
		}

		var selected string
		err := survey.AskOne(prompt, &selected, survey.WithStdio(stdio.In, stdio.Out, stdio.Err))
		if err != nil {
			return nil, fmt.Errorf("conflict selection canceled or unavailable: %w", err)
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

	return result, nil
}

func playlistConflictText(conflict merger.MergeConflict, db *model.Database) (string, string) {
	fields := model.PlaylistChangedFields(conflict.Left, conflict.Right)
	if len(fields) == 0 {
		return conflict.Left.PrettyPrint(db), conflict.Right.PrettyPrint(db)
	}
	return model.PlaylistChangedFieldText(conflict.Left, fields), model.PlaylistChangedFieldText(conflict.Right, fields)
}

func playlistConflictContext(conflict merger.MergeConflict, db *model.Database) string {
	if len(model.PlaylistChangedFields(conflict.Left, conflict.Right)) == 0 {
		return ""
	}
	related := conflict.Left.RelatedEntries(db)
	lines := []string{}
	if related.Tag != nil {
		lines = append(lines, "Playlist: "+related.Tag.Name)
	}
	item := related.PlaylistItem
	if direct, ok := conflict.Left.(*model.PlaylistItem); ok {
		item = direct
	}
	// A changing label belongs in the comparison; only an unchanged label is
	// common context. Mapping conflicts already relate to the chosen item.
	labelChanged := false
	for _, field := range model.PlaylistChangedFields(conflict.Left, conflict.Right) {
		labelChanged = labelChanged || field == "label"
	}
	if item != nil && item.Label != "" && !labelChanged {
		lines = append(lines, "Item: "+item.Label)
	}
	if related.IndependentMedia != nil {
		name := related.IndependentMedia.OriginalFilename
		if name == "" {
			name = related.IndependentMedia.FilePath
		}
		lines = append(lines, "Source: "+name)
	} else if related.Location != nil {
		name := related.Location.Title.String
		if name == "" {
			name = related.Location.KeySymbol.String
		}
		if name != "" {
			lines = append(lines, "Source: "+name)
		}
	}
	return strings.Join(lines, "\n")
}

func init() {
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().StringVar(&BookmarkResolver, "bookmarks", "", "Resolve conflicting bookmarks with resolver (can be 'chooseLeft' or 'chooseRight')")
	mergeCmd.Flags().StringVar(&MarkingResolver, "markings", "", "Resolve conflicting markings with resolver (can be 'chooseLeft' or 'chooseRight')")
	mergeCmd.Flags().StringVar(&NoteResolver, "notes", "", "Resolve conflicting notes with resolver (can be 'chooseNewest', 'chooseLeft', or 'chooseRight')")
	mergeCmd.Flags().StringVar(&InputFieldResolver, "inputFields", "", "Resolve conflicting inputFields with resolver (can be 'chooseLeft', or 'chooseRight')")
	mergeCmd.Flags().StringVar(&PlaylistResolver, "playlists", "", "Resolve conflicting playlist items with resolver (can be 'chooseLeft' or 'chooseRight')")
	mergeCmd.Flags().BoolVar(&SkipPlaylists, "skipPlaylists", false, "Skip playlists when importing and merging backups.")
}
