package main

import (
	"flag"
	"log"
)

func main() {
	RequireMainBranch()
	var draft bool
	flag.BoolVar(&draft, "draft", true, "Whether to create the PR as draft")
	flag.Parse()

	branchInfo := GetBranchInfo(flag.Arg(0))
	prText := GetPullRequestText(branchInfo.CommitHash)
	commitToBranchFrom := FirstOriginMainCommit("main")
	log.Println("Switching to branch", branchInfo.BranchName, "based off commit", commitToBranchFrom)
	ExecuteFailable("git", "branch", "--no-track", branchInfo.BranchName, commitToBranchFrom)
	Execute("git", "switch", branchInfo.BranchName)
	log.Println("Cherry picking", branchInfo.CommitHash)
	cherryPickOutput, cherryPickError := ExecuteFailable("git", "cherry-pick", branchInfo.CommitHash)
	if cherryPickError != nil {
		log.Println("Could not cherry-pick, aborting...", cherryPickOutput, cherryPickError)
		Execute("git", "cherry-pick", "--abort")
		Execute("git", "switch", "main")
		return
	}
	log.Println("Pushing to remote")
	Execute("git", "-c", "push.default=current", "push", "-f")
	log.Println("Creating PR via gh")
	var createPrOutput string
	if draft {
		createPrOutput = Execute("gh", "pr", "create", "--draft", "--title", prText.Title, "--body", prText.Description, "--fill")
	} else {
		createPrOutput = Execute("gh", "pr", "create", "--title", prText.Title, "--body", prText.Description, "--fill")
	}
	log.Println("Created PR", createPrOutput)
	Execute("gh", "pr", "view", "--web")
	log.Println("Switching back to main")
	Execute("git", "switch", "main")
	/*
	   This avoids this hint when using `git fetch && git-rebase origin/main` which is not appropriate for stacked diff workflow:
	   > hint: use --reapply-cherry-picks to include skipped commits
	   > hint: Disable this message with "git config advice.skippedCherryPicks false"
	*/
	Execute("git", "config", "advice.skippedCherryPicks", "false")
}
