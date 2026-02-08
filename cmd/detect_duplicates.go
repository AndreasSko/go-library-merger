package cmd

import (
	"fmt"
	"os"

	"github.com/AndreasSko/go-jwlm/merger"
	"github.com/AndreasSko/go-jwlm/model"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
)

// detectDuplicatesCmd represents the detect-duplicates command
var detectDuplicatesCmd = &cobra.Command{
	Use:   "detect-duplicates <backup-file>",
	Short: "Detect duplicate notes in a JW Library backup file",
	Long: `detect-duplicates scans a .jwlibrary backup file and identifies
duplicate notes based on matching title and content. It displays all
duplicate groups found, showing details for each duplicate note.`,
	Example: `go-jwlm detect-duplicates backup.jwlibrary`,
	RunE: func(cmd *cobra.Command, args []string) error {
		backupFilename := args[0]
		return detectDuplicates(backupFilename)
	},
	Args: cobra.ExactArgs(1),
}

func detectDuplicates(backupFilename string) error {
	fmt.Println("Importing backup file")
	db := model.Database{
		SkipPlaylists: true,
	}
	err := db.ImportJWLBackup(backupFilename)
	if err != nil {
		return fmt.Errorf("failed to import backup: %w", err)
	}

	fmt.Println("🔍 Scanning for duplicate notes...")
	duplicateGroups := merger.DetectDuplicateNotes(db.Note)

	if len(duplicateGroups) == 0 {
		fmt.Println("✅ No duplicate notes found!")
		return nil
	}

	fmt.Printf("\n⚠️  Found %d group(s) of duplicate notes:\n\n", len(duplicateGroups))

	for i, group := range duplicateGroups {
		fmt.Printf("Duplicate Group #%d (%d notes):\n", i+1, len(group.Notes))
		fmt.Println(string([]rune{'\u2500'}[:1]) + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		t := table.NewWriter()
		t.SetStyle(table.StyleRounded)
		t.Style().Options = table.Options{
			DrawBorder:      true,
			SeparateColumns: true,
			SeparateHeader:  true,
			SeparateRows:    false,
		}
		t.SetOutputMirror(os.Stdout)

		t.AppendHeader(table.Row{"Note ID", "GUID", "Title", "Content Preview", "Created", "Last Modified"})

		for _, note := range group.Notes {
			title := ""
			if note.Title.Valid {
				title = note.Title.String
			}

			content := ""
			if note.Content.Valid {
				content = note.Content.String
				// Limit content preview to 50 characters
				if len(content) > 50 {
					content = content[:50] + "..."
				}
			}

			t.AppendRow(table.Row{
				note.NoteID,
				note.GUID,
				title,
				content,
				note.Created,
				note.LastModified,
			})
		}

		t.Render()
		fmt.Println()
	}

	fmt.Printf("Total duplicate notes found: %d\n", countTotalDuplicates(duplicateGroups))
	return nil
}

func countTotalDuplicates(groups []merger.DuplicateGroup) int {
	total := 0
	for _, group := range groups {
		total += len(group.Notes)
	}
	return total
}

func init() {
	rootCmd.AddCommand(detectDuplicatesCmd)
}
