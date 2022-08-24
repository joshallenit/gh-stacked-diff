#!/bin/bash
# Merges a PR by adding it to the merge queue

set -euo pipefail

readonly pullRequest=${1:-""}
readonly isInt='^[0-9]+$'

if [[ $pullRequest =~ $isInt ]]
then
    gh pr edit $pullRequest --add-label "mergequeue:queued"
else
    echo "Please specify a GitHub Pull Request number"
    echo "Example: $0 55555"
fi