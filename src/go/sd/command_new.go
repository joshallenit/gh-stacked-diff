package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	sd "stackeddiff"
	ex "stackeddiff/execute"
	"time"
)

func CreateNewCommand() Command {
	flagSet := flag.NewFlagSet("new", flag.ExitOnError)

	var draft bool
	var featureFlag string
	var baseBranch string
	flagSet.BoolVar(&draft, "draft", true, "Whether to create the PR as draft")
	flagSet.StringVar(&featureFlag, "feature-flag", "", "Value for FEATURE_FLAG in PR description")
	flagSet.StringVar(&baseBranch, "base", ex.GetMainBranch(), "Base branch for Pull Request")

	var reviewers string
	flagSet.StringVar(&reviewers, "reviewers", "", "Comma-separated list of Github usernames to add as reviewers once checks have passed.")
	var silent bool
	var minChecks int
	flagSet.BoolVar(&silent, "silent", false, "Whether to use voice output (false) or be silent (true) to notify that reviewers have been added.")
	flagSet.IntVar(&minChecks, "min-checks", 4,
		"Minimum number of checks to wait for before verifying that checks have passed. "+
			"It takes some time for checks to be added to a PR by Github, "+
			"and if you add-reviewers too soon it will think that they have all passed.")

	indicatorTypeString := AddIndicatorFlag(flagSet)

	flagSet.Usage = func() {
		fmt.Fprint(os.Stderr,
			ex.Reset+"Create a new PR with a cherry-pick of the given commit hash\n"+
				"\n"+
				"new-pr [flags] [commit hash to make PR for (default is top commit on "+ex.GetMainBranch()+")]\n"+
				"\n"+
				ex.White+"Note on Ticket Number:"+ex.Reset+"\n"+
				"\n"+
				"If you prefix the Jira ticket to the git commit summary then the `Ticket` section of the PR description will be populated with it.\n"+
				"For example:\n"+
				"\"CONV-9999 Add new feature\"\n"+
				"\n"+
				ex.White+"Note on Templates:"+ex.Reset+"\n"+
				"\n"+
				"The Pull Request Title, Body (aka Description), and Branch Name are created from golang templates. The defaults are:\n"+
				"\n"+
				"- branch-name.template - src/config/branch-name.template\n"+
				"- pr-description.template - src/config/pr-description.template\n"+
				"- pr-title.template - src/config/pr-title.template\n"+
				"\n"+
				"The possible values for the templates are:\n"+
				"\n"+
				"- **CommitBody** - Body of the commit message\n"+
				"- **CommitSummary** - Summary line of the commit message\n"+
				"- **CommitSummaryCleaned** - Summary line of the commit message without spaces or special characters\n"+
				"- **CommitSummaryWithoutTicket** - Summary line of the commit message without the prefix of the ticket number\n"+
				"- **FeatureFlag** - Value passed to feature-flag flag\n"+
				"- **TicketNumber** - Jira ticket as parsed from the commit summary\n"+
				"- **Username** -  Name as parsed from git config email\n"+
				"\n"+
				"To change a template, copy the default from [src/config/](src/config/) into `~/.stacked-diff-workflow/` and modify.\n"+
				"\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flagSet.PrintDefaults()
	}

	return Command{FlagSet: flagSet, OnSelected: func() {
		if flagSet.NArg() > 1 {
			fmt.Fprintln(os.Stderr, "Too many arguments")
			flagSet.Usage()
			os.Exit(1)
		}

		indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
		branchInfo := sd.GetBranchInfo(flagSet.Arg(0), indicatorType)
		sd.CreateNewPr(draft, featureFlag, baseBranch, branchInfo, log.Default())
		if reviewers != "" {
			sd.AddReviewersToPr([]string{branchInfo.CommitHash}, sd.IndicatorTypeCommit, true, silent, minChecks, reviewers, 30*time.Second)
		}
	}}
}
