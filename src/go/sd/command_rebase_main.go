package main

import (
	"flag"
	"log"
	sd "stackeddiff"
)

func CreateRebaseMainCommand() Command {
	flagSet := flag.NewFlagSet("rebase-main", flag.ExitOnError)

	return Command{FlagSet: flagSet, OnSelected: func() {
		sd.RebaseMain(log.Default())
	}}
}
