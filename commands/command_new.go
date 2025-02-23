package commands

import (
	"time"

	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/fatih/color"
	ex "github.com/joshallenit/stacked-diff/v2/execute"
	"github.com/joshallenit/stacked-diff/v2/templates"

	"github.com/joshallenit/stacked-diff/v2/util"
)

func createNewCommand() Command {
	flagSet := flag.NewFlagSet("new", flag.ContinueOnError)

	draft := flagSet.Bool("draft", true, "Whether to create the PR as draft")
	featureFlag := flagSet.String("feature-flag", "", "Value for FEATURE_FLAG in PR description")
	baseBranch := flagSet.String("base", util.GetMainBranchOrDie(), "Base branch for Pull Request")

	reviewers, silent, minChecks := addReviewersFlags(flagSet, "")

	indicatorTypeString := addIndicatorFlag(flagSet)

	return Command{
		FlagSet: flagSet,
		Summary: "Create a new pull request from a commit on main",
		Description: "Create a new PR with a cherry-pick of the given commit indicator.\n" +
			"\n" +
			"This command first creates an associated branch, (with a name based\n" +
			"on the commit summary), and then uses Github CLI to create a PR.\n" +
			"\n" +
			"Can also add reviewers once PR checks have passed, see \"--reviewers\" flag.",
		Usage: "sd new [flags] [commitIndicator (default is HEAD commit on " + util.GetMainBranchForHelp() + ")]\n" +
			"\n" +
			color.HiWhiteString("Ticket Number:") + "\n" +
			"\n" +
			"If you prefix a (Jira-like formatted) ticket number to the git commit\n" +
			"summary then the \"Ticket\" section of the PR description will be \n" +
			"populated with it.\n" +
			"\n" +
			"For example:\n" +
			"\n" +
			"\"CONV-9999 Add new feature\"\n" +
			"\n" +
			color.HiWhiteString("Templates:") + "\n" +
			"\n" +
			"The Pull Request Title, Body (aka Description), and Branch Name are\n" +
			"created from golang templates.\n" +
			"\n" +
			"The default templates are:\n" +
			"\n" +
			"   branch-name.template:      templates/config/branch-name.template\n" +
			"   pr-description.template:   templates/config/pr-description.template\n" +
			"   pr-title.template:         templates/config/pr-title.template\n" +
			"\n" +
			"To change a template, copy the default from templates/config/ into\n" +
			"~/.stacked-diff/ and modify contents.\n" +
			"\n" +
			"The possible values for the templates are:\n" +
			"\n" +
			"   CommitBody                   Body of the commit message\n" +
			"   CommitSummary                Summary line of the commit message\n" +
			"   CommitSummaryCleaned         Summary line of the commit message without\n" +
			"                                spaces or special characters\n" +
			"   CommitSummaryWithoutTicket   Summary line of the commit message without\n" +
			"                                the prefix of the ticket number\n" +
			"   FeatureFlag                  Value passed to feature-flag flag\n" +
			"   TicketNumber                 Jira ticket as parsed from the commit summary\n" +
			"   Username                     Name as parsed from git config email.\n" +
			"   UsernameCleaned              Username with dots (.) converted to dashes (-).\n",
		OnSelected: func(command Command) {
			if flagSet.NArg() > 1 {
				commandError(flagSet, "too many arguments", command.Usage)
			}

			indicatorType := checkIndicatorFlag(command, indicatorTypeString)
			gitLog := templates.GetBranchInfo(flagSet.Arg(0), indicatorType)
			createNewPr(*draft, *featureFlag, *baseBranch, gitLog)
			if *reviewers != "" {
				addReviewersToPr([]string{gitLog.Commit}, templates.IndicatorTypeCommit, true, *silent, *minChecks, *reviewers, 30*time.Second)
			}
		}}
}

