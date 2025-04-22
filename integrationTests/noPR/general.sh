#!/bin/bash

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
  sleep 10

  repo_info=$(gh repo view --json owner,name -q '.owner.login + "/" + .name')
  gh api -X DELETE repos/$repo_info/git/refs/heads/$branch
done
