package main

import (
	"flag"
	sd "stackeddiff"
	ex "stackeddiff/execute"
	"time"
)

func CreateNewCommand() Command {
	flagSet := flag.NewFlagSet("new", flag.ContinueOnError)

	draft := flagSet.Bool("draft", true, "Whether to create the PR as draft")
	featureFlag := flagSet.String("feature-flag", "", "Value for FEATURE_FLAG in PR description")
	baseBranch := flagSet.String("base", ex.GetMainBranch(), "Base branch for Pull Request")

	reviewers, silent, minChecks := AddReviewersFlags(flagSet, "")

	indicatorTypeString := AddIndicatorFlag(flagSet)

	return Command{
		FlagSet: flagSet,
		Summary: "Create a new pull request from a commit on main",
		Description: "Create a new PR with a cherry-pick of the given commit indicator.\n" +
			"\n" +
			"This command first creates an associated branch, (with a name based\n" +
			"on the commit summary), and then uses Github CLI to create a PR.\n" +
			"\n" +
			"Can also add reviewers once PR checks have passed, see \"--reviewers\" flag.",
		Usage: "sd new [flags] [commitIndicator (default is HEAD commit on " + ex.GetMainBranch() + ")]\n" +
			"\n" +
			ex.White + "Ticket Number:" + ex.Reset + "\n" +
			"\n" +
			"If you prefix a (Jira-like formatted) ticket number to the git commit\n" +
			"summary then the \"Ticket\" section of the PR description will be \n" +
			"populated with it.\n" +
			"\n" +
			"For example:\n" +
			"\n" +
			"\"CONV-9999 Add new feature\"\n" +
			"\n" +
			ex.White + "Templates:" + ex.Reset + "\n" +
			"\n" +
			"The Pull Request Title, Body (aka Description), and Branch Name are\n" +
			"created from golang templates.\n" +
			"\n" +
			"The default templates are:\n" +
			"\n" +
			"   branch-name.template:      src/go/config/branch-name.template\n" +
			"   pr-description.template:   src/go/config/pr-description.template\n" +
			"   pr-title.template:         src/go/config/pr-title.template\n" +
			"\n" +
			"To change a template, copy the default from src/go/config/ into\n" +
			"~/.stacked-diff-workflow/ and modify contents.\n" +
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
			"   UsernameCleaned              Username with dots (.) to dashes (-).\n",
		OnSelected: func(command Command) {
			if flagSet.NArg() > 1 {
				commandError(flagSet, "too many arguments", command.Usage)
			}

			indicatorType := CheckIndicatorFlag(command, indicatorTypeString)
			branchInfo := sd.GetBranchInfo(flagSet.Arg(0), indicatorType)
			sd.CreateNewPr(*draft, *featureFlag, *baseBranch, branchInfo)
			if *reviewers != "" {
				sd.AddReviewersToPr([]string{branchInfo.CommitHash}, sd.IndicatorTypeCommit, true, *silent, *minChecks, *reviewers, 30*time.Second)
			}
		}}
}
