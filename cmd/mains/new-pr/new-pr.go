package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	sd "stacked-diff-workflow/cmd/stacked-diff"
)

func main() {
	sd.RequireMainBranch()
	var draft bool
	var featureFlag string
	var baseBranch string
	var logFlags int
	flag.BoolVar(&draft, "draft", true, "Whether to create the PR as draft")
	flag.StringVar(&featureFlag, "feature-flag", "None", "Value for FEATURE_FLAG in PR description")
	flag.StringVar(&baseBranch, "base", sd.GetMainBranch(), "Base branch for Pull Request")
	flag.IntVar(&logFlags, "logFlags", 0, "Log flags, see https://pkg.go.dev/log#pkg-constants")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			sd.Reset+"Create a new PR with a cherry-pick of the given commit hash\n"+
				"\n"+
				"new-pr [flags] [commit hash to make PR for (default is top commit on "+sd.GetMainBranch()+")]\n"+
				"\n"+
				sd.White+"Note on Ticket Number:"+sd.Reset+"\n"+
				"\n"+
				"If you prefix the Jira ticket to the git commit summary then the `Ticket` section of the PR description will be populated with it.\n"+
				"For example:\n"+
				"\"CONV-9999 Add new feature\"\n"+
				"\n"+
				sd.White+"Note on Templates:"+sd.Reset+"\n"+
				"\n"+
				"The Pull Request Title, Body (aka Description), and Branch Name are created from golang templates. The defaults are:\n"+
				"\n"+
				"- branch-name.template - cmd/config/branch-name.template\n"+
				"- pr-description.template - cmd/config/pr-description.template\n"+
				"- pr-title.template - cmd/config/pr-title.template\n"+
				"\n"+
				"The possible values for the templates are:\n"+
				"\n"+
				"- **TicketNumber** - Jira ticket as parsed from the commit summary\n"+
				"- **Username** -  Name as parsed from git config email\n"+
				"- **CommitBody** - Body of the commit message\n"+
				"- **CommitSummary** - Summary line of the commit message\n"+
				"- **CommitSummaryCleaned** - Summary line of the commit message without spaces or special characters\n"+
				"- **CommitSummaryWithoutTicket** - Summary line of the commit message without the prefix of the ticket number\n"+
				"- **CodeOwners** - Code owners from CODEOWNERS of any matching files\n"+
				"- **FeatureFlag** - Value passed to feature-flag flag\n"+
				"\n"+
				"To change a template, copy the default from [cmd/config/](cmd/config/) into `~/.stacked-diff-workflow/` and modify.\n"+
				"\n"+
				sd.White+"Flags:"+sd.Reset+"\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(1)
	}
	log.SetFlags(logFlags)
	branchInfo := sd.GetBranchInfo(flag.Arg(0))
	var commitToBranchFrom string
	shouldPopStash := sd.Stash("update-pr " + flag.Arg(0))
	if baseBranch == sd.GetMainBranch() {
		commitToBranchFrom = sd.FirstOriginMainCommit(sd.GetMainBranch())
		log.Println("Switching to branch", branchInfo.BranchName, "based off commit", commitToBranchFrom)
	} else {
		commitToBranchFrom = baseBranch
		log.Println("Switching to branch", branchInfo.BranchName, "based off branch", baseBranch)
	}
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "branch", "--no-track", branchInfo.BranchName, commitToBranchFrom)
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "switch", branchInfo.BranchName)
	log.Println("Cherry picking", branchInfo.CommitHash)
	cherryPickOutput, cherryPickError := sd.Execute(sd.ExecuteOptions{}, "git", "cherry-pick", branchInfo.CommitHash)
	if cherryPickError != nil {
		log.Println(sd.Red+"Could not cherry-pick, aborting..."+sd.Reset, cherryPickOutput, cherryPickError)
		sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "cherry-pick", "--abort")
		sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "switch", sd.GetMainBranch())
		log.Println("Deleting created branch", branchInfo.BranchName)
		sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "branch", "-D", branchInfo.BranchName)
		sd.PopStash(shouldPopStash)
		os.Exit(1)
	}
	log.Println("Pushing to remote")
	pushOutput, pushErr := sd.Execute(sd.ExecuteOptions{}, "git", "-c", "push.default=current", "push", "-f")
	if pushErr != nil {
		log.Println(sd.Red+"Could not push: "+sd.Reset, pushOutput)
		sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "switch", sd.GetMainBranch())
		log.Println("Deleting created branch", branchInfo.BranchName)
		sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "branch", "-D", branchInfo.BranchName)
		sd.PopStash(shouldPopStash)
		os.Exit(1)
	}
	prText := sd.GetPullRequestText(branchInfo.CommitHash, featureFlag)
	log.Println("Creating PR via gh")
	createPrArgs := []string{"pr", "create", "--title", prText.Title, "--body", prText.Description, "--fill", "--base", baseBranch}
	if draft {
		createPrArgs = append(createPrArgs, "--draft")
	}
	createPrOutput, createPrErr := sd.Execute(sd.ExecuteOptions{}, "gh", createPrArgs...)
	if createPrErr != nil {
		log.Println(sd.Red+"Could not create PR:"+sd.Reset, createPrOutput, createPrErr)
		sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "switch", sd.GetMainBranch())
		log.Println("Deleting created branch", branchInfo.BranchName)
		sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "branch", "-D", branchInfo.BranchName)
		sd.PopStash(shouldPopStash)
		os.Exit(1)
	} else {
		log.Println("Created PR", createPrOutput)
	}
	if prViewOutput, prViewErr := sd.Execute(sd.ExecuteOptions{}, "gh", "pr", "view", "--web"); prViewErr != nil {
		log.Println(sd.Red+"Could not open browser to PR:"+sd.Reset, prViewOutput, prViewErr)
	}
	log.Println("Switching back to " + sd.GetMainBranch())
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "switch", sd.GetMainBranch())
	sd.PopStash(shouldPopStash)
	/*
	   This avoids this hint when using `git fetch && git-rebase origin/main` which is not appropriate for stacked diff workflow:
	   > hint: use --reapply-cherry-picks to include skipped commits
	   > hint: Disable this message with "git config advice.skippedCherryPicks false"
	*/
	sd.ExecuteOrDie(sd.ExecuteOptions{}, "git", "config", "advice.skippedCherryPicks", "false")
}
