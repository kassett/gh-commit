# gh-commit

> **A fast, GitHub-native CLI tool to commit files using the GitHub API.**

---

## ✨ Overview

`gh-commit` is a GitHub CLI extension that allows you to commit changes directly to your repositories using the GitHub API. It is ideal for use in ephemeral environments, CI workflows, or anywhere you want fast, authenticated commits without a local Git identity.

Supports both direct commits and pull request-based workflows.

---

## 🚀 Features

- 📂 Commit selected files or all changes
- 📎 Create commits using the GitHub API (signed in CI)
- 🌟 Automatically create branches and PRs
- 🔍 Smart file detection (staged, tracked, untracked)
- ✨ Fully styled CLI output with colorized logging

---

## 💡 Usage

```bash
gh commit [files...] -B <branch> -m <message> [flags]
```

### Example

Commit all changes:
```bash
gh commit -B main -A -m "fix: update configs"
```

Create a PR from a new branch:
```bash
gh commit -B main -A -P -T "Update Configs" -D "This PR updates the configs." -l feature -l ci
```

Dry run (shows what would be committed):
```bash
gh commit -B main -A -d
```

---

## 🔧 Flags

| Short | Long           | Type         | Description                                                                 |
|-------|----------------|--------------|-----------------------------------------------------------------------------|
| -B    | --branch       | `string`     | Target branch (base for PRs or direct commit target) **(required)**        |
| -m    | --message      | `string`     | Commit message (and PR title if applicable) **(required)**                |
| -P    | --use-pr       | `bool`       | Create a pull request instead of committing directly                       |
| -H    | --head-ref     | `string`     | PR head branch name (auto-generated if omitted)                            |
| -T    | --title        | `string`     | Pull request title (defaults to commit message)                            |
| -D    | --pr-description | `string`   | Pull request body (defaults to commit message)                             |
| -l    | --label        | `stringSlice`| Add one or more labels to the pull request                                 |
| -A    | --all          | `bool`       | Include all tracked files with changes                                     |
| -U    | --untracked    | `bool`       | Include untracked files (requires `--all`)                                 |
| -d    | --dry-run      | `bool`       | Show which files would be committed, without committing                    |
| -b    | --base-commit  | `string`     | Specify a specific commit SHA to use as the base for the new commit       |
| -f    | --allow-fast-forward | `bool` | Allow fast-forward operations when using --base-commit                    |
| -V    | --version      | `bool`       | Show version                                                               |
| -h    | --help         | `bool`       | Show help text                                                             |

---

## 📆 GitHub Actions

```yaml
- name: Commit and Push Changes
  run: |
    gh commit -B main -A -m "ci: auto-commit"
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

Signed commits are automatically created when using GitHub Actions.

---

## 🔄 Base Commit and Fast-Forward

The `--base-commit` flag allows you to specify a specific commit SHA to use as the base for your new commit. This is particularly useful for:

- **Avoiding duplicate commits** after rebases
- **Simulating force pushes** without actually force pushing
- **Maintaining clean commit history** in PRs

### Usage Examples

**Basic base commit usage:**
```bash
gh commit -B main -m "fix: update configs" -b a1b2c3d4e5f6789012345678901234567890abcd -f file.txt
```

**With fast-forward (simulates force push):**
```bash
gh commit -B main -m "fix: update configs" -b a1b2c3d4e5f6789012345678901234567890abcd --allow-fast-forward -f file.txt
```

### How It Works

- **`--base-commit` only**: Creates a new commit on top of the specified commit, but doesn't modify the branch pointer
- **`--base-commit` + `--allow-fast-forward`**: Fast-forwards the branch to the specified commit, then creates the new commit on top

This effectively simulates a force push by allowing you to "rewrite history" from a specific point, preventing the "Update is not a fast forward" error.

---

## 📂 Project Structure

- `cmd/`
    - CLI definitions and flag parsing
    - Commit/PR creation logic

---

## 🚨 Errors & Validation

- Ensures repo has a remote and is a Git repo
- Validates presence of commit message and branch
- Prevents mixed usage of `--all`, `--untracked`, and file args
- PRs auto-create branches if not found
- Label validation before PR creation

---

## 📄 License

MIT © github.com/kassett

---

## 🌐 Links

- GitHub: [kassett/gh-commit](https://github.com/kassett/gh-commit)
- GH CLI: [cli.github.com](https://cli.github.com/)

