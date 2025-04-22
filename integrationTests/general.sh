#!/bin/bash

current_branch=$(git rev-parse --abbrev-ref HEAD)

for i in {1..2}; do
  mkdir tmp > /dev/null 2>&1
  echo "RANDOM1" > tmp/random1.txt
  echo "RANDOM2" > tmp/random2.txt

  branch=$(echo random-nightly-branch-$(date +%F)-$((RANDOM % 100 + 1)))

  if [[ $i -eq 1 ]]; then
    gh commit tmp/random1.txt tmp/random2.txt \
      -m "Randomly commit for nightly test." \
      -B $branch
  else
    gh commit -U -A \
      -m "Randomly commit for nightly test." \
      -B $branch
  fi

  git fetch
  git checkout $branch

  cat tmp/random1.txt
  cat tmp/random2.txt

  rm -rf tmp

  git switch -
  sleep 1

  repo_info=$(gh repo view --json owner,name -q '.owner.login + "/" + .name')
  gh api -X DELETE repos/$repo_info/git/refs/heads/$branch
  git checkout $current_branch
done

for i in {1..2}; do
  mkdir tmp > /dev/null 2>&1
  echo "RANDOM1" > tmp/random1.txt
  echo "RANDOM2" > tmp/random2.txt

  branch=$(echo random-nightly-branch-$(date +%F)-$((RANDOM % 100 + 1)))

  if [[ $i -eq 1 ]]; then
    gh commit tmp/random1.txt tmp/random2.txt \
      -m "Randomly commit for nightly test." \
      -B $branch
  else
    gh commit -U -A \
      -m "Randomly commit for nightly test." \
      -B $branch
  fi

  git fetch
  git checkout $branch

  cat tmp/random1.txt
  cat tmp/random2.txt

  rm -rf tmp

  git switch -
  sleep 1

  repo_info=$(gh repo view --json owner,name -q '.owner.login + "/" + .name')
  gh api -X DELETE repos/$repo_info/git/refs/heads/$branch
  git checkout $current_branch
done


for i in {1..2}; do
  mkdir tmp > /dev/null 2>&1
  echo "RANDOM1" > tmp/random1.txt
  echo "RANDOM2" > tmp/random2.txt

  branch=$(echo random-nightly-branch-$(date +%F)-$((RANDOM % 100 + 1)))
  headRef=$(echo random-nightly-branch-$(date +%F)-$((RANDOM % 1000 + 1)))

  if [[ $i -eq 1 ]]; then
    gh commit tmp/random1.txt tmp/random2.txt -P \
      -m "Randomly commit for nightly test." \
      -B $branch -H $headRef -T "Title of PR"
  else
    gh commit -U -A \
      -m "Randomly commit for nightly test." -P \
      -B $branch -H $headRef -T "Title of PR"
  fi

  git fetch
  git checkout $branch

  cat tmp/random1.txt
  cat tmp/random2.txt

  rm -rf tmp

  git switch -
  sleep 1

  pr_number=$(gh pr list --head "$headRef" --json number -q '.[0].number')
  echo $pr_number

  repo_info=$(gh repo view --json owner,name -q '.owner.login + "/" + .name')
  gh pr close $pr_number
  gh api -X DELETE repos/$repo_info/git/refs/heads/$branch
  gh api -X DELETE repos/$repo_info/git/refs/heads/$headRef
  git checkout $current_branch
done