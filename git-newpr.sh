#!/bin/bash
# Modified from https://kastiglione.github.io/git/2020/09/11/git-stacked-commits.html
# Usage git-newpr <commit hash or PR number>
#   Note: PR number doesn't really make sense unless you are recreating an old PR
#   If argument is missing "main" is used.

set -euo pipefail

read username pr_commit branch_name < <(git-get-commit-branch ${1:-main})

readonly commit_summary="$(git --no-pager show --no-patch --format="%s" "$pr_commit")"
readonly commit_body="$(git --no-pager show --no-patch --format="%b" "$pr_commit")"
# Get the commit summary without the first ticket number (if any)
readonly pr_title="$(echo "$commit_summary" | pcregrep -o2 '^(\S+[[:digit:]]+ )?(.*)')"
readonly ticket="$(echo "$commit_summary" | pcregrep -o1 '^(\S+[[:digit:]]+ )?(.*)')"
readonly newline=$'\n'
readonly body="$commit_body$newline\
$newline\
<!-- <img src=\"XXXCOPYURLXXX\" alt=\"\" width=\"300\"/> -->$newline\
<!-- <video src=\"XXXCOPYURLXXX\" alt=\"\" width=\"300\"/> -->$newline\
$newline\
#### Ticket: [$ticket](https://jira.tinyspeck.com/browse/$ticket)$newline\
$newline\
#### Feature flag(s): \`None\`"

# Create the new branch and switch to it.
git branch --no-track "$branch_name" origin/main
git switch "$branch_name"

# Cherry pick the desired commit.
if ! git cherry-pick "$pr_commit"; then
    git cherry-pick --abort
    git switch main
    exit 1
fi

# Create a new remote branch by the same name.
git -c push.default=current push -f

# Use GitHub's cli to create the PR from the branch.
# See: https://github.com/cli/cli
gh pr create --draft --title "$pr_title" --body "$body" --fill

gh pr view --web

# Go back to main branch.
git switch main
