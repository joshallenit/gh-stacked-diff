package stackeddiff

import (
	"log"
	"strings"

	ex "stackeddiff/execute"
)

func ReplaceCommit(commitOrBranch string) {
	branchInfo := GetBranchInfo(commitOrBranch)
	shouldPopStash := Stash("replace-commit " + commitOrBranch)
	replaceCommitOfBranchInfo(branchInfo)
	PopStash(shouldPopStash)
}

// Replaces commit `branchInfo.CommitHash“ with the contents of branch `branchInfo.BranchName`
func replaceCommitOfBranchInfo(branchInfo BranchInfo) {
	commitsAfter := strings.Fields(strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "log", branchInfo.CommitHash+"..HEAD", "--pretty=format:%h")))
	reverseArrayInPlace(commitsAfter)
	commitToDiffFrom := FirstOriginCommit(branchInfo.BranchName)
	if commitToDiffFrom == "" {
		panic("replace-commit cannot be used to replace the root commit")
	}

	log.Println("Resetting to ", branchInfo.CommitHash+"~1")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "reset", "--hard", branchInfo.CommitHash+"~1")
	log.Println("Adding diff from commits ", branchInfo.BranchName)
	diff := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "diff", "--binary", commitToDiffFrom, branchInfo.BranchName)
	ex.ExecuteOrDie(ex.ExecuteOptions{Stdin: &diff}, "git", "apply")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "add", ".")
	commitSummary := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "--no-pager", "show", "--no-patch", "--format=%s", branchInfo.CommitHash)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "commit", "-m", strings.TrimSpace(commitSummary))
	if len(commitsAfter) != 0 {
		log.Println("Cherry picking commits back on top ", commitsAfter)
		cherryPickAndSkipAllEmpty(commitsAfter)
	}
}

func reverseArrayInPlace(array []string) {
	for i, j := 0, len(array)-1; i < j; i, j = i+1, j-1 {
		array[i], array[j] = array[j], array[i]
	}
}

func cherryPickAndSkipAllEmpty(commits []string) {
	cherryPickArgs := make([]string, 2+len(commits))
	cherryPickArgs[0] = "cherry-pick"
	cherryPickArgs[1] = "--ff"
	for i, commit := range commits {
		cherryPickArgs[i+2] = commit
	}
	out, err := ex.Execute(ex.ExecuteOptions{}, "git", cherryPickArgs...)
	for err != nil {
		if strings.Index(out, "git commit --allow-empty") != -1 {
			out, err = ex.Execute(ex.ExecuteOptions{}, "git", "cherry-pick", "--skip")
		} else {
			log.Fatal("Unexpected cherry-pick error", out, cherryPickArgs, err)
		}
	}
}
