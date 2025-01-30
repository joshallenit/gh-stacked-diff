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
	indicatorTypeString := AddIndicatorFlag(flagSet)
	flagSet.Usage = func() {
		fmt.Fprint(os.Stderr,
			ex.Reset+"Add one or more commits to a PR.\n"+
				"\n"+
				"sd update [flags] <pr-commit> [fixup commit (defaults to top commit)] [other fixup commit...]\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flagSet.PrintDefaults()
	}

	return Command{FlagSet: flagSet, OnSelected: func() {
		if flagSet.NArg() == 0 {
			flagSet.Usage()
			os.Exit(1)
		}
		indicatorType := CheckIndicatorFlag(flagSet, indicatorTypeString)
		var otherCommits []string
		if len(flagSet.Args()) > 1 {
			otherCommits = flagSet.Args()[1:]
		}
		sd.UpdatePr(flagSet.Arg(0), otherCommits, indicatorType, log.Default())
	}}
}
