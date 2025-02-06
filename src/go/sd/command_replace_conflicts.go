package main

import (
	"flag"
	"io"
	sd "stackeddiff"
)

func CreateReplaceConflictsCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("replace-conflicts", flag.ContinueOnError)
	confirmed := flagSet.Bool("confirm", false, "Whether to automatically confirm to do this rather than ask for y/n input")
	return Command{
		FlagSet: flagSet,
		Summary: "For failed rebase: replace changes with its associated branch",
		Description: "During a rebase that failed because of merge conflicts, replace the\n" +
			"current uncommitted changes (merge conflicts), with the contents\n" +
			"(diff between origin/" + sd.GetMainBranchForHelp() + " and HEAD) of its associated branch.",
		Usage: "sd " + flagSet.Name(),
		OnSelected: func(command Command) {
			if flagSet.NArg() > 0 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			sd.ReplaceConflicts(stdOut, *confirmed)
		}}
}
