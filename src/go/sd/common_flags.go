package main

import (
	"flag"
	sd "stackeddiff"
)

func AddIndicatorFlag(flagSet *flag.FlagSet) *string {
	var usage string = "Indicator type being used for which git commit is being selected. " +
		"Default is " + string(sd.IndicatorTypeGuess) + "\n" +
		string(sd.IndicatorTypeCommit) + " - a commit hash, can be abbreviated,\n" +
		string(sd.IndicatorTypePr) + " - a github Pull Request number,\n" +
		string(sd.IndicatorTypeList) + " - the order of commit listed in the git log, as indicated by `sd log`\n" +
		string(sd.IndicatorTypeGuess) + " - the command will guess the indicator type based on length and whether it is all numeric:\n" +
		"   - Number between 0 and 99: list\n" +
		"   - Number between 100 and 999999: pr number\n" +
		"   - Otherwise: commit\n"
	return flagSet.String("indicator", string(sd.IndicatorTypeGuess), usage)
}
