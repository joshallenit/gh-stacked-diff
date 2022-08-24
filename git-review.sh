#!/bin/bash
# Sets the given PR to 'Ready For Review' and requests review from users defined
# in your PR_REVIEWERS environment variable.
#
# Reviewers can be specified with a comma delimiter:
# export PR_REVIEWERS=first-user,second-user,third-user

set -euo pipefail

readonly pullRequest=${1:-""}
readonly reviewers="${PR_REVIEWERS:-""}"
readonly isInt='^[0-9]+$'

if [[ $pullRequest =~ $isInt && $reviewers ]]
then
    gh pr ready $pullRequest
    gh pr edit $pullRequest --add-reviewer $reviewers
elif [[ -z $reviewers ]]
then
    echo "Add PR reviewers to your PR_REVIEWERS environment variable"
    echo "Example: export PR_REVIEWERS=first-user,second-user,third-user"
    echo ""
    echo "**Note: Add this to your shell rc file (~/.zshrc or ~/.bashrc)"
else
    echo "Please specify a GitHub Pull Request number"
    echo "Example: $0 55555"
fi