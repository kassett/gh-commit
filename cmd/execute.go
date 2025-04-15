package cmd

import (
	"errors"
	"fmt"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

// VERSION number: changed in CI
const VERSION = "0.0.3"

var RootPath string
var repo repository.Repository
var client api.RESTClient

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

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

	// Create branches so we don't have to worry about those errors later
	if rn.PrSettings != nil {
		err = EnsureBranchesExist(rn.PrSettings.BaseRef, rn.PrSettings.HeadRef, rn.RepoSettings)
	} else {
		err = EnsureBranchesExist(rn.CommitSettings.CommitToBranch, "", rn.RepoSettings)
	}

	if err != nil {
		return err
	}

	return nil
}

//
//func CreateBranches(
//	client api.RESTClient,
//	repo repository.Repository,
//	defaultRepoBranch string,
//	headRepoSha string,
//	destinationBranch string,
//	intermediaryBranch string) (string, error){
//
//	// First step is to check if the destination branch exists.
//	// If not, we get the headRef of the default branch of the repo
//	owner := repo.Owner()
//	repoName := repo.Name()
//
//	var headRefForIntermediaryBranch string
//
//	// Check if destination branch exists
//	url := fmt.Sprintf("repos/%s/%s/git/refs/heads/%s", owner, repoName, destinationBranch)
//	var apiResults map[string]interface{}
//	err := client.Get(url, &apiResults)
//	if err != nil {
//		if httpErr, ok := err.(api.HTTPError); ok && httpErr.StatusCode == http.StatusNotFound {
//
//		}
//	}
//}

func GetHeadSha(client api.RESTClient, repo repository.Repository, branch string) (string, error) {
	owner := repo.Owner()
	repoName := repo.Name()

	url := fmt.Sprintf("repos/%s/%s/git/refs/heads/%s", owner, repoName, branch)
	var apiResults map[string]interface{}
	err := client.Get(url, &apiResults)
	if err != nil {
		return "", err
	}

	if sha, ok := apiResults["object"].(map[string]interface{})["sha"]; ok {
		return sha.(string), nil
	}

	return "", errors.New("failed to get head SHA")
}
