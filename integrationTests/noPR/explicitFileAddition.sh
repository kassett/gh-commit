#!/bin/bash

mkdir tmp > /dev/null 2>&1
echo "RANDOM1" > tmp/random1.txt
echo "RANDOM2" > tmp/random2.txt

branch=$(echo random-nightly-branch-$(date +%F)-$((RANDOM % 100 + 1)))

gh commit tmp/random1.txt tmp/random2.txt \
    -m "Randomly commit for nightly test." \
    -B $branch

rm -rf tmp

git fetch
git checkout $branch

cat tmp/random1.txt
cat tmp/random2.txt

git switch -
sleep 10

repo_info=$(gh repo view --json owner,name -q '.owner.login + "/" + .name')
gh api -X DELETE repos/$repo_info/git/refs/heads/$branch

