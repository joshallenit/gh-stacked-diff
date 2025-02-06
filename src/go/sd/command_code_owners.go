package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	sd "stackeddiff"
)

func createCodeOwnersCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("code-owners", flag.ContinueOnError)

	return Command{
		FlagSet: flagSet,
		Summary: "Outputs code owners for all of the changes in branch",
		Description: "Outputs code owners for each file that has been modified\n" +
			"in the current local branch when compared to the remote main branch",
		Usage:           "sd " + flagSet.Name(),
		DefaultLogLevel: slog.LevelError,
		OnSelected: func(command Command) {
			if flagSet.NArg() != 0 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			fmt.Fprint(stdOut, sd.ChangedFilesOwnersString())
		}}
}
