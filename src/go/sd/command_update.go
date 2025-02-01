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

func CreateUpdateCommand() Command {
	flagSet := flag.NewFlagSet("update", flag.ExitOnError)
	indicatorTypeString := AddIndicatorFlag(flagSet)
	reviewers, silent, minChecks := AddReviewersFlags(flagSet, "")
	flagSet.Usage = func() {
		fmt.Fprint(flagSet.Output(),
			ex.Reset+"Add one or more commits to a PR.\n"+
				"\n"+
				"sd update [flags] <pr-commit> [fixup commit (defaults to top commit)] [other fixup commit...]\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flagSet.PrintDefaults()
	}

	return Command{
		FlagSet:      flagSet,
		UsageSummary: "Add commits from main to an existing PR",
		OnSelected: func() {
			if flagSet.NArg() == 0 {
				flagSet.Usage()
				os.Exit(1)
			}
			indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
			var otherCommits []string
			if len(flagSet.Args()) > 1 {
				otherCommits = flagSet.Args()[1:]
			}
			destCommit := sd.GetBranchInfo(flagSet.Arg(0), indicatorType)
			sd.UpdatePr(destCommit, otherCommits, indicatorType, log.Default())
			if *reviewers != "" {
				sd.AddReviewersToPr([]string{destCommit.CommitHash}, sd.IndicatorTypeCommit, true, *silent, *minChecks, *reviewers, 30*time.Second)
			}
		}}
}
