#!/bin/bash

# Checks out the branch associated with the PR or commit.
# Usage: git-checkout <commit hash or PR number>

set -euo pipefail

read username pr_commit branch_name < <(git-get-commit-branch ${1:-main})

git checkout $branch_name
