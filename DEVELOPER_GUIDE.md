# Developer Guide

## How To Build

The Stacked Diff Workflow CLI (also known as, the `sd` command) is written in golang. 

1. Install [golang](https://go.dev/dl/).
2. Install make. This is already installed on Mac, but instructions for windows are [here](https://leangaurav.medium.com/how-to-setup-install-gnu-make-on-windows-324480f1da69).

Then run:

```bash
make build
```

Binaries are created in `./bin`.

### Installation From Source

Clone the repository, [then build](DEVELOPER_GUIDE.md#how-to-build), and then follow the [install instructions](#installation-from-a-release) for your platform.


## Code Organization

The main entry point to the Stacked Diff Workflow CLI ("sd") is [main.go]. The commands are implemented under [commands].

## How to Make a Release

1. Set releaseVersion in [project.properties](project.properties).
2. On each platform (Windows and Mac) run `make release`, setting the PLATFORM environment variable accordingly.
```bash
# On a Windows machine
export PLATFORM=windows; make release
# On a Mac machine
export PLATFORM=mac; make release
```

## How to Debug Unit Tests

If one of the command*_test fails you can pass "--log-level=debug" to `parseArguments` for more detailed logging. For more detailed logging up until the `parseArguments` call use `testutil.InitTest(t, slog.LevelDebug)`


## Making a Release

Follow the steps in golang docs [Publishing a module](https://go.dev/doc/modules/publishing):

```bash
# Update latest and stable version numbers in [project.properties], merge changes, update local, and then:
go mod tidy
make test
# Make sure all changes merged into main, git status and sd log should be empty.
# Otherwise save your changes, "git reset --hard origin/main", create tag, then restore your changes
git status && sd log
export RELEASE_VERSION=`grep "latestVersion" "project.properties" | cut -d '=' -f2`;\
git tag v$RELEASE_VERSION
git push origin v$RELEASE_VERSION
GOPROXY=proxy.golang.org go list -m github.com/joshallenit/gh-stacked-diff/v2@v$RELEASE_VERSION
```

For bubbles and bubbletea forks:

```bash
GOPROXY=proxy.golang.org go list -m github.com/joshallenit/bubbles@v0.20.1

GOPROXY=proxy.golang.org go list -m github.com/joshallenit/bubbletea@v1.3.5
```


Once a tag is created [.github/workflows/release.yml] kicks off and creates the binaries for the release.

## Usage as a golang Library

Look at [main.go] for example usage.
