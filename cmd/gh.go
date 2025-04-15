package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"net/http"
)

func ValidateGitRemote() (*RepoSettings, error) {
	restClient, err := gh.RESTClient(nil)
	// Any error here is fatal
	if err != nil {
		return nil, err
	}

	repoObj, err := gh.CurrentRepository()
	if err != nil {
		return nil, err
	}

	// Set Globals
	repo = repoObj
	client = restClient

	owner := repo.Owner()
	name := repo.Name()

	var repoDescriptionResult RepoDescriptionResponse
	url := fmt.Sprintf("repos/%s/%s", owner, name)
	err = client.Get(url, &repoDescriptionResult)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Error getting repo description: ", err))
	}

	url = fmt.Sprintf("repos/%s/%s/branches/%s", owner, name, repoDescriptionResult.DefaultBranch)
	var branchDescriptionResult BranchDescriptionResponse
	err = client.Get(url, &branchDescriptionResult)
	if err != nil {
		return nil, errors.New("using the gh-commit extension requires the repo already having an initial commit")
	}

	repoSettings := &RepoSettings{
		DefaultBranch:    repoDescriptionResult.DefaultBranch,
		DefaultBranchSha: branchDescriptionResult.Commit.SHA,
	}
	// Now we get the HeadSha for the default branch
	return repoSettings, nil
}

func EnsureBranchesExist(targetBranch, intermediateBranch string, repoSettings *RepoSettings) error {
	headShaForIntermediateBranch := repoSettings.DefaultBranchSha

	var targetBranchResponse BranchDescriptionResponse
	err := client.Get(
		fmt.Sprintf("repos/%s/%s/branches/%s", repo.Owner(), repo.Name(), targetBranch),
		&targetBranchResponse)
	if err != nil {
		if httpErr, ok := err.(api.HTTPError); ok && httpErr.StatusCode == http.StatusNotFound {
			// Now we create the branch from the default repository
			err = client.Post(
				fmt.Sprintf("repos/%s/%s/git/refs", repo.Owner(), repo.Name()),
				bytes.NewBuffer([]byte(
					fmt.Sprintf(`{"ref": "refs/heads/%s", "sha": "%s"}`,
						targetBranch, repoSettings.DefaultBranchSha),
				)),
				nil,
			)
			if err != nil {
				if httpErr, ok = err.(api.HTTPError); ok && httpErr.StatusCode == http.StatusUnauthorized || httpErr.StatusCode == http.StatusForbidden {
					return errors.New(fmt.Sprintf("you are not authorized to create the ref %s", targetBranch))
				} else {
					return errors.New(fmt.Sprintf("error creating branch: %s", err))
				}
			}
		} else {
			return errors.New(fmt.Sprint("Error getting branch description: ", err))
		}
	} else {
		headShaForIntermediateBranch = targetBranchResponse.Commit.SHA
	}

	// intermediaryBranch is only for the usePr workflow
	if intermediateBranch != "" {
		// Now we create the branch from the default repository
		err = client.Post(
			fmt.Sprintf("repos/%s/%s/git/refs", repo.Owner(), repo.Name()),
			bytes.NewBuffer([]byte(
				fmt.Sprintf(`{"ref": "refs/heads/%s", "sha": "%s"}`,
					intermediateBranch, headShaForIntermediateBranch),
			)),
			nil,
		)

		if err != nil {
			if httpErr, ok := err.(api.HTTPError); ok && httpErr.StatusCode == http.StatusUnauthorized || httpErr.StatusCode == http.StatusForbidden {
				return errors.New(fmt.Sprintf("you are not authorized to create the ref %s", intermediateBranch))
			} else {
				return errors.New(fmt.Sprintf("error creating branch: %s", err))
			}
		}
	}

	return nil
}
