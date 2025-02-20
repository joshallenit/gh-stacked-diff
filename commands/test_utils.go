package commands

import (
	"flag"
	"io"
	"log/slog"

	"stackeddiff/testinginit"
)

func testParseArguments(commandLineArgs ...string) string {
	panicOnExit := func(stdErr io.Writer, errorCode int, logLevelVar *slog.LevelVar, err any) {
		panic(err)
	}
	out := testinginit.NewWriteRecorder()
	parseArguments(out, out, flag.NewFlagSet("sd", flag.ContinueOnError), commandLineArgs, panicOnExit)
	return out.String()
}
