package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	sd "stackeddiff"
)

func CreateMainBranchCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("branch-name", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Outputs name of the main branch: main or master")
		flagSet.PrintDefaults()
	}

	return Command{FlagSet: flagSet, DefaultLogLevel: slog.LevelError, OnSelected: func() {
		if flagSet.NArg() != 1 {
			flagSet.Usage()
			os.Exit(1)
		}
		branchName := sd.GetBranchInfo(flagSet.Arg(0)).BranchName
		fmt.Fprint(stdOut, branchName)
	}}
}
