package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"

	gh "github.com/cli/go-gh"
)

// VERSION number: changed in CI
const VERSION = "0.0.3"

var RootPath string

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func (rn *RunSettings) ExecuteDryRun() {
	fmt.Print("The following files would be committed:\n\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for i, item := range rn.FileSelection {
		_, _ = fmt.Fprintf(w, "%d.\t%s\n", i+1, item)
	}
	_ = w.Flush()
}

func (rn *RunSettings) Commit() error {
	client, err := gh.RestClient()
	if err != nil {
		return err
	}

	return nil

}
