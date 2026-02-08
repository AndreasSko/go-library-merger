package cmd

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/AndreasSko/go-jwlm/merger"
	"github.com/AndreasSko/go-jwlm/model"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
)

// removeDuplicatesCmd represents the remove-duplicates command
var removeDuplicatesCmd = &cobra.Command{
	Use:   "remove-duplicates <backup-file> <output-file>",
	Short: "Remove duplicate notes from a JW Library backup file",
	Long: `remove-duplicates scans a .jwlibrary backup file, identifies
duplicate notes, and allows you to interactively choose which duplicates
to keep and which to remove. The cleaned backup is saved to the output file.`,
	Example: `go-jwlm remove-duplicates backup.jwlibrary cleaned.jwlibrary`,
	RunE: func(cmd *cobra.Command, args []string) error {
		backupFilename := args[0]
		outputFilename := args[1]
		return removeDuplicates(backupFilename, outputFilename, terminal.Stdio{In: os.Stdin, Out: os.Stdout, Err: os.Stderr})
	},
	Args: cobra.ExactArgs(2),
}

func removeDuplicates(backupFilename string, outputFilename string, stdio terminal.Stdio) error {
	fmt.Fprintln(stdio.Out, "Importing backup file")
	db := model.Database{
		SkipPlaylists: true,
	}
	err := db.ImportJWLBackup(backupFilename)
	if err != nil {
		return fmt.Errorf("failed to import backup: %w", err)
	}

	// Count initial notes
	initialNoteCount := 0
	for _, note := range db.Note {
		if note != nil {
			initialNoteCount++
		}
	}
	fmt.Fprintf(stdio.Out, "📊 Total notes: %d\n", initialNoteCount)

	fmt.Fprintln(stdio.Out, "🔍 Scanning for duplicate notes...")
	duplicateGroups := merger.DetectDuplicateNotes(db.Note)

	if len(duplicateGroups) == 0 {
		fmt.Fprintln(stdio.Out, "✅ No duplicate notes found!")
		fmt.Fprintln(stdio.Out, "Exporting backup to output file (unchanged)")
		if err = db.ExportJWLBackup(outputFilename); err != nil {
			return fmt.Errorf("failed to export backup: %w", err)
		}
		return nil
	}

	fmt.Fprintf(stdio.Out, "\n⚠️  Found %d group(s) of duplicate notes\n", len(duplicateGroups))
	fmt.Fprintln(stdio.Out, "You will be asked to select which note to keep from each group.\n")

	notesToRemove := make(map[int]bool)

	for i, group := range duplicateGroups {
		fmt.Fprintf(stdio.Out, "Duplicate Group #%d (%d notes):\n", i+1, len(group.Notes))

		t := table.NewWriter()
		t.SetStyle(table.StyleRounded)
		t.Style().Options = table.Options{
			DrawBorder:      true,
			SeparateColumns: true,
			SeparateHeader:  true,
			SeparateRows:    true,
		}
		t.SetOutputMirror(stdio.Out)

		options := []string{"Keep all (skip this group)"}
		for j, note := range group.Notes {
			title := ""
			if note.Title.Valid {
				title = note.Title.String
			}

			t.AppendRow(table.Row{
				fmt.Sprintf("Option %d", j+1),
				note.PrettyPrint(&db),
			})

			optionLabel := fmt.Sprintf("Option %d: ID=%d, Created=%s", j+1, note.NoteID, note.Created)
			if title != "" {
				optionLabel += fmt.Sprintf(", Title='%s'", title)
			}
			options = append(options, optionLabel)
		}

		t.Render()
		fmt.Fprint(stdio.Out, "\n")

		prompt := &survey.Select{
			Message: "Which note would you like to KEEP?",
			Options: options,
			Help:    "Select the note you want to keep. All other notes in this group will be removed. Choose 'Keep all' to skip this group.",
		}

		var selectedIndex int
		err := survey.AskOne(prompt, &selectedIndex, survey.WithStdio(stdio.In, stdio.Out, stdio.Err))
		if err == terminal.InterruptErr {
			fmt.Fprintln(stdio.Out, "interrupted")
			return nil
		} else if err != nil {
			return fmt.Errorf("failed to get user input: %w", err)
		}

		// If "Keep all" was selected (index 0), skip this group
		if selectedIndex == 0 {
			fmt.Fprintln(stdio.Out, "✓ Keeping all notes in this group\n")
			continue
		}

		// Adjust index to account for "Keep all" option
		actualNoteIndex := selectedIndex - 1

		// Mark all notes except the selected one for removal
		for j, note := range group.Notes {
			if j != actualNoteIndex {
				notesToRemove[note.NoteID] = true
			}
		}

		fmt.Fprintf(stdio.Out, "✓ Keeping option %d\n\n", actualNoteIndex+1)
	}

	// Remove the duplicates and rebuild the Note array with proper indexing
	fmt.Fprintln(stdio.Out, "🗑  Removing duplicate notes...")

	// Collect notes to keep
	notesToKeep := []*model.Note{}
	removedCount := 0
	for _, note := range db.Note {
		if note == nil {
			continue
		}
		if !notesToRemove[note.NoteID] {
			notesToKeep = append(notesToKeep, note)
		} else {
			removedCount++
		}
	}

	// Track ID changes for updating references
	noteIDChanges := make(map[int]int)

	// Rebuild the Note array with proper ID indexing (starts at 1)
	filteredNotes := make([]*model.Note, len(notesToKeep)+1)
	for i, note := range notesToKeep {
		oldID := note.NoteID
		newID := i + 1
		if oldID != newID {
			noteIDChanges[oldID] = newID
		}
		note.NoteID = newID
		filteredNotes[newID] = note
	}

	db.Note = filteredNotes

	// Count final notes
	finalNoteCount := 0
	for _, note := range db.Note {
		if note != nil {
			finalNoteCount++
		}
	}
	fmt.Fprintf(stdio.Out, "📊 Notes after cleanup: %d\n", finalNoteCount)

	// Update TagMap references to the new Note IDs
	if len(noteIDChanges) > 0 {
		fmt.Fprintln(stdio.Out, "🔄 Updating Note ID references in TagMaps...")
		for _, tagMap := range db.TagMap {
			if tagMap == nil || !tagMap.NoteID.Valid {
				continue
			}
			oldNoteID := int(tagMap.NoteID.Int32)
			if newNoteID, changed := noteIDChanges[oldNoteID]; changed {
				tagMap.NoteID.Int32 = int32(newNoteID)
			}
		}
	}

	fmt.Fprintf(stdio.Out, "✅ Removed %d duplicate note(s)\n", removedCount)

	fmt.Fprintln(stdio.Out, "Exporting cleaned backup")

	if err = db.ExportJWLBackup(outputFilename); err != nil {
		return fmt.Errorf("failed to export backup: %w", err)
	}

	fmt.Fprintf(stdio.Out, "✅ Successfully saved cleaned backup to: %s\n", outputFilename)
	return nil
}

func init() {
	rootCmd.AddCommand(removeDuplicatesCmd)
}
