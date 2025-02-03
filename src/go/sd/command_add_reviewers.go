package main

import (
	"flag"
	"log/slog"
	"os"
	sd "stackeddiff"
	ex "stackeddiff/execute"
	"time"
)

func CreateAddReviewersCommand() Command {
	flagSet := flag.NewFlagSet("add-reviewers", flag.ContinueOnError)
	var indicatorTypeString *string = AddIndicatorFlag(flagSet)

	var whenChecksPass *bool = flagSet.Bool("when-checks-pass", true, "Poll until all checks pass before adding reviewers")
	var defaultPollFrequency time.Duration = 30 * time.Second
	var pollFrequency *time.Duration = flagSet.Duration("poll-frequency", defaultPollFrequency,
		"Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration")
	reviewers, silent, minChecks := AddReviewersFlags(flagSet, "Falls back to "+ex.White+"PR_REVIEWERS"+ex.Reset+" environment variable.")

	return Command{
		FlagSet:     flagSet,
		Summary:     "Add reviewers to Pull Request on Github once its checks have passed",
		Description: "Mark a Draft PR as \"Ready for Review\" and automatically add reviewers.",
		Usage:       "sd " + flagSet.Name() + " [flags] [commitIndicator [commitIndicator]...]",
		OnSelected: func(command Command) {
			commitIndicators := flagSet.Args()
			if len(commitIndicators) == 0 {
				slog.Debug("Using main branch because commitIndicators is empty")
				commitIndicators = []string{ex.GetMainBranch()}
				*indicatorTypeString = string(sd.IndicatorTypeCommit)
			}
			if *reviewers == "" {
				*reviewers = os.Getenv("PR_REVIEWERS")
				if *reviewers == "" {
					commandError(flagSet, "reviewers not specified. Use reviewers flag or set PR_REVIEWERS environment variable", command.Usage)
				}
			}
			indicatorType := CheckIndicatorFlag(command, indicatorTypeString)
			sd.AddReviewersToPr(commitIndicators, indicatorType, *whenChecksPass, *silent, *minChecks, *reviewers, *pollFrequency)
		}}
}
