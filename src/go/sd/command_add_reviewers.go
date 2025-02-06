package main

import (
	"flag"
	"log/slog"
	"os"
	sd "stackeddiff"
	ex "stackeddiff/execute"
	"time"
)

func createAddReviewersCommand() Command {
	flagSet := flag.NewFlagSet("add-reviewers", flag.ContinueOnError)
	var indicatorTypeString *string = addIndicatorFlag(flagSet)

	var whenChecksPass *bool = flagSet.Bool("when-checks-pass", true, "Poll until all checks pass before adding reviewers")
	var defaultPollFrequency time.Duration = 30 * time.Second
	var pollFrequency *time.Duration = flagSet.Duration("poll-frequency", defaultPollFrequency,
		"Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration")
	reviewers, silent, minChecks := addReviewersFlags(flagSet, "Falls back to "+ex.White+"PR_REVIEWERS"+ex.Reset+" environment variable.")

	return Command{
		FlagSet: flagSet,
		Summary: "Add reviewers to Pull Request on Github once its checks have passed",
		Description: "Add reviewers to Pull Request on Github once its checks have passed.\n" +
			"\n" +
			"If PR is marked as a Draft, it is first marked as \"Ready for Review\".",
		Usage: "sd " + flagSet.Name() + " [flags] [commitIndicator [commitIndicator]...]",
		OnSelected: func(command Command) {
			commitIndicators := flagSet.Args()
			if len(commitIndicators) == 0 {
				slog.Debug("Using main branch because commitIndicators is empty")
				commitIndicators = []string{sd.GetMainBranchOrDie()}
				*indicatorTypeString = string(sd.IndicatorTypeCommit)
			}
			if *reviewers == "" {
				*reviewers = os.Getenv("PR_REVIEWERS")
				if *reviewers == "" {
					commandError(flagSet, "reviewers not specified. Use reviewers flag or set PR_REVIEWERS environment variable", command.Usage)
				}
			}
			indicatorType := checkIndicatorFlag(command, indicatorTypeString)
			sd.AddReviewersToPr(commitIndicators, indicatorType, *whenChecksPass, *silent, *minChecks, *reviewers, *pollFrequency)
		}}
}
