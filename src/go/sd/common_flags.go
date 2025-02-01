package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	sd "stackeddiff"
	ex "stackeddiff/execute"
	"strings"
)

func AddIndicatorFlag(flagSet *flag.FlagSet) *string {
	var usage string = "Indicator type to use to interpret commitIndicator:\n" +
		"   commit   a commit hash, can be abbreviated,\n" +
		"   pr       a github Pull Request number,\n" +
		"   list     the order of commit listed in the git log, as indicated by \"sd log\"\n" +
		"   guess    the command will guess the indicator type:\n" +
		"      Number between 0 and 99:       list\n" +
		"      Number between 100 and 999999: pr\n" +
		"      Otherwise:                     commit\n"
	return flagSet.String("indicator", string(sd.IndicatorTypeGuess), usage)
}

func CheckIndicatorFlag(flagSet *flag.FlagSet, indicatorTypeString *string) sd.IndicatorType {
	indicatorType := sd.IndicatorType(*indicatorTypeString)
	if !indicatorType.IsValid() {
		fmt.Fprintln(flagSet.Output(), "Invalid indicator type: "+*indicatorTypeString)
		flagSet.Usage()
		os.Exit(1)
	}
	return indicatorType
}

func AddReviewersFlags(flagSet *flag.FlagSet, reviewersUseCaseExtra string) (*string, *bool, *int) {
	reviewers := flagSet.String("reviewers", "", "Comma-separated list of Github usernames to add as reviewers once checks have passed.\n"+
		reviewersUseCaseExtra)
	silent := AddSilentFlag(flagSet, "reviewers have been added")
	minChecks := flagSet.Int("min-checks", 4,
		"Minimum number of checks to wait for before verifying that checks have passed before adding reviewers.\n"+
			"It takes some time for checks to be added to a PR by Github,\n"+
			"and if you add-reviewers too soon it will think that they have all passed.")
	return reviewers, silent, minChecks
}

func AddSilentFlag(flagSet *flag.FlagSet, usageUseCase string) *bool {
	if runtime.GOOS == "darwin" {
		// Only supported on Mac OS X because it uses "say" command.
		return flagSet.Bool("silent", false, "Whether to use voice output (false) or be silent (true) to notify that "+usageUseCase+".")
	} else {
		silent := new(bool)
		*silent = true
		return silent
	}
}

func commandHelp(flagSet *flag.FlagSet, description string, usage string, isError bool) {
	fmt.Fprintln(flagSet.Output(), ex.Reset+description)
	printUsage(flagSet, usage)
	if isError {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func commandError(flagSet *flag.FlagSet, errMessage string, usage string) {
	if !strings.HasPrefix(errMessage, "flag provided but not defined:") {
		fmt.Fprintln(flagSet.Output(), ex.Reset+"error: "+errMessage)
	}
	printUsage(flagSet, usage)
	os.Exit(1)
}

func printUsage(flagSet *flag.FlagSet, usage string) {
	fmt.Fprintln(flagSet.Output(), "")
	fmt.Fprintln(flagSet.Output(), "usage: "+usage)
	hasFlags := false
	// There's no other way to get the number of possible flags, so use VisitAll.
	flagSet.VisitAll(func(flag *flag.Flag) {
		hasFlags = true
	})
	if hasFlags {
		fmt.Fprintln(flagSet.Output(), "")
		fmt.Fprintln(flagSet.Output(), "flags:")
		fmt.Fprintln(flagSet.Output(), "")
		flagSet.PrintDefaults()
	}
}
