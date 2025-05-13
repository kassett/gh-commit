#!/bin/bash

set -e

current_branch=$(git rev-parse --abbrev-ref HEAD)

# --------------------------
# **TEST BLOCK 1: Simple commit and branch delete**
# Purpose: Ensure basic `gh commit` works with/without staging flags
# --------------------------
echo "[TEST 1] Starting basic branch commit/delete tests"
for i in {1..2}; do
  echo "[TEST 1.$i] Creating temp files"
  mkdir tmp > /dev/null 2>&1
  echo "RANDOM1" > tmp/random1.txt
  echo "RANDOM2" > tmp/random2.txt

  branch="random-nightly-branch-$(date +%F)-$((RANDOM % 100 + 1))"
  echo "[TEST 1.$i] Using branch $branch"

  if [[ $i -eq 1 ]]; then
    echo "[TEST 1.$i] Committing explicitly specified files"
    gh commit tmp/random1.txt tmp/random2.txt -m "Randomly commit for nightly test." -B $branch
  else
    echo "[TEST 1.$i] Committing all staged + untracked files (-U -A)"
    gh commit -U -A -m "Randomly commit for nightly test." -B $branch
  fi

  rm -rf tmp
  git fetch
  git checkout $branch
  echo "[TEST 1.$i] Viewing committed files:"
  cat tmp/random1.txt || true
  cat tmp/random2.txt || true

  rm -rf tmp
  git switch -
  sleep 1

  echo "[TEST 1.$i] Deleting branch $branch"
  repo_info=$(gh repo view --json owner,name -q '.owner.login + "/" + .name')
  gh api -X DELETE repos/$repo_info/git/refs/heads/$branch
  git checkout $current_branch
done

# --------------------------
# **TEST BLOCK 2: Repeat of Test 1**
# Purpose: Verify stability across repeated runs
# --------------------------
echo "[TEST 2] Repeating commit/delete sequence"
for i in {1..2}; do
  echo "[TEST 2.$i] Creating temp files"
  mkdir tmp > /dev/null 2>&1
  echo "RANDOM1" > tmp/random1.txt
  echo "RANDOM2" > tmp/random2.txt

  branch="random-nightly-branch-$(date +%F)-$((RANDOM % 100 + 1))"
  echo "[TEST 2.$i] Using branch $branch"

  if [[ $i -eq 1 ]]; then
    echo "[TEST 2.$i] Committing with file paths"
    gh commit tmp/random1.txt tmp/random2.txt -m "Randomly commit for nightly test." -B $branch
  else
    echo "[TEST 2.$i] Committing with -U -A"
    gh commit -U -A -m "Randomly commit for nightly test." -B $branch
  fi

  rm -rf tmp
  git fetch
  git checkout $branch
  echo "[TEST 2.$i] Viewing committed files:"
  cat tmp/random1.txt || true
  cat tmp/random2.txt || true

  rm -rf tmp
  git switch -
  sleep 1

  echo "[TEST 2.$i] Deleting branch $branch"
  repo_info=$(gh repo view --json owner,name -q '.owner.login + "/" + .name')
  gh api -X DELETE repos/$repo_info/git/refs/heads/$branch
  git checkout $current_branch
done

# --------------------------
# **TEST BLOCK 3: PR creation and cleanup**
# Purpose: Validate PR creation using -P, with custom base and head
# --------------------------
echo "[TEST 3] Starting PR creation/cleanup tests"
for i in {1..2}; do
  echo "[TEST 3.$i] Creating temp files"
  mkdir tmp > /dev/null 2>&1
  echo "RANDOM1" > tmp/random1.txt
  echo "RANDOM2" > tmp/random2.txt

  branch="random-nightly-branch-$(date +%F)-$((RANDOM % 100 + 1))"
  headRef="random-nightly-branch-$(date +%F)-$((RANDOM % 1000 + 1))"
  echo "[TEST 3.$i] Base: $branch, Head: $headRef"

  if [[ $i -eq 1 ]]; then
    echo "[TEST 3.$i] Creating PR using file paths"
    gh commit tmp/random1.txt tmp/random2.txt -P -m "Randomly commit for nightly test." -B $branch -H $headRef -T "Title of PR"
  else
    echo "[TEST 3.$i] Creating PR with -U -A"
    gh commit -U -A -P -m "Randomly commit for nightly test." -B $branch -H $headRef -T "Title of PR"
  fi

  git fetch
  git checkout $branch
  echo "[TEST 3.$i] Viewing committed files:"
  cat tmp/random1.txt || true
  cat tmp/random2.txt || true

  rm -rf tmp
  git switch -
  sleep 1

  echo "[TEST 3.$i] Closing PR and cleaning up branches"
  pr_number=$(gh pr list --head "$headRef" --json number -q '.[0].number')
  echo "[TEST 3.$i] PR #$pr_number"

  repo_info=$(gh repo view --json owner,name -q '.owner.login + "/" + .name')
  gh pr close $pr_number
  gh api -X DELETE repos/$repo_info/git/refs/heads/$branch
  gh api -X DELETE repos/$repo_info/git/refs/heads/$headRef
  git checkout $current_branch
done

# --------------------------
# **FINAL TEST: PR with empty commit**
# Purpose: Ensure `gh commit` can push empty changes via PR
# --------------------------
echo "[TEST 4] Creating empty PR commit"
gh commit -P -B random-branch-name -m "Random empty commit" -A -U
