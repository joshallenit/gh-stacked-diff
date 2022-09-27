#!/bin/bash
# Returns "username commit_hash branch_name" given the commit or PR number.

readonly arg=${1?commit or PR missing from command line arguments}
readonly result=$(git cat-file -t $arg 2>/dev/null)

if [ "$result" == "commit" ];
then
    # if the argument is a pr commit
    readonly pr_commit=$1
    readonly email=`git config user.email`
    readonly username=${email%@*}
    readonly branch_name="$username/$(git show --no-patch --format="%f" "$pr_commit")"
else
    # if the argument is a pr number
    readonly username=`gh auth status 2>&1 | grep "Logged in" | sed 's/.*Logged in.* \(.*\) .*/\1/'`
    readonly branch_name="$(gh pr view $arg --json headRefName -q '.headRefName')"
    readonly pr_commit="$(gh pr view $arg --json commits -q '[.commits[].oid] | first')"
fi

echo "$username $pr_commit $branch_name"
