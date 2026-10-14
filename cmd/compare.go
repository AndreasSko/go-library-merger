package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/AndreasSko/go-library-merger/model"
	"github.com/spf13/cobra"
)

var compareCmd = &cobra.Command{
	Use:     "compare <left-backup> <right-backup>",
	Short:   "Compare two JW Library backup files to see if they are equal",
	Example: `library-merger compare left.jwlibrary right.jwlibrary`,
	RunE: func(cmd *cobra.Command, args []string) error {
		leftFilename := args[0]
		rightFilename := args[1]
		return compare(leftFilename, rightFilename, terminal.Stdio{In: os.Stdin, Out: os.Stdout, Err: os.Stderr})
	},
	Args: cobra.ExactArgs(2),
}

func compare(leftFilename string, rightFilename string, stdio terminal.Stdio) (resultErr error) {
	fmt.Fprintln(stdio.Out, "Importing left backup")
	left := &model.Database{}
	err := left.ImportJWLBackup(leftFilename)
	if err != nil {
		return err
	}

	defer func() { resultErr = errors.Join(resultErr, left.Close()) }()

	fmt.Fprintln(stdio.Out, "Importing right backup")
	right := &model.Database{}
	err = right.ImportJWLBackup(rightFilename)
	if err != nil {
		return err
	}

	defer func() { resultErr = errors.Join(resultErr, right.Close()) }()

	equal := left.Equals(right)
	if equal {
		fmt.Fprintln(stdio.Out, "✅ Backups are equal")
	} else {
		fmt.Fprintln(stdio.Out, "❌ Backups are NOT equal")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(compareCmd)
}
