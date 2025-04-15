package cmd

import (
	"errors"
	"fmt"
	"github.com/cli/go-gh"
	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

type Flag struct {
	Short       string
	Long        string
	Description string
	Required    bool
	Type        string // "bool", "string", "stringSlice"
	Default     string
}

var (
	BranchFlag = Flag{
		Short: "B",
		Long:  "branch",
		Description: "The name of the target branch of the commit. When used in conjunction with the --use-pr " +
			"flag, the branch is the base ref of the PR created. Otherwise, the commit is pushed directly to the " +
			"branch.",
		Required: true,
		Type:     "string",
	}

	MessageFlag = Flag{
		Short: "m",
		Long:  "message",
		Description: "The message connected to the commit. When used in conjunction with --use-pr, the " +
			"commit message is used as the PR title and PR description, unless overridden.",
		Required: true,
		Type:     "string",
	}

	UsePrFlag = Flag{
		Short: "P",
		Long:  "use-pr",
		Description: "Create a PR rather than committing directly to the branch. Unless a head-ref is " +
			"specified manually, the head ref name is generated in the format of <BASE REF>-<RANDOM SALT>.",
		Required: false,
		Type:     "bool",
	}

	HeadRefFlag = Flag{Short: "H", Long: "head-ref", Description: "The name of the branch created with the base ref being `--branch`. Only relevant if used in conjunction with the --use-pr flag.", Type: "string"}
	PrTitleFlag = Flag{Short: "T", Long: "title", Description: "The title of the PR created. Only relevant if used in conjunction with the --use-pr flag. If not specified, the PR title will be the commit message.", Type: "string"}
	PrDescFlag  = Flag{Short: "D", Long: "pr-description", Description: "The description of the PR created. Only relevant if used in conjunction with the --use-pr flag. If not specified, the PR title will be the commit message.", Type: "string"}
	PrLabelFlag = Flag{Short: "l", Long: "label", Description: "A list of labels to add to the PR created. Only relevant if used in conjunction with the --use-pr flag. Labels can be added recursively -- i.e. -l feature -l blocked.", Type: "stringSlice"}
	SyncLocal   = Flag{Short: "s", Long: "sync-local", Description: "Sync the local branch with the remote branch. Only relevant if the target branch is the same as the local branch. Incompatible with --use-pr flag.", Type: "bool", Default: "false"}
	AllFlag     = Flag{Short: "A", Long: "all", Description: "Commit all tracked files that have changed. Only relevant if the target branch is the same as the local branch.", Type: "bool", Default: "false"}
	Untracked   = Flag{Short: "U", Long: "untracked", Description: "Include untracked files in the commit. Only relevant if used in conjunction with the --all flag.", Type: "bool", Default: "false"}
	DryRun      = Flag{Short: "d", Long: "dry-run", Description: "Show which files would be committed.", Type: "bool", Default: "false"}
)

var allFlags = []Flag{
	BranchFlag,
	MessageFlag,
	UsePrFlag,
	HeadRefFlag,
	PrTitleFlag,
	PrDescFlag,
	PrLabelFlag,
	SyncLocal,
	AllFlag,
	Untracked,
	DryRun,
}

type PrSettings struct {
	BaseRef     string
	HeadRef     string
	Title       string
	Description string
	Labels      []string
}

type CommitSettings struct {
	CommitMessage  string
	CommitToBranch string
}

type RepoSettings struct {
	DefaultBranch    string
	DefaultBranchSha string
}

type RunSettings struct {
	PrSettings     *PrSettings
	CommitSettings *CommitSettings
	RepoSettings   *RepoSettings
	FileSelection  []string
	SyncLocal      bool
	DryRun         bool
}

func GetFileSelection(args []string, commitAll bool, commitUntracked bool) ([]string, error) {
	if (commitAll || commitUntracked) && len(args) > 0 {
		return nil, errors.New("`all` and `untracked` cannot be used with explicit file selection")
	}

	if len(args) == 0 {
		log.Println("[DEBUG] No explicit file selection.")
		if !commitAll {
			return nil, errors.New("no files were selected for commit")
		}
	}

	stagedFiles, err := ListStagedFiles()
	if err == nil && len(stagedFiles) > 0 {
		log.Printf("[WARN] %d file(s) are already staged for commit\n", len(stagedFiles))
	}

	untrackedFiles, err := ListUntrackedFiles()
	if err != nil {
		return nil, err
	}

	var filesToAdd []string
	if commitAll {
		filesToAdd, err = ListAllFilesByPattern("*")
		if err != nil {
			return nil, err
		}
		if !commitUntracked {
			untrackedSet := make(map[string]struct{}, len(untrackedFiles))
			for _, f := range untrackedFiles {
				untrackedSet[f] = struct{}{}
			}
			var filtered []string
			for _, f := range filesToAdd {
				if _, isUntracked := untrackedSet[f]; !isUntracked {
					filtered = append(filtered, f)
				}
			}
			filesToAdd = filtered
		}
	} else {
		filesToAdd, err = ListAllFilesByPattern(args...)
		if err != nil {
			return nil, err
		}
	}
	return append(filesToAdd, stagedFiles...), nil
}

func ValidateAndConfigureRun(args []string, cmd *cobra.Command, rs *RepoSettings) (*RunSettings, error) {
	fileSelection, err := GetFileSelection(
		args,
		func() bool { b, _ := cmd.Flags().GetBool(AllFlag.Long); return b }(),
		func() bool { b, _ := cmd.Flags().GetBool(Untracked.Long); return b }(),
	)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Selected %d file(s) for commit\n", len(fileSelection))

	var prSettings *PrSettings
	var commitSettings *CommitSettings
	dryRun, _ := cmd.Flags().GetBool(DryRun.Long)
	usePr, _ := cmd.Flags().GetBool(UsePrFlag.Long)
	branch, _ := cmd.Flags().GetString(BranchFlag.Long)
	commitMessage, _ := cmd.Flags().GetString(MessageFlag.Long)
	syncLocal, _ := cmd.Flags().GetBool(SyncLocal.Long)

	if usePr {
		headRef, _ := cmd.Flags().GetString(HeadRefFlag.Long)
		if headRef == "" {
			uuidValue, _ := uuid.NewV7()
			headRef = fmt.Sprintf("%s-%s", branch, uuidValue)
		}

		labels, _ := cmd.Flags().GetStringSlice(PrLabelFlag.Long)
		title, _ := cmd.Flags().GetString(PrTitleFlag.Long)
		description, _ := cmd.Flags().GetString(PrDescFlag.Long)

		if title == "" {
			title = commitMessage
		}

		if description == "" {
			title = description
		}

		prSettings = &PrSettings{
			BaseRef:     branch,
			HeadRef:     headRef,
			Labels:      labels,
			Description: description,
			Title:       title,
		}

		commitSettings = &CommitSettings{
			CommitMessage:  commitMessage,
			CommitToBranch: headRef,
		}

	} else {
		prSettings = nil
		commitSettings = &CommitSettings{
			CommitMessage:  commitMessage,
			CommitToBranch: branch,
		}
	}

	runSettings := &RunSettings{
		PrSettings:     prSettings,
		CommitSettings: commitSettings,
		FileSelection:  fileSelection,
		SyncLocal:      syncLocal,
		DryRun:         dryRun,
		RepoSettings:   rs,
	}

	fmt.Println(fileSelection)
	return runSettings, nil
}

func generateHelpText() string {
	builder := &strings.Builder{}
	builder.WriteString(`gh-commit: Commit files using the GitHub API.

Commits made via the API will be recognized as signed if used in a GitHub
Actions runner. Commits made with a Personal Access Token (PAT) will also
appear as signed.

Synopsis:
  gh commit [files] -B <branch> -m <message> [flags]

Flags:
`)

	table := tablewriter.NewWriter(builder)
	table.SetAutoWrapText(true)
	table.SetColWidth(80) // widen description column
	table.SetBorder(false)
	table.SetColumnSeparator(" ")
	table.SetCenterSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("  ") // pad columns with 2 spaces
	table.SetNoWhiteSpace(true)

	for _, f := range allFlags {
		table.Append([]string{
			"-" + f.Short + ",", "--" + f.Long, f.Description,
		})
	}
	table.Append([]string{"-V,", "--version", "Print current version"})
	table.Append([]string{"-h,", "--help", "Show this help message"})

	table.Render()
	return builder.String()
}

var rootCmd = &cobra.Command{
	Use:   "gh-commit",
	Short: "gh-commit: commit files easily to git using the Github API",
	Long:  "gh-commit: a CLI tool for committing changes via the Github API, especially useful for working in ephemeral environments.",
	RunE: func(cmd *cobra.Command, args []string) error {
		version, _ := cmd.Flags().GetBool("version")
		if version {
			fmt.Println("gh-commit", VERSION)
		}

		path, err := ValidateLocalGit()
		if err != nil {
			return err
		} else {
			RootPath = path
		}

		repoSettings, err := ValidateGitRemote()
		if err != nil {
			return err
		}

		_, err = gh.CurrentRepository()
		if err != nil {
			return err
		}

		settings, _ := ValidateAndConfigureRun(args, cmd, repoSettings)
		if settings.DryRun {
			settings.ExecuteDryRun()
		} else {
			err := settings.Commit()
			return err
		}

		return nil
	},
}
