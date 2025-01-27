package main

import (
	"flag"
	"fmt"
	"io"
	sd "stackeddiff"
)

func CreateCodeOwnersCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("code-owners", flag.ExitOnError)

	return Command{FlagSet: flagSet, OnSelected: func() {
		fmt.Fprint(stdOut, sd.ChangedFilesOwnersString())
	}}
}
