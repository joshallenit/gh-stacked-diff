package main

import (
	"flag"
	"fmt"
	"os"
	sd "stackeddiff"
)

func AddIndicatorFlag(flagSet *flag.FlagSet) *string {
	var usage string = "Indicator type being used for which git commit is being selected:\n" +
		"- " + string(sd.IndicatorTypeCommit) + ": a commit hash, can be abbreviated,\n" +
		"- " + string(sd.IndicatorTypePr) + ": a github Pull Request number,\n" +
		"- " + string(sd.IndicatorTypeList) + ": the order of commit listed in the git log, as indicated by \"sd log\"\n" +
		"- " + string(sd.IndicatorTypeGuess) + ": the command will guess the indicator type:\n" +
		"   - Number between 0 and 99: list\n" +
		"   - Number between 100 and 999999: pr number\n" +
		"   - Otherwise: commit\n"
	return flagSet.String("indicator", string(sd.IndicatorTypeGuess), usage)
}

func CheckIndicatorFlag(flagSet *flag.FlagSet, indicatorTypeString *string) sd.IndicatorType {
	indicatorType := sd.IndicatorType(*indicatorTypeString)
	if !indicatorType.IsValid() {
		fmt.Fprintln(os.Stderr, "Invalid indicator type: "+*indicatorTypeString)
		flagSet.Usage()
		os.Exit(1)
	}
	return indicatorType
}

// add reviewers to update as well
func AddReviewersFlag(flagSet *flag.FlagSet) (*string, *bool, *int) {
	reviewers := flagSet.String("reviewers", "", "Comma-separated list of Github usernames to add as reviewers once checks have passed.")
	silent := flagSet.Bool("silent", false, "Whether to use voice output (false) or be silent (true) to notify that reviewers have been added.")
	minChecks := flagSet.Int("min-checks", 4,
		"Minimum number of checks to wait for before verifying that checks have passed before adding reviewers. "+
			"It takes some time for checks to be added to a PR by Github, "+
			"and if you add-reviewers too soon it will think that they have all passed.")
	return reviewers, silent, minChecks
}

/*
move silent flag here
*/

/*
then check out the help
*/
