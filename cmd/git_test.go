package cmd

import (
	"errors"
	"strings"
	"testing"
)

type CommandOutput struct {
	Output []byte
	Err    error
}

// MockCommandExecutor is a mock implementation of CommandExecutor for testing.
type MockCommandExecutor struct {
	Simple  *CommandOutput
	Complex map[string]*CommandOutput
}

// RunCommand simulates command execution for testing.
func (m *MockCommandExecutor) RunCommand(name string, arg ...string) ([]byte, error) {
	if m.Simple != nil {
		return m.Simple.Output, m.Simple.Err
	}
	currentCommand := []string{name}
	for _, arg := range arg {
		currentCommand = append(currentCommand, arg)
	}
	key := strings.Join(currentCommand, " ")
	if element, ok := m.Complex[key]; ok {
		return element.Output, element.Err
	}
	return nil, errors.New(name + " not found")

}

func TestListUntrackedFiles(t *testing.T) {
	tests := []struct {
		name          string
		expectedFiles []string
		expectedError error
		complex       map[string]*CommandOutput
	}{
		{
			name:          "No untracked files",
			expectedFiles: []string{},
			expectedError: nil,
			complex: map[string]*CommandOutput{
				"git ls-files --others --exclude-standard": {Output: []byte(""), Err: nil},
			},
		},
		{
			name:          "One untracked file",
			expectedFiles: []string{"file1.txt"},
			expectedError: nil,
			complex: map[string]*CommandOutput{
				"git ls-files --others --exclude-standard": {Output: []byte("file1.txt\n"), Err: nil},
			},
		},
		{
			name:          "Multiple untracked files",
			expectedFiles: []string{"file1.txt", "file2.txt"},
			expectedError: nil,
			complex: map[string]*CommandOutput{
				"git ls-files --others --exclude-standard": {Output: []byte("file1.txt\nfile2.txt\n"), Err: nil},
			},
		},
		{
			name:          "Command error",
			expectedFiles: nil,
			expectedError: errors.New("command error"),
			complex: map[string]*CommandOutput{
				"git ls-files --others --exclude-standard": {Output: nil, Err: errors.New("command error")},
			},
		},
	}

	// Save the original executor
	originalExecutor := executor
	defer func() { executor = originalExecutor }() // Restore original executor after tests

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor = &MockCommandExecutor{Complex: tt.complex} // Use the mock executor

			files, err := ListUntrackedFiles()

			if err != nil && tt.expectedError != nil {
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else if err != nil || tt.expectedError != nil {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}

			if !equal(files, tt.expectedFiles) {
				t.Errorf("expected files %v, got %v", tt.expectedFiles, files)
			}
		})
	}
}

func TestListStagedFiles(t *testing.T) {
	tests := []struct {
		name          string
		expectedFiles []string
		expectedError error
		complex       map[string]*CommandOutput
	}{
		{
			name:          "No staged files",
			expectedFiles: []string{},
			expectedError: nil,
			complex: map[string]*CommandOutput{
				"git diff --name-only --cached": {Output: []byte(""), Err: nil},
			},
		},
		{
			name:          "One staged file",
			expectedFiles: []string{"file1.txt"},
			expectedError: nil,
			complex: map[string]*CommandOutput{
				"git diff --name-only --cached": {Output: []byte("file1.txt\n"), Err: nil},
			},
		},
		{
			name:          "Multiple staged files",
			expectedFiles: []string{"file1.txt", "file2.txt"},
			expectedError: nil,
			complex: map[string]*CommandOutput{
				"git diff --name-only --cached": {Output: []byte("file1.txt\nfile2.txt\n"), Err: nil},
			},
		},
		{
			name:          "Command error",
			expectedFiles: nil,
			expectedError: errors.New("command error"),
			complex: map[string]*CommandOutput{
				"git diff --name-only --cached": {Output: nil, Err: errors.New("command error")},
			},
		},
	}

	// Save the original executor
	originalExecutor := executor
	defer func() { executor = originalExecutor }() // Restore original executor after tests

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor = &MockCommandExecutor{Complex: tt.complex} // Use the mock executor

			files, err := ListStagedFiles()

			if err != nil && tt.expectedError != nil {
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else if err != nil || tt.expectedError != nil {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}

			if !equal(files, tt.expectedFiles) {
				t.Errorf("expected files %v, got %v", tt.expectedFiles, files)
			}
		})
	}
}

