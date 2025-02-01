package main

import (
	"flag"
	"log"
	sd "stackeddiff"
)

func CreateRebaseMainCommand() Command {
	flagSet := flag.NewFlagSet("rebase-main", flag.ExitOnError)

	return Command{
		FlagSet:      flagSet,
		UsageSummary: "Bring your main branch up to date with remote",
		OnSelected: func() {
			sd.RebaseMain(log.Default())
		}}
}
