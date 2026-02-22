package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/MakeNowJust/heredoc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var purgeCmd = &cobra.Command{
	Use:   "purge <input-backup> <output-backup>",
	Short: "Remove all entries of specific tables",
	Long: heredoc.Doc(`Remove all entries of the given tables from the input backup and store it as output. 
	This can be useful in case you want to share the backup with someone else, but want to remove some data.
	
	Valid table names are: 
	 * BlockRange
	 * Bookmark
	 * InputField
	 * Location
	 * Note
	 * Tag
	 * TagMap
	 * UserMark`),
	Example: `library-merger purge original.jwlibrary purged.jwlibrary --tables=Note,Tag,TagMap`,
	Run: func(cmd *cobra.Command, args []string) {
		inputFilename := args[0]
		outputFilename := args[1]
		purge(inputFilename, outputFilename, terminal.Stdio{In: os.Stdin, Out: os.Stdout, Err: os.Stderr})
	},
	Args: cobra.ExactArgs(2),
}

var Tables string

func purge(inputFilename string, outputFilename string, stdio terminal.Stdio) {
	fmt.Fprintln(stdio.Out, "Importing backup")
	db := &model.Database{}
	err := db.ImportJWLBackup(inputFilename)
	if err != nil {
		log.Fatal(err)
	}

	Tables = strings.ReplaceAll(Tables, " ", "")
	tableSlice := strings.Split(Tables, ",")

	fmt.Fprintln(stdio.Out, "🔥 Purging the following tables:", strings.TrimSuffix(strings.Join(tableSlice, ", "), ","))
	err = db.PurgeTables(tableSlice)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(stdio.Out, "💾 Storing backup")
	if err = db.ExportJWLBackup(outputFilename); err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(stdio.Out, "🎉 Done")
}

func init() {
	rootCmd.AddCommand(purgeCmd)
	purgeCmd.Flags().StringVar(&Tables, "tables", "", "Comma-separated list of tables that should be purged from the backup")
}
