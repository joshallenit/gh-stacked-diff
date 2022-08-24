#!/bin/bash
# Lists all Pull Requests you have open. You must first log-in using 'gh auth login'

set -euo pipefail

username=`gh auth status 2>&1 | grep "Logged in" | sed 's/.*Logged in.* \(.*\) .*/\1/'`

if [[ $username ]]
then
    eval "gh pr list -A $username"
else
    echo "Failed to list your open PRs. Are you authenticated? - gh auth login"
fi
