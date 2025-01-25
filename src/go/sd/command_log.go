package main

import (
	"flag"
	"io"
	sd "stackeddiff"
)

func CreateLogCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("log", flag.ExitOnError)

	return Command{FlagSet: flagSet, OnSelected: func() {
		sd.PrintGitLog(stdOut)
	}}
}
