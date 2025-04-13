package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func ListUntrackedFiles() ([]string, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = RootPath
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}
	return files, nil
}

func ListStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--cached")
	cmd.Dir = RootPath
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	files := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}
	return files, nil
}

func ListAllFilesByPattern(patterns ...string) ([]string, error) {
	args := append([]string{"add", "--dry-run", "--verbose"}, patterns...)
	cmd := exec.Command("git", args...)
	cmd.Dir = RootPath
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, errors.New("the pattern(s) did not match any files")
	}

	var files []string
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		// output format: "add 'filename'"
		parts := strings.SplitN(line, "'", 2)
		if len(parts) == 2 {
			files = append(files, strings.TrimSuffix(parts[1], "'"))
		}
	}

	return files, nil
}

func ValidateGitRepo() (string, error) {
	// Ensure we are inside a Git repo
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("not a git repository")
	}

	// Get the repo root path
	out := &bytes.Buffer{}
	cmd = exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get git root: %w", err)
	}

	repoRoot := strings.TrimSpace(out.String())

	// Ensure at least one remote exists
	out.Reset()
	cmd = exec.Command("git", "remote")
	cmd.Dir = RootPath
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to check remotes: %w", err)
	}
	if len(strings.TrimSpace(out.String())) == 0 {
		return "", fmt.Errorf("git repository has no remotes configured")
	}

	return repoRoot, nil
}
