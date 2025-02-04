package main

import (
	"flag"
	sd "stackeddiff"
	ex "stackeddiff/execute"
	"time"
)

func CreateUpdateCommand() Command {
	flagSet := flag.NewFlagSet("update", flag.ContinueOnError)
	indicatorTypeString := AddIndicatorFlag(flagSet)
	reviewers, silent, minChecks := AddReviewersFlags(flagSet, "")
	return Command{
		FlagSet: flagSet,
		Summary: "Add commits from " + ex.GetMainBranch() + " to an existing PR",
		Description: "Add commits from local " + ex.GetMainBranch() + " branch to an existing PR.\n" +
			"\n" +
			"Can also add reviewers once PR checks have passed, see \"--reviewers\" flag.",
		Usage: "sd update [flags] <PR commitIndicator> [fixup commitIndicator (defaults to head commit) [fixup commitIndicator...]]",
		OnSelected: func(command Command) {
			if flagSet.NArg() == 0 {
				commandError(flagSet, "missing commitIndicator", command.Usage)
			}
			indicatorType := CheckIndicatorFlag(command, indicatorTypeString)
			var otherCommits []string
			if len(flagSet.Args()) > 1 {
				otherCommits = flagSet.Args()[1:]
			}
			destCommit := sd.GetBranchInfo(flagSet.Arg(0), indicatorType)
			sd.UpdatePr(destCommit, otherCommits, indicatorType)
			if *reviewers != "" {
				sd.AddReviewersToPr([]string{destCommit.CommitHash}, sd.IndicatorTypeCommit, true, *silent, *minChecks, *reviewers, 30*time.Second)
			}
		}}
}
