package main

import (
	"flag"
	"log"
)

func main() {
	RequireMainBranch()
	var draft bool
	var featureFlag string
	var baseBranch string
	flag.BoolVar(&draft, "draft", true, "Whether to create the PR as draft")
	flag.StringVar(&featureFlag, "feature-flag", "None", "Value for FEATURE_FLAG in PR description")
	flag.StringVar(&baseBranch, "base", "main", "Base branch for Pull Request")
	flag.Parse()

	branchInfo := GetBranchInfo(flag.Arg(0))
	prText := GetPullRequestText(branchInfo.CommitHash, featureFlag)
	var commitToBranchFrom string
	if baseBranch == "main" {
		commitToBranchFrom = FirstOriginMainCommit("main")
		log.Println("Switching to branch", branchInfo.BranchName, "based off commit", commitToBranchFrom)
	} else {
		commitToBranchFrom = baseBranch
		log.Println("Switching to branch", branchInfo.BranchName, "based off branch", baseBranch)
	}
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
	createPrArgs := []string{"pr", "create", "--title", prText.Title, "--body", prText.Description, "--fill", "--base", baseBranch}
	if draft {
		createPrArgs = append(createPrArgs, "--draft")
	}
	createPrOutput := Execute("gh", createPrArgs...)
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