func TestListAllFilesByPattern(t *testing.T) {
	tests := []struct {
		name          string
		expectedFiles []string
		expectedError error
		patterns      []string
		complex       map[string]*CommandOutput
	}{
		{
			name:          "No matching files",
			expectedFiles: nil,
			expectedError: errors.New("the pattern(s) did not match any files"),
			patterns:      []string{"*.go"},
			complex: map[string]*CommandOutput{
				"git add --dry-run --verbose *.go": {Output: []byte(""), Err: errors.New("command error")},
			},
		},
		{
			name:          "One matching file",
			expectedFiles: []string{"file1.go"},
			expectedError: nil,
			patterns:      []string{"*.go"},
			complex: map[string]*CommandOutput{
				"git add --dry-run --verbose *.go": {Output: []byte("add 'file1.go'\n"), Err: nil},
			},
		},
		{
			name:          "Multiple matching files",
			expectedFiles: []string{"file1.go", "file2.go"},
			expectedError: nil,
			patterns:      []string{"*.go"},
			complex: map[string]*CommandOutput{
				"git add --dry-run --verbose *.go": {Output: []byte("add 'file1.go'\nadd 'file2.go'\n"), Err: nil},
			},
		},
		{
			name:          "Command error",
			expectedFiles: nil,
			expectedError: errors.New("the pattern(s) did not match any files"),
			patterns:      []string{"*.go"},
			complex: map[string]*CommandOutput{
				"git add --dry-run --verbose *.go": {Output: nil, Err: errors.New("command error")},
			},
		},
	}

	// Save the original executor
	originalExecutor := executor
	defer func() { executor = originalExecutor }() // Restore original executor after tests

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor = &MockCommandExecutor{Complex: tt.complex} // Use the mock executor

			files, err := ListAllFilesByPattern(tt.patterns...)

			if err != nil && tt.expectedError != nil {
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else if err != nil || tt.expectedError != nil {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}

			if !equal(files, tt.expectedFiles) {
				t.Errorf("expected files %v, got %v", tt.expectedFiles, files)
			}
		})
	}
}

func TestValidateLocalGit(t *testing.T) {
	tests := []struct {
		name          string
		expectedRoot  string
		expectedError error
		complex       map[string]*CommandOutput
	}{
		{
			name:          "Not a git repository",
			expectedRoot:  "",
			expectedError: errors.New("not a git repository"),
			complex: map[string]*CommandOutput{
				"git rev-parse --is-inside-work-tree": {Output: nil, Err: errors.New("command error")},
			},
		},
		{
			name:          "Failed to get repo root",
			expectedRoot:  "",
			expectedError: errors.New("failed to get git root: command error"),
			complex: map[string]*CommandOutput{
				"git rev-parse --is-inside-work-tree": {Output: []byte("true"), Err: nil},
				"git rev-parse --show-toplevel":       {Output: nil, Err: errors.New("command error")},
			},
		},
		{
			name:          "No remotes configured",
			expectedRoot:  "",
			expectedError: errors.New("git repository has no remotes configured"),
			complex: map[string]*CommandOutput{
				"git rev-parse --is-inside-work-tree": {Output: []byte("true"), Err: nil},
				"git rev-parse --show-toplevel":       {Output: []byte("/path/to/repo"), Err: nil},
				"git remote":                          {Output: []byte(""), Err: nil},
			},
		},
		{
			name:          "Valid git repository",
			expectedRoot:  "/path/to/repo",
			expectedError: nil,
			complex: map[string]*CommandOutput{
				"git rev-parse --is-inside-work-tree": {Output: []byte("true"), Err: nil},
				"git rev-parse --show-toplevel":       {Output: []byte("/path/to/repo"), Err: nil},
				"git remote":                          {Output: []byte("origin\n"), Err: nil},
			},
		},
	}

	// Save the original executor
	originalExecutor := executor
	defer func() { executor = originalExecutor }() // Restore original executor after tests

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor = &MockCommandExecutor{Complex: tt.complex} // Use the mock executor

			repoRoot, err := ValidateLocalGit()

			if err != nil && tt.expectedError != nil {
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else if err != nil || tt.expectedError != nil {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}

			if repoRoot != tt.expectedRoot {
				t.Errorf("expected repo root %v, got %v", tt.expectedRoot, repoRoot)
			}
		})
	}
}

// Helper function to compare slices
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
