package cmd

type RepoDescriptionResponse struct {
	DefaultBranch string `json:"default_branch"`
}

type BranchDescriptionResponse struct {
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

type ShaResponse struct {
	Sha string `json:"sha"`
}

type BlobInfo struct {
	Path string  `json:"path"`
	Mode string  `json:"mode"`
	Type string  `json:"type"`
	Sha  *string `json:"sha"`
}

type PrResponse struct {
	Url    string `json:"url"`
	Number int    `json:"number"`
}

type PrRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

type LabelResponse struct {
	Name string `json:"name"`
}

type LabelRequest struct {
	Labels []string `json:"labels"`
}
