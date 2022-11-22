#!/bin/bash
# Use iTerm for linking to work and then fix key mapping: https://apple.stackexchange.com/a/293988

set -euo pipefail

if [[ $# -ne 1 ]]; then
    echo "usage: $0 <pr-commit>" 2>&1
    exit
fi

read username pr_commit branch_name < <(git-get-commit-branch $1)

git switch "$branch_name"

# Cherrypick the latest commit to the PR branch.
if ! git cherry-pick main; then
    git cherry-pick --abort
    git switch main
    exit 1
fi

# Push the updated branch.
git push origin "$branch_name" || git push -f origin "$branch_name"

# Go back to main.
git switch main

# This allows for scripted (non-interactive) use of interactive rebase.
export GIT_SEQUENCE_EDITOR=/usr/bin/true

# In two steps, squash the latest commit into its PR commit.
# 1. Mark the commit as a fixup
git commit --amend --fixup="$pr_commit"
# 2. Use the autosquash feature of interactive rebase to perform the squash.
git rebase --interactive --autosquash "${pr_commit}^"
