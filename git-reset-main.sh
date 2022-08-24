#!/bin/bash
# Resets the squashed commit on main

set -euo pipefail

readonly pr_commit="${1:-main}"

readonly email=`git config user.email`
readonly username=${email%@*}

# Autogenerate a branch name based on the commit subject.
readonly branch_name="$username/$(git show --no-patch --format="%f" "$pr_commit")"

git switch "$branch_name"

git branch -D forsquashing || true
git checkout -b forsquashing

GIT_SEQUENCE_EDITOR="go run /usr/local/bin/squash-commits.go \$1" git rebase -i HEAD~`gitlog | wc -l | xargs`

readonly squashed_commit=`git rev-parse HEAD`

git switch main

readonly commits_after=`git --no-pager log $pr_commit..HEAD --pretty=format:"%h " | tail -r | tr '\n' ' '`

echo Resetting to $pr_commit~1 
git reset --hard $pr_commit~1

echo Adding new squashed commit $squashed_commit
git cherry-pick $squashed_commit

echo Cherry picking commits back on top $commits_after

git cherry-pick --ff $commits_after
