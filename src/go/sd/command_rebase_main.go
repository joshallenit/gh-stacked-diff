package main

import (
	"flag"
	"log"
	sd "stackeddiff"
	ex "stackeddiff/execute"
)

func CreateRebaseMainCommand() Command {
	flagSet := flag.NewFlagSet("rebase-main", flag.ContinueOnError)

	return Command{
		FlagSet: flagSet,
		Summary: "Bring your main branch up to date with remote",
		Description: "Rebase with origin/" + ex.GetMainBranch() + ", dropping any commits who's branches have been merged.\n" +
			"This avoids having to manually call \"git reset --hard head\" whenever you have merge conflicts with a commit that\n" +
			"has already been merged but has slight variation with local main because, for example, a change was \n" +
			"made with the Github Web UI",
		Usage: "sd " + flagSet.Name(),
		OnSelected: func(command Command) {
			if flagSet.NArg() != 0 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			sd.RebaseMain(log.Default())
		}}
}
