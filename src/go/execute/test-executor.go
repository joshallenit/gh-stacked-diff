package execute

import (
	"log/slog"
	"slices"
)

type ExecuteResponse struct {
	Out         string
	Err         error
	ProgramName string
	Args        []string
}

type TestExecutor struct {
	TestLogger    *slog.Logger
	fakeResponses []ExecuteResponse
	Responses     []ExecuteResponse
}

const MatchAnyRemainingArgs = "MatchCommandWithAnyRemainingArgs"

// Ensure that [TestExecutor] implements [Executor].
var _ Executor = &TestExecutor{}

func (t *TestExecutor) Execute(options ExecuteOptions, programName string, args ...string) (string, error) {
	for _, response := range t.fakeResponses {
		if response.ProgramName == programName {
			matchArgs := args
			matchResponseArgs := response.Args
			if len(response.Args) > 0 && response.Args[len(response.Args)-1] == MatchAnyRemainingArgs && len(args) >= len(response.Args)-1 {
				matchArgs = args[0 : len(response.Args)-1]
				matchResponseArgs = response.Args[0 : len(response.Args)-1]
			}
			matched := true
			if len(matchArgs) == len(matchResponseArgs) {
				for i := range matchArgs {
					if matchArgs[i] != matchResponseArgs[i] {
						matched = false
						break
					}
				}
			} else {
				matched = false
			}
			if matched {
				t.Responses = append(t.Responses, ExecuteResponse{
					Out:         response.Out,
					Err:         response.Err,
					ProgramName: programName,
					Args:        args},
				)
				return response.Out, response.Err
			}
		}
	}
	out, err := (&DefaultExecutor{}).Execute(options, programName, args...)
	t.Responses = append(t.Responses, ExecuteResponse{Out: out, Err: err, ProgramName: programName, Args: args})
	return out, err
}

func (t TestExecutor) Logger() *slog.Logger {
	return t.TestLogger
}

func (t *TestExecutor) SetResponse(out string, err error, programName string, args ...string) {
	t.fakeResponses = append(t.fakeResponses, ExecuteResponse{Out: out, Err: err, ProgramName: programName, Args: args})
	slices.SortFunc(t.fakeResponses, func(a ExecuteResponse, b ExecuteResponse) int {
		// Check the response that has the longer argument list first

		// cmp(a, b) should return a negative number when a < b, a positive number when a > b
		// so longer parameter list should be negative.
		argDiff := len(b.Args) - len(a.Args)
		if argDiff != 0 {
			return argDiff
		}
		// Check the response that does not have a wildcard first
		if len(a.Args) > 0 && a.Args[len(a.Args)-1] == MatchAnyRemainingArgs &&
			len(b.Args) > 0 && b.Args[len(b.Args)-1] != MatchAnyRemainingArgs {
			return 1
		}
		if len(b.Args) > 0 && b.Args[len(b.Args)-1] == MatchAnyRemainingArgs &&
			len(a.Args) > 0 && a.Args[len(a.Args)-1] != MatchAnyRemainingArgs {
			return -1
		}
		// Otherwise it doesn't matter which is checked first
		return 0
	})
}
