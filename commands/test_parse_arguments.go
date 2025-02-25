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
	// bin directory must be on PATH for tests to pass so that sequenceEditorPrefix will execute.
	parseArguments(out, out, flag.NewFlagSet("sd", flag.ContinueOnError), commandLineArgs, "sd --log-level=INFO ", panicOnExit)
	return out.String()
}
