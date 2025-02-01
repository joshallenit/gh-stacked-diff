package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	sd "stackeddiff"
)

func CreateCodeOwnersCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("code-owners", flag.ExitOnError)

	return Command{
		FlagSet:         flagSet,
		UsageSummary:    "Outputs code owners for all of the changes in branch",
		DefaultLogLevel: slog.LevelError,
		OnSelected: func() {
			fmt.Fprint(stdOut, sd.ChangedFilesOwnersString())
		}}
}
