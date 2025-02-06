package execute

import (
	"fmt"
	"log/slog"
	"slices"
)

// Response that was executed or will be faked.
type ExecuteResponse struct {
	Out         string // Out to return.
	Err         error  // error to return.
	ProgramName string // Program name to match or that was used.
	// Arguments to match or that were used.
	// Set last value to MatchAnyRemainingArgs to match any remaining arguments.
	Args []string
}

// Fake [Executor] for testing.
type TestExecutor struct {
	fakeResponses []ExecuteResponse
	Responses     []ExecuteResponse
}

// Can be used use as last value of [TestExecutor.fakeResponses] [ExecuteResponse.Args]
const MatchAnyRemainingArgs = "MatchCommandWithAnyRemainingArgs"

// Ensure that [TestExecutor] implements [Executor].
var _ Executor = &TestExecutor{}

// Checks [TestExecutor.fakeResponses] for any match before calling [DefaultExecutor.Execute].
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
				fakeResponse := ExecuteResponse{
					Out:         response.Out,
					Err:         response.Err,
					ProgramName: programName,
					Args:        args}
				t.Responses = append(t.Responses, fakeResponse)
				slog.Debug(fmt.Sprint("Matched ", fakeResponse))
				return response.Out, response.Err
			}
		}
	}
	out, err := (&DefaultExecutor{}).Execute(options, programName, args...)
	t.Responses = append(t.Responses, ExecuteResponse{Out: out, Err: err, ProgramName: programName, Args: args})
	return out, err
}

// Adds a response to [TestExecutor.fakeResponses].
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
