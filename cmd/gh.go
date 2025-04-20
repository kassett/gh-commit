package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/fatih/color"
	"net/http"
	"os"
	"strconv"
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

func EnsureBranchesExist(targetBranch, intermediateBranch string, repoSettings *RepoSettings) (string, error) {
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
					return "", errors.New(fmt.Sprintf("you are not authorized to create the ref %s", targetBranch))
				} else {
					return "", errors.New(fmt.Sprintf("error creating branch: %s", err))
				}
			}
		} else {
			return "", errors.New(fmt.Sprint("Error getting branch description: ", err))
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
				return "", errors.New(fmt.Sprintf("you are not authorized to create the ref %s", intermediateBranch))
			} else {
				return "", errors.New(fmt.Sprintf("error creating branch: %s", err))
			}
		}
	}

	return headShaForIntermediateBranch, nil
}

func GetTreeTip(commitSha string) (string, error) {
	// We already validated that we have a commit, so there should be no errors here
	var res ShaResponse
	err := client.Get(fmt.Sprintf("repos/%s/%s/git/trees/%s", repo.Owner(), repo.Name(), commitSha), &res)
	// I don't know what a relevant error here would be
	if err != nil {
		return "", errors.New(fmt.Sprint("error getting tree description: ", err))
	}
	return res.Sha, nil
}

// CreateBlobs creates the leaves of the trees that commits reference.
func CreateBlobs(files []string) ([]BlobInfo, error) {
	blobs := make([]BlobInfo, 0)
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			// Nil is for when we delete a file
			blobs = append(blobs, BlobInfo{
				Path: file,
				Mode: "100644",
				Type: "blob",
				Sha:  nil,
			})

		} else {

			blobSha, err := CreateBlob(file)
			if err != nil {
				return nil, err
			}

			blobs = append(blobs, BlobInfo{
				Path: file,
				Mode: "100644",
				Type: "blob",
				Sha:  &blobSha,
			})
		}
	}
	return blobs, nil
}

func CreateBlob(file string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", errors.New(fmt.Sprint("Error reading file: ", err))
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	// In the first GH action version of this, we would get errors because the encoding
	// would be too large for Bash to handle, so we would use the --input argument
	// to pass a file name. We do not need to do this here.
	var blobResponse ShaResponse
	err = client.Post(
		fmt.Sprintf("repos/%s/%s/git/blobs", repo.Owner(), repo.Name()),
		bytes.NewBuffer([]byte(fmt.Sprintf(
			`{"content": "%s", "encoding": "base64"}`, encoded,
		))),
		&blobResponse,
	)

	if err != nil {
		return "", errors.New(fmt.Sprint("Error creating blob: ", err))
	}

	return blobResponse.Sha, nil
}

func CreateTree(baseTree string, blobs []BlobInfo) (string, error) {
	tree := map[string]interface{}{
		"base_tree": baseTree,
		"tree":      blobs,
	}

	marshalled, _ := json.Marshal(tree)

	var treeResponse ShaResponse
	err := client.Post(
		fmt.Sprintf("repos/%s/%s/git/trees", repo.Owner(), repo.Name()),
		bytes.NewBuffer(marshalled),
		&treeResponse)
	if err != nil {
		return "", errors.New(fmt.Sprint("error creating tree: ", err))
	}

	return treeResponse.Sha, nil
}

func CreateCommitFromTree(latestCommit, treeSha, commitMessage string) (string, error) {
	body := map[string]interface{}{
		"message": commitMessage,
		"tree":    treeSha,
		"parents": []string{latestCommit},
	}
	marshalled, _ := json.Marshal(body)
	var newCommitResponse ShaResponse
	err := client.Post(
		fmt.Sprintf("repos/%s/%s/git/commits", repo.Owner(), repo.Name()),
		bytes.NewBuffer(marshalled),
		&newCommitResponse,
	)

	if err != nil {
		return "", errors.New(fmt.Sprint("error creating commit: ", err))
	}

	return newCommitResponse.Sha, nil
}

func AssociateCommitWithBranch(branch string, commitSha string) error {
	body := map[string]interface{}{
		"sha": commitSha,
	}
	marshalled, _ := json.Marshal(body)
	err := client.Post(fmt.Sprintf("repos/%s/%s/git/refs/heads/%s", repo.Owner(), repo.Name(), branch), bytes.NewBuffer(marshalled), nil)
	if err != nil {
		if httpErr, ok := err.(api.HTTPError); ok && httpErr.StatusCode == http.StatusForbidden {
			return errors.New(fmt.Sprintf("you are not authorized to make commits on this branch %s", branch))
		}
	}

	if isGitHubAction() {
		_ = exportGitHubOutput("sha", commitSha)
	}

	return err
}

func ValidateAllLabels(labels []string) error {
	for _, label := range labels {
		err := client.Get(
			fmt.Sprintf("repos/%s/%s/labels/%s", repo.Owner(), repo.Name(), label),
			nil)
		if err != nil {
			if err, ok := err.(api.HTTPError); ok && err.StatusCode == http.StatusNotFound {
				return errors.New(fmt.Sprintf("Label %s not found. Create the label first", label))
			}
		}
	}
	return nil
}

func CreatePullRequest(baseRef, headRef, title, description string, labels []string) error {
	var prResponse PrResponse
	body := PrRequest{
		Title: title,
		Body:  description,
		Head:  headRef,
		Base:  baseRef,
	}
	marshalled, _ := json.Marshal(body)
	err := client.Post(
		fmt.Sprintf("repos/%s/%s/pulls", repo.Owner(), repo.Name()),
		bytes.NewBuffer(marshalled),
		&prResponse)
	if err != nil {
		return errors.New(fmt.Sprint("error creating pull request: ", err))
	}

	if len(labels) > 0 {
		labelRequest := LabelRequest{
			Labels: labels,
		}
		marshalled, err = json.Marshal(labelRequest)
		err = client.Put(
			fmt.Sprintf("repos/%s/%s/issues/%d/labels", repo.Owner(), repo.Name(), prResponse.Number),
			bytes.NewBuffer(marshalled),
			nil,
		)
		if err != nil {
			return errors.New(fmt.Sprint("error adding labels to pull request: ", err))
		}
	}

	link := color.New(color.FgBlue, color.Bold).Sprintf("🔗 Pull Request URL: %s", prResponse.Url)
	fmt.Println(link)

	if isGitHubAction() {
		_ = exportGitHubOutput("pr-number", strconv.Itoa(prResponse.Number))
		_ = exportGitHubOutput("branch", headRef)
	}

	return nil
}
