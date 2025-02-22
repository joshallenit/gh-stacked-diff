package commands

import (
	"flag"
	"io"
	"log/slog"

	"github.com/joshallenit/stacked-diff/v2/testutil"
)

func testParseArguments(commandLineArgs ...string) string {
	panicOnExit := func(stdErr io.Writer, errorCode int, logLevelVar *slog.LevelVar, err any) {
		panic(err)
	}
	out := testutil.NewWriteRecorder()
	parseArguments(out, out, flag.NewFlagSet("sd", flag.ContinueOnError), commandLineArgs, panicOnExit)
	return out.String()
}
