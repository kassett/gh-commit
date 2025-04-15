package cmd

type RepoDescriptionResponse struct {
	DefaultBranch string `json:"default_branch"`
}

type BranchDescriptionResponse struct {
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}
