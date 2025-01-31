package stackeddiff

import (
	"flag"
	"log"
	"os"

	ex "stackeddiff/execute"
)

func CreateNewPr(draft bool, featureFlag string, baseBranch string, branchInfo BranchInfo, logger *log.Logger) {
	RequireMainBranch()
	RequireCommitOnMain(branchInfo.CommitHash)

	var commitToBranchFrom string
	shouldPopStash := Stash("update-pr " + flag.Arg(0))
	if baseBranch == ex.GetMainBranch() {
		commitToBranchFrom = FirstOriginCommit(ex.GetMainBranch())
		logger.Println("Switching to branch", branchInfo.BranchName, "based off commit", commitToBranchFrom)
	} else {
		commitToBranchFrom = baseBranch
		logger.Println("Switching to branch", branchInfo.BranchName, "based off branch", baseBranch)
	}
	if commitToBranchFrom == "" {
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "--no-track", branchInfo.BranchName)
	} else {
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "--no-track", branchInfo.BranchName, commitToBranchFrom)
	}
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", branchInfo.BranchName)
	if commitToBranchFrom != "" {
		logger.Println("Cherry picking", branchInfo.CommitHash)
		cherryPickOutput, cherryPickError := ex.Execute(ex.ExecuteOptions{}, "git", "cherry-pick", branchInfo.CommitHash)
		if cherryPickError != nil {
			logger.Println(ex.Red+"Could not cherry-pick, aborting..."+ex.Reset, cherryPickOutput, cherryPickError)
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "cherry-pick", "--abort")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", ex.GetMainBranch())
			logger.Println("Deleting created branch", branchInfo.BranchName)
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-D", branchInfo.BranchName)
			PopStash(shouldPopStash)
			os.Exit(1)
		}
	}
	logger.Println("Pushing to remote")
	pushOutput, pushErr := ex.Execute(ex.ExecuteOptions{}, "git", "-c", "push.default=current", "push", "-f")
	if pushErr != nil {
		logger.Println(ex.Red+"Could not push: "+ex.Reset, pushOutput)
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", ex.GetMainBranch())
		logger.Println("Deleting created branch", branchInfo.BranchName)
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-D", branchInfo.BranchName)
		PopStash(shouldPopStash)
		os.Exit(1)
	}
	prText := GetPullRequestText(branchInfo.CommitHash, featureFlag)
	logger.Println("Creating PR via gh")
	createPrArgs := []string{"pr", "create", "--title", prText.Title, "--body", prText.Description, "--fill", "--base", baseBranch}
	if draft {
		createPrArgs = append(createPrArgs, "--draft")
	}
	createPrOutput, createPrErr := ex.Execute(ex.ExecuteOptions{}, "gh", createPrArgs...)
	if createPrErr != nil {
		logger.Println(ex.Red+"Could not create PR:"+ex.Reset, createPrOutput, createPrErr)
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", ex.GetMainBranch())
		logger.Println("Deleting created branch", branchInfo.BranchName)
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-D", branchInfo.BranchName)
		PopStash(shouldPopStash)
		os.Exit(1)
	} else {
		logger.Println("Created PR", createPrOutput)
	}
	if prViewOutput, prViewErr := ex.Execute(ex.ExecuteOptions{}, "gh", "pr", "view", "--web"); prViewErr != nil {
		logger.Println(ex.Red+"Could not open browser to PR:"+ex.Reset, prViewOutput, prViewErr)
	}
	logger.Println("Switching back to " + ex.GetMainBranch())
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", ex.GetMainBranch())
	PopStash(shouldPopStash)
	/*
	   This avoids this hint when using `git fetch && git-rebase origin/main` which is not appropriate for stacked diff workflow:
	   > hint: use --reapply-cherry-picks to include skipped commits
	   > hint: Disable this message with "git config advice.skippedCherryPicks false"
	*/
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "config", "advice.skippedCherryPicks", "false")
}
