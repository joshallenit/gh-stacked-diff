package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	sd "stackeddiff"
)

func CreateWaitForMergeCommand() Command {
	flagSet := flag.NewFlagSet("wait-for-merge", flag.ExitOnError)
	var silent bool
	flagSet.BoolVar(&silent, "silent", false, "Whether to use voice output")

	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Waits for a pull request to be merged. Polls PR every 5 minutes. Useful for custom scripting.")
		fmt.Fprintln(os.Stderr, "sd wait-for-merge [flags] <commit hash or pull request number>")
		flagSet.PrintDefaults()
	}

	return Command{FlagSet: flagSet, DefaultLogLevel: slog.LevelError, OnSelected: func() {
		if flagSet.NArg() != 1 {
			flagSet.Usage()
			os.Exit(1)
		}
		sd.WaitForMerge(flagSet.Arg(0), silent)
	}}
}
