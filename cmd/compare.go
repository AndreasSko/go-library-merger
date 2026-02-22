package cmd

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/AndreasSko/go-library-merger/model"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var compareCmd = &cobra.Command{
	Use:     "compare <left-backup> <right-backup>",
	Short:   "Compare two JW Library backup files to see if they are equal",
	Example: `library-merger compare left.jwlibrary right.jwlibrary`,
	Run: func(cmd *cobra.Command, args []string) {
		leftFilename := args[0]
		rightFilename := args[1]
		compare(leftFilename, rightFilename, terminal.Stdio{In: os.Stdin, Out: os.Stdout, Err: os.Stderr})
	},
	Args: cobra.ExactArgs(2),
}

func compare(leftFilename string, rightFilename string, stdio terminal.Stdio) {
	fmt.Fprintln(stdio.Out, "Importing left backup")
	left := &model.Database{}
	err := left.ImportJWLBackup(leftFilename)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(stdio.Out, "Importing right backup")
	right := &model.Database{}
	err = right.ImportJWLBackup(rightFilename)
	if err != nil {
		log.Fatal(err)
	}

	equal := left.Equals(right)
	if equal {
		fmt.Fprintln(stdio.Out, "✅ Backups are equal")
	} else {
		fmt.Fprintln(stdio.Out, "❌ Backups are NOT equal")
	}
}

func init() {
	rootCmd.AddCommand(compareCmd)
}
