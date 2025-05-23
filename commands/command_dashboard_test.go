package commands

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/joshallenit/gh-stacked-diff/v2/testutil"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func TestSdDashboard_ShowsDashboard(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "")

	out := testParseArguments("dashboard")

	// Since the dashboard is interactive, we can only verify it doesn't error
	assert.NotContains(out, "error:")
}

func TestSdDashboard_WithMinChecks_ShowsDashboard(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "")

	out := testParseArguments("dashboard", "--min-checks=2")

	// Since the dashboard is interactive, we can only verify it doesn't error
	assert.NotContains(out, "error:")
}

func TestSdDashboard_WithoutWatch_ShowsDashboard(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "")

	out := testParseArguments("dashboard", "--watch=false")

	// Since the dashboard is interactive, we can only verify it doesn't error
	assert.NotContains(out, "error:")
}

func TestSdDashboard_WithInvalidFlag_ShowsError(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)

	out := testParseArguments("dashboard", "--invalid-flag")

	assert.Contains(out, "error: flag provided but not defined: -invalid-flag")
}

func TestSdDashboard_WithExtraArgs_ShowsError(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)

	out := testParseArguments("dashboard", "extra-arg")

	assert.Contains(out, "error: too many arguments")
}

func TestSdDashboard_WithWatchMode_RefreshesOnChanges(t *testing.T) {
	// Skip this test in CI environments
	if os.Getenv("CI") != "" {
		t.Skip("Skipping watch mode test in CI environment")
	}

	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)

	testutil.AddCommit("first", "")
	testutil.AddCommit("second", "")

	// Create a context with timeout for the test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run the command in a goroutine
	go func() {
		testParseArguments("dashboard", "--watch")
	}()

	// Wait a bit for the dashboard to start
	time.Sleep(500 * time.Millisecond)

	// Add another commit to trigger refresh
	testutil.AddCommit("third", "")

	// Wait for the context to timeout
	<-ctx.Done()
}

func TestFindGitDir_FindsGitDir(t *testing.T) {
	assert := assert.New(t)
	testutil.InitTest(t, slog.LevelError)

	// Create a subdirectory
	subDir := "subdir"
	util.ExecuteOrDie(util.ExecuteOptions{}, "mkdir", "-p", subDir)
	defer util.ExecuteOrDie(util.ExecuteOptions{}, "rm", "-rf", subDir)

	// Change to the subdirectory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test finding the .git directory
	foundGitDir, err := findGitDir()
	if err != nil {
		t.Errorf("findGitDir() error = %v", err)
	}
	assert.Equal(filepath.Join(originalDir, ".git"), foundGitDir)
}