// Creates a new pull request via Github CLI.
func createNewPr(draft bool, featureFlag string, baseBranch string, gitLog templates.GitLog) {
	util.RequireMainBranch()
	templates.RequireCommitOnMain(gitLog.Commit)

	var commitToBranchFrom string
	shouldPopStash := util.Stash("sd new " + flag.Arg(0))
	if baseBranch == util.GetMainBranchOrDie() {
		commitToBranchFrom = util.FirstOriginMainCommit(util.GetMainBranchOrDie())
		slog.Info(fmt.Sprint("Switching to branch ", gitLog.Branch, " based off commit ", commitToBranchFrom))
	} else {
		commitToBranchFrom = baseBranch
		slog.Info(fmt.Sprint("Switching to branch ", gitLog.Branch, " based off branch ", baseBranch))
	}
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "--no-track", gitLog.Branch, commitToBranchFrom)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", gitLog.Branch)
	slog.Info(fmt.Sprint("Cherry picking ", gitLog.Commit))
	cherryPickOutput, cherryPickError := ex.Execute(ex.ExecuteOptions{}, "git", "cherry-pick", gitLog.Commit)
	if cherryPickError != nil {
		slog.Info(fmt.Sprint(color.RedString("Could not cherry-pick, aborting... "), cherryPickOutput, " ", cherryPickError))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "cherry-pick", "--abort")
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", util.GetMainBranchOrDie())
		slog.Info(fmt.Sprint("Deleting created branch ", gitLog.Branch))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-D", gitLog.Branch)
		util.PopStash(shouldPopStash)
		os.Exit(1)
	}
	slog.Info("Pushing to remote")
	// -u is required because in newer versions of Github CLI the upstream must be set.
	pushOutput, pushErr := ex.Execute(ex.ExecuteOptions{}, "git", "-c", "push.default=current", "push", "-f", "-u")
	if pushErr != nil {
		slog.Info(fmt.Sprint(color.RedString("Could not push: "), " ", pushOutput))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", util.GetMainBranchOrDie())
		slog.Info(fmt.Sprint("Deleting created branch ", gitLog.Branch))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-D", gitLog.Branch)
		util.PopStash(shouldPopStash)
		os.Exit(1)
	}
	prText := templates.GetPullRequestText(gitLog.Commit, featureFlag)
	slog.Info("Creating PR via gh")
	createPrOutput, createPrErr := createPr(prText, baseBranch, draft)
	if createPrErr != nil {
		slog.Info(fmt.Sprint(color.RedString("Could not create PR: "), createPrOutput, " ", createPrErr))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", util.GetMainBranchOrDie())
		slog.Info(fmt.Sprint("Deleting created branch ", gitLog.Branch))
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "branch", "-D", gitLog.Branch)
		util.PopStash(shouldPopStash)
		os.Exit(1)
	} else {
		slog.Info(fmt.Sprint("Created PR ", createPrOutput))
	}
	if prViewOutput, prViewErr := ex.Execute(ex.ExecuteOptions{}, "gh", "pr", "view", "--web"); prViewErr != nil {
		slog.Info(fmt.Sprint(color.RedString("Could not open browser to PR: "), prViewOutput, " ", prViewErr))
	}
	slog.Info(fmt.Sprint("Switching back to " + util.GetMainBranchOrDie()))
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", util.GetMainBranchOrDie())
	util.PopStash(shouldPopStash)
	/*
	   This avoids this hint when using `git fetch && git-rebase origin/main` which is not appropriate for stacked diff workflow:
	   > hint: use --reapply-cherry-picks to include skipped commits
	   > hint: Disable this message with "git config advice.skippedCherryPicks false",
	*/
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "config", "advice.skippedCherryPicks", "false")
}

func createPr(prText templates.PullRequestText, baseBranch string, draft bool) (string, error) {
	createPrArgsNoDraft := []string{"pr", "create", "--title", prText.Title, "--body", prText.Description, "--fill", "--base", baseBranch}
	createPrArgs := createPrArgsNoDraft
	if draft {
		createPrArgs = append(createPrArgs, "--draft")
	}
	createPrOutput, createPrErr := ex.Execute(ex.ExecuteOptions{}, "gh", createPrArgs...)
	if createPrErr != nil && draft && strings.Contains(createPrOutput, "Draft pull requests are not supported") {
		slog.Warn("Draft PRs not supported, trying again without draft.\nUse \"--draft=false\" to avoid this warning.")
		createPrOutput, createPrErr = ex.Execute(ex.ExecuteOptions{}, "gh", createPrArgsNoDraft...)
	}
	return createPrOutput, createPrErr
}
