package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: update-pr <pr-commit> [fixup commit (defaults to top commit)] [other fixup commit...]")
		os.Exit(1)
	}
	RequireMainBranch()
	branchInfo := GetBranchInfo(os.Args[1])
	var commitsToCherryPick []string
	if len(os.Args) > 2 {
		commitsToCherryPick = os.Args[2:]
	} else {
		commitsToCherryPick = make([]string, 1)
		commitsToCherryPick[0] = Execute("git", "rev-parse", "--short", "HEAD")
	}
	log.Println("Switching to branch", branchInfo.BranchName)
	Execute("git", "switch", branchInfo.BranchName)
	log.Println("Fast forwarding in case there were any commits made via github web interface")
	Execute("git", "fetch", "origin", branchInfo.BranchName)
	Execute("git", "merge", "--ff-only", "origin/"+branchInfo.BranchName)

	log.Println("Cherry picking", commitsToCherryPick)
	cherryPickArgs := make([]string, 1+len(commitsToCherryPick))
	cherryPickArgs[0] = "cherry-pick"
	for i, commit := range commitsToCherryPick {
		cherryPickArgs[i+1] = commit
	}
	cherryPickOutput, cherryPickError := ExecuteFailable("git", cherryPickArgs...)
	if cherryPickError != nil {
		log.Println("Could not cherry-pick, aborting...", cherryPickArgs, cherryPickOutput, cherryPickError)
		Execute("git", "cherry-pick", "--abort")
		Execute("git", "switch", "main")
		os.Exit(1)
	}
	log.Println("Pushing to remote")
	Execute("git", "push", "origin", branchInfo.BranchName)
	log.Println("Switching back to main")
	Execute("git", "switch", "main")
	log.Println("Rebasing, marking as fixup", commitsToCherryPick, "for target", branchInfo.CommitHash)
	options := ExecuteOptions{EnvironmentVariables: make([]string, 1)}
	options.EnvironmentVariables[0] = "GIT_SEQUENCE_EDITOR=sequence-editor-mark-as-fixup " + branchInfo.CommitHash + " " + strings.Join(commitsToCherryPick, " ")
	ExecuteWithOptions(options, "git", "rebase", "-i", branchInfo.CommitHash+"^")
}
