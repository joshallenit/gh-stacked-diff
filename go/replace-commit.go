package main

import (
	"log"
	"os"
	"strings"
)

/*
replace the a given commit with a squashed version of the commits on the associated branch.
*/
func main() {
	branchInfo := GetBranchInfo(os.Args[1])
	commitsAfter := strings.Fields(Execute("git", "--no-pager", "log", branchInfo.CommitHash+"..HEAD", "--pretty=format:%h"))
	reverseArrayInPlace(commitsAfter)
	commitToDiffFrom := firstMainCommit(branchInfo.BranchName)
	diff := ExecuteWithOptions(ExecuteOptions{TrimSpace: false}, "git", "diff", "--binary", commitToDiffFrom, branchInfo.BranchName)

	log.Println("Resetting to ", branchInfo.CommitHash+"~1")
	Execute("git", "reset", "--hard", branchInfo.CommitHash+"~1")
	log.Println("Adding diff from commits ", branchInfo.BranchName)
	ExecuteWithOptions(ExecuteOptions{Stdin: &diff}, "git", "apply")
	Execute("git", "add", ".")
	commitSummary := Execute("git", "--no-pager", "show", "--no-patch", "--format=%s", branchInfo.CommitHash)
	Execute("git", "commit", "-m", commitSummary)
	log.Println("Cherry picking commits back on top ", commitsAfter)
	cherryPickArgs := make([]string, 2+len(commitsAfter))
	cherryPickArgs[0] = "cherry-pick"
	cherryPickArgs[1] = "--ff"
	for i, commit := range commitsAfter {
		cherryPickArgs[i+2] = commit
	}
	Execute("git", cherryPickArgs...)
}

func reverseArrayInPlace(array []string) {
	for i, j := 0, len(array)-1; i < j; i, j = i+1, j-1 {
		array[i], array[j] = array[j], array[i]
	}
}

// Returns first commit of the given branch that is on origin/main.
func firstMainCommit(branchName string) string {
	allNewCommits := strings.Fields(Execute("git", "--no-pager", "log", "origin/main.."+branchName, "--pretty=format:%h", "--abbrev-commit"))
	if len(allNewCommits) == 0 {
		log.Fatal("No commits on ", branchName, "other than what is on main, nothing to create a commit from")
	}
	return allNewCommits[len(allNewCommits)-1] + "~1"
}
