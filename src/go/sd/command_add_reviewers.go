package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	sd "stackeddiff"
	ex "stackeddiff/execute"
	"time"
)

func CreateAddReviewersCommand() Command {
	flagSet := flag.NewFlagSet("add-reviewers", flag.ExitOnError)
	var indicatorTypeString *string = AddIndicatorFlag(flagSet)

	var whenChecksPass *bool = flagSet.Bool("when-checks-pass", true, "Poll until all checks pass before adding reviewers")
	var defaultPollFrequency time.Duration = 30 * time.Second
	var pollFrequency *time.Duration = flagSet.Duration("poll-frequency", defaultPollFrequency,
		"Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration")
	reviewers, silent, minChecks := AddReviewersFlags(flagSet, "Falls back to "+ex.White+"PR_REVIEWERS"+ex.Reset+" environment variable.")

	flagSet.Usage = func() {
		fmt.Fprint(flagSet.Output(),
			ex.Reset+"Mark a Draft PR as \"Ready for Review\" and automatically add reviewers.\n"+
				"\n"+
				"add-reviewers [flags] [commitIndicator [commitIndicator]...]\n"+
				"\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flagSet.PrintDefaults()
	}
	return Command{
		FlagSet:      flagSet,
		UsageSummary: "Add reviewers to Pull Request on Github once its checks have passed",
		OnSelected: func() {
			commitIndicators := flagSet.Args()
			if len(commitIndicators) == 0 {
				slog.Debug("Using main branch because commitIndicators is empty")
				commitIndicators = []string{ex.GetMainBranch()}
				*indicatorTypeString = string(sd.IndicatorTypeCommit)
			}
			if *reviewers == "" {
				*reviewers = os.Getenv("PR_REVIEWERS")
				if *reviewers == "" {
					fmt.Fprintln(flagSet.Output(), "error: reviewers not specified. Use reviewers flag or set PR_REVIEWERS environment variable")
					flagSet.Usage()
					os.Exit(1)
				}
			}
			indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
			sd.AddReviewersToPr(commitIndicators, indicatorType, *whenChecksPass, *silent, *minChecks, *reviewers, *pollFrequency)
		}}
}
