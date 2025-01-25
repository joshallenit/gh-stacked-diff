package main

import (
	"flag"
	"fmt"
	"os"
	sd "stackeddiff"
	ex "stackeddiff/execute"
	"time"
)

func CreateAddReviewersCommand() Command {
	flagSet := flag.NewFlagSet("add-reviewers", flag.ExitOnError)
	var reviewers string

	flagSet.StringVar(&reviewers, "reviewers", "", "Comma-separated list of Github usernames to add as reviewers. "+
		"Falls back to "+ex.White+"PR_REVIEWERS"+ex.Reset+" environment variable. "+
		"You can specify more than one reviewer using a comma-delimited string.")
	var whenChecksPass bool
	var pollFrequency time.Duration
	var defaultPollFrequency time.Duration = 30 * time.Second
	var silent bool
	var minChecks int
	flagSet.BoolVar(&whenChecksPass, "when-checks-pass", true, "Poll until all checks pass before adding reviewers")
	flagSet.DurationVar(&pollFrequency, "poll-frequency", defaultPollFrequency,
		"Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration")
	flagSet.BoolVar(&silent, "silent", false, "Whether to use voice output")
	flagSet.IntVar(&minChecks, "min-checks", 4,
		"Minimum number of checks to wait for before verifying that checks have passed. "+
			"It takes some time for checks to be added to a PR by Github, "+
			"and if you add-reviewers too soon it will think that they have all passed.")
	flagSet.Usage = func() {
		fmt.Fprint(os.Stderr,
			ex.Reset+"Mark a Draft PR as \"Ready for Review\" and automatically add reviewers.\n"+
				"\n"+
				"add-reviewers [flags] <commit hash or pull request number>\n"+
				"\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flagSet.PrintDefaults()
	}
	return Command{FlagSet: flagSet, OnSelected: func() {
		if flagSet.NArg() == 0 {
			flagSet.Usage()
			os.Exit(1)
		}
		sd.AddReviewersToPr(flagSet.Args(), whenChecksPass, silent, minChecks, reviewers, pollFrequency)
	}}
}
