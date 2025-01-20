package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	sd "stackeddiff"

	ex "stackeddiff/execute"
)

func main() {
	var draft bool
	var featureFlag string
	var baseBranch string
	var logFlags int
	flag.BoolVar(&draft, "draft", true, "Whether to create the PR as draft")
	flag.StringVar(&featureFlag, "feature-flag", "None", "Value for FEATURE_FLAG in PR description")
	flag.StringVar(&baseBranch, "base", ex.GetMainBranch(), "Base branch for Pull Request")
	flag.IntVar(&logFlags, "logFlags", 0, "Log flags, see https://pkg.go.dev/log#pkg-constants")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
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
				"- **TicketNumber** - Jira ticket as parsed from the commit summary\n"+
				"- **Username** -  Name as parsed from git config email\n"+
				"- **CommitBody** - Body of the commit message\n"+
				"- **CommitSummary** - Summary line of the commit message\n"+
				"- **CommitSummaryCleaned** - Summary line of the commit message without spaces or special characters\n"+
				"- **CommitSummaryWithoutTicket** - Summary line of the commit message without the prefix of the ticket number\n"+
				"- **CodeOwners** - Code owners from CODEOWNERS of any matching files\n"+
				"- **FeatureFlag** - Value passed to feature-flag flag\n"+
				"\n"+
				"To change a template, copy the default from [src/config/](src/config/) into `~/.stacked-diff-workflow/` and modify.\n"+
				"\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(1)
	}
	log.SetFlags(logFlags)
	branchInfo := sd.GetBranchInfo(flag.Arg(0))
	sd.CreateNewPr(draft, featureFlag, baseBranch, logFlags, branchInfo, log.Default())
}
