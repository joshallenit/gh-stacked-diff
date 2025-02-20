package commands

import (
	"flag"
	sd "stackeddiff"
	"time"
)

func createUpdateCommand() Command {
	flagSet := flag.NewFlagSet("update", flag.ContinueOnError)
	indicatorTypeString := addIndicatorFlag(flagSet)
	reviewers, silent, minChecks := addReviewersFlags(flagSet, "")
	return Command{
		FlagSet: flagSet,
		Summary: "Add commits from " + sd.GetMainBranchForHelp() + " to an existing PR",
		Description: "Add commits from local " + sd.GetMainBranchForHelp() + " branch to an existing PR.\n" +
			"\n" +
			"Can also add reviewers once PR checks have passed, see \"--reviewers\" flag.",
		Usage: "sd " + flagSet.Name() + " [flags] <PR commitIndicator> [fixup commitIndicator (defaults to head commit) [fixup commitIndicator...]]",
		OnSelected: func(command Command) {
			if flagSet.NArg() == 0 {
				commandError(flagSet, "missing commitIndicator", command.Usage)
			}
			indicatorType := checkIndicatorFlag(command, indicatorTypeString)
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
