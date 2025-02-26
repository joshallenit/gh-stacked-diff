package commands

import (
	"flag"
	"io"
	"log/slog"

	"github.com/joshallenit/stacked-diff/v2/testutil"
)

// Calls [parseArguments] for unit tests.
func testParseArguments(commandLineArgs ...string) string {
	panicOnExit := func(stdErr io.Writer, errorCode int, logLevelVar *slog.LevelVar, err any) {
		panic(err)
	}
	out := testutil.NewWriteRecorder()
	// Executable must be on PATH for tests to pass so that sequenceEditorPrefix will execute.
	// PATH is set in ../Makefile
	parseArguments(out, out, flag.NewFlagSet("sd", flag.ContinueOnError), commandLineArgs, "stacked-diff --log-level=INFO ", panicOnExit)
	return out.String()
}
