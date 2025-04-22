package cmd

import (
	"fmt"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

// VERSION number: changed in CI
const VERSION = "v0.2.1"

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
	}

	rootCmd.Flags().BoolP("version", "V", false, "Print current version")
	rootCmd.SetHelpTemplate(generateHelpText())
}

func isGitHubAction() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

func exportGitHubOutput(key, value string) error {
	outputPath := os.Getenv("GITHUB_OUTPUT")
	if outputPath == "" {
		return fmt.Errorf("GITHUB_OUTPUT not set")
	}

	f, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open GITHUB_OUTPUT file: %w", err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = fmt.Fprintf(f, "%s=%s\n", key, value)
	return err
}

func (rn *RunSettings) ExecuteDryRun() {
	heading := color.New(color.FgCyan, color.Bold).SprintFunc()
	fmt.Printf("%s\n\n", heading("The following files would be committed:"))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for i, item := range rn.FileSelection {
		index := color.New(color.FgYellow).Sprintf("%d.", i+1)
		filename := color.New(color.FgWhite).Sprint(item)
		_, _ = fmt.Fprintf(w, "%s\t%s\n", index, filename)
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
