package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	sd "stackeddiff"
	ex "stackeddiff/execute"
)

func CreateUpdateCommand() Command {
	flagSet := flag.NewFlagSet("update", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprint(os.Stderr,
			ex.Reset+"Add one or more commits to a PR.\n"+
				"\n"+
				"update-pr <pr-commit> [fixup commit (defaults to top commit)] [other fixup commit...]\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flagSet.PrintDefaults()
	}

	return Command{FlagSet: flagSet, OnSelected: func() {
		if flagSet.NArg() == 0 {
			flagSet.Usage()
			os.Exit(1)
		}

		var otherCommits []string
		if len(flagSet.Args()) > 1 {
			otherCommits = flagSet.Args()[1:]
		}
		sd.UpdatePr(flagSet.Arg(0), otherCommits, log.Default())
	}}
}
