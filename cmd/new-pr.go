package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	RequireMainBranch()
	var draft bool
	var featureFlag string
	var baseBranch string
	var logFlags int
	flag.BoolVar(&draft, "draft", true, "Whether to create the PR as draft")
	flag.StringVar(&featureFlag, "feature-flag", "None", "Value for FEATURE_FLAG in PR description")
	flag.StringVar(&baseBranch, "base", GetMainBranch(), "Base branch for Pull Request")
	flag.IntVar(&logFlags, "logFlags", 0, "Log flags, see https://pkg.go.dev/log#pkg-constants")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			Reset+"Create a new PR with a cherry-pick of the given commit hash\n"+
				"\n"+
				"new-pr [flags] [commit hash to make PR for (default is top commit on "+GetMainBranch()+")]\n"+
				"\n"+
				White+"Note on Ticket Number:"+Reset+"\n"+
				"\n"+
				"If you prefix the Jira ticket to the git commit summary then the `Ticket` section of the PR description will be populated with it.\n"+
				"For example:\n"+
				"\"CONV-9999 Add new feature\"\n"+
				"\n"+
				White+"Note on Templates:"+Reset+"\n"+
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
				"- **FeatureFlag** - Value passed to feature-flag flag\n"+
				"\n"+
				"To change a template, copy the default from [cmd/config/](cmd/config/) into `~/.stacked-diff-workflow/` and modify.\n"+
				"\n"+
				White+"Flags:"+Reset+"\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(1)
	}
	log.SetFlags(logFlags)
	branchInfo := GetBranchInfo(flag.Arg(0))
	prText := GetPullRequestText(branchInfo.CommitHash, featureFlag)
	var commitToBranchFrom string
	shouldPopStash := false
	stashResult := Execute("git", "stash", "save", "-u", "before update-pr "+flag.Arg(0))
	if strings.HasPrefix(stashResult, "Saved working") {
		log.Println(stashResult)
		shouldPopStash = true
	}
	if baseBranch == GetMainBranch() {
		commitToBranchFrom = FirstOriginMainCommit(GetMainBranch())
		log.Println("Switching to branch", branchInfo.BranchName, "based off commit", commitToBranchFrom)
	} else {
		commitToBranchFrom = baseBranch
		log.Println("Switching to branch", branchInfo.BranchName, "based off branch", baseBranch)
	}
	Execute("git", "branch", "--no-track", branchInfo.BranchName, commitToBranchFrom)
	Execute("git", "switch", branchInfo.BranchName)
	log.Println("Cherry picking", branchInfo.CommitHash)
	cherryPickOutput, cherryPickError := ExecuteFailable("git", "cherry-pick", branchInfo.CommitHash)
	if cherryPickError != nil {
		log.Println(Red+"Could not cherry-pick, aborting..."+Reset, cherryPickOutput, cherryPickError)
		Execute("git", "cherry-pick", "--abort")
		Execute("git", "switch", GetMainBranch())
		log.Println("Deleting created branch", branchInfo.BranchName)
		Execute("git", "branch", "-D", branchInfo.BranchName)
		PopStash(shouldPopStash)
		os.Exit(1)
	}
	log.Println("Pushing to remote")
	pushOutput, pushErr := ExecuteFailable("git", "-c", "push.default=current", "push", "-f")
	if pushErr != nil {
		log.Println(Red+"Could not push: "+Reset, pushOutput)
		Execute("git", "switch", GetMainBranch())
		log.Println("Deleting created branch", branchInfo.BranchName)
		Execute("git", "branch", "-D", branchInfo.BranchName)
		PopStash(shouldPopStash)
		os.Exit(1)
	}
	log.Println("Creating PR via gh")
	createPrArgs := []string{"pr", "create", "--title", prText.Title, "--body", prText.Description, "--fill", "--base", baseBranch}
	if draft {
		createPrArgs = append(createPrArgs, "--draft")
	}
	createPrOutput, createPrErr := ExecuteFailable("gh", createPrArgs...)
	if createPrErr != nil {
		log.Println(Red+"Could not create PR:"+Reset, createPrOutput, createPrErr)
		Execute("git", "switch", GetMainBranch())
		log.Println("Deleting created branch", branchInfo.BranchName)
		Execute("git", "branch", "-D", branchInfo.BranchName)
		PopStash(shouldPopStash)
		os.Exit(1)
	} else {
		log.Println("Created PR", createPrOutput)
	}
	if prViewOutput, prViewErr := ExecuteFailable("gh", "pr", "view", "--web"); prViewErr != nil {
		log.Println(Red+"Could not open browser to PR:"+Reset, prViewOutput, prViewErr)
	}
	log.Println("Switching back to " + GetMainBranch())
	Execute("git", "switch", GetMainBranch())
	PopStash(shouldPopStash)
	/*
	   This avoids this hint when using `git fetch && git-rebase origin/main` which is not appropriate for stacked diff workflow:
	   > hint: use --reapply-cherry-picks to include skipped commits
	   > hint: Disable this message with "git config advice.skippedCherryPicks false"
	*/
	Execute("git", "config", "advice.skippedCherryPicks", "false")
}
