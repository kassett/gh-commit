package cmd

import (
	"fmt"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

// VERSION number: changed in CI
const VERSION = "v0.0.13"

var rootPath string
var repo repository.Repository
var client api.RESTClient

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

var executor CommandExecutor = &DefaultCommandExecutor{}

func init() {
	for _, flag := range allFlags {
		switch flag.Type {
		case "bool":
			rootCmd.Flags().BoolP(flag.Long, flag.Short, flag.Default == "true", flag.Description)
		case "string":
			rootCmd.Flags().StringP(flag.Long, flag.Short, flag.Default, flag.Description)
		case "stringSlice":
			rootCmd.Flags().StringSliceP(flag.Long, flag.Short, []string{}, flag.Description)
		}
		if flag.Required {
			_ = rootCmd.MarkFlagRequired(flag.Long)
		}
	}

	rootCmd.Flags().BoolP("version", "V", false, "Print current version")
	rootCmd.SetHelpTemplate(generateHelpText())
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
	var err error
	var commitSha string

	// Create branches so we don't have to worry about those errors later
	if rn.PrSettings != nil {
		commitSha, err = EnsureBranchesExist(rn.PrSettings.BaseRef, rn.PrSettings.HeadRef, rn.RepoSettings)
	} else {
		commitSha, err = EnsureBranchesExist(rn.CommitSettings.CommitToBranch, "", rn.RepoSettings)
	}
	if err != nil {
		return err
	}

	// Commits reference trees. Trees have their own hashes. Get the hash
	// of the tip of the tree that we are pushing to
	currentTreeSha, err := GetTreeTip(commitSha)
	if err != nil {
		return err
	}

	blobs, err := CreateBlobs(rn.FileSelection)
	if err != nil {
		return err
	}

	newTreeSha, err := CreateTree(currentTreeSha, blobs)
	newCommit, err := CreateCommitFromTree(commitSha, newTreeSha, rn.CommitSettings.CommitMessage)

	if err != nil {
		return err
	}

	err = AssociateCommitWithBranch(rn.CommitSettings.CommitToBranch, newCommit)
	if err != nil {
		return err
	}

	if rn.PrSettings != nil {
		err = CreatePullRequest(
			rn.PrSettings.BaseRef,
			rn.PrSettings.HeadRef,
			rn.PrSettings.Title,
			rn.PrSettings.Description,
			rn.PrSettings.Labels)
		if err != nil {
			return err
		}
	}

	return nil
}
