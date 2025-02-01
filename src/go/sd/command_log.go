package main

import (
	"flag"
	"io"
	"log/slog"
	sd "stackeddiff"
)

func CreateLogCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("log", flag.ExitOnError)

	return Command{
		FlagSet:         flagSet,
		UsageSummary:    "Displays git log of your changes",
		DefaultLogLevel: slog.LevelError,
		OnSelected: func() {
			sd.PrintGitLog(stdOut)
		}}
}
