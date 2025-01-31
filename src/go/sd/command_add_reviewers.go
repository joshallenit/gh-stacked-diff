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
	var reviewers string

	var indicatorTypeString *string = AddIndicatorFlag(flagSet)

	flagSet.StringVar(&reviewers, "reviewers", "", "Comma-separated list of Github usernames to add as reviewers. "+
		"Falls back to "+ex.White+"PR_REVIEWERS"+ex.Reset+" environment variable. "+
		"You can specify more than one reviewer using a comma-delimited string.")
	var whenChecksPass *bool = flagSet.Bool("when-checks-pass", true, "Poll until all checks pass before adding reviewers")
	var defaultPollFrequency time.Duration = 30 * time.Second
	var pollFrequency *time.Duration = flagSet.Duration("poll-frequency", defaultPollFrequency,

		"Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration")
	var silent bool
	var minChecks int

	flagSet.BoolVar(&silent, "silent", false, "Whether to use voice output")
	flagSet.IntVar(&minChecks, "min-checks", 4,
		"Minimum number of checks to wait for before verifying that checks have passed. "+
			"It takes some time for checks to be added to a PR by Github, "+
			"and if you add-reviewers too soon it will think that they have all passed.")

	flagSet.Usage = func() {
		fmt.Fprint(os.Stderr,
			ex.Reset+"Mark a Draft PR as \"Ready for Review\" and automatically add reviewers.\n"+
				"\n"+
				"add-reviewers [flags] [commitIndicator [commitIndicator]...]\n"+
				"\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flagSet.PrintDefaults()
	}
	return Command{FlagSet: flagSet, OnSelected: func() {
		commitIndicators := flagSet.Args()
		if len(commitIndicators) == 0 {
			slog.Debug("Using main branch because commitIndicators is empty")
			commitIndicators = []string{ex.GetMainBranch()}
			*indicatorTypeString = string(sd.IndicatorTypeCommit)
		}
		if reviewers == "" {
			reviewers = os.Getenv("PR_REVIEWERS")
			if reviewers == "" {
				fmt.Fprintln(os.Stderr, "Use reviewers flag or set PR_REVIEWERS environment variable")
				flagSet.Usage()
				os.Exit(1)
			}
		}
		indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
		sd.AddReviewersToPr(commitIndicators, indicatorType, *whenChecksPass, silent, minChecks, reviewers, *pollFrequency)
	}}
}
