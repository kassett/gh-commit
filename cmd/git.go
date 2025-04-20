package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// CommandExecutor is an interface for executing commands.
type CommandExecutor interface {
	RunCommand(name string, arg ...string) ([]byte, error)
}

// DefaultCommandExecutor is the default implementation of CommandExecutor.
type DefaultCommandExecutor struct{}

// RunCommand executes a command and returns its output.
func (d *DefaultCommandExecutor) RunCommand(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = rootPath
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func ListUntrackedFiles() ([]string, error) {
	out, err := executor.RunCommand("git", "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}
	return files, nil
}

func ListStagedFiles() ([]string, error) {
	out, err := executor.RunCommand("git", "diff", "--name-only", "--cached")
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}
	return files, nil
}

func ListAllFilesByPattern(patterns ...string) ([]string, error) {
	args := append([]string{"add", "--dry-run", "--verbose"}, patterns...)
	out, err := executor.RunCommand("git", args...)
	if err != nil {
		return nil, errors.New("the pattern(s) did not match any files")
	}

	var files []string
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		// output format: "add 'filename'"
		parts := strings.SplitN(line, "'", 2)
		if len(parts) == 2 {
			files = append(files, strings.TrimSuffix(parts[1], "'"))
		}
	}

	return files, nil
}

func ValidateLocalGit() (string, error) {
	// Ensure we are inside a Git repo
	if _, err := executor.RunCommand("git", "rev-parse", "--is-inside-work-tree"); err != nil {
		return "", fmt.Errorf("not a git repository")
	}

	// Get the repo root path
	out, err := executor.RunCommand("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to get git root: %w", err)
	}

	repoRoot := strings.TrimSpace(string(out))

	// Ensure at least one remote exists
	out, err = executor.RunCommand("git", "remote")
	if err != nil {
		return "", fmt.Errorf("failed to check remotes: %w", err)
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		return "", fmt.Errorf("git repository has no remotes configured")
	}

	return repoRoot, nil
}
