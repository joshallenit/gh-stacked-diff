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

## How to Debug Unit Tests

If one of the command*_test fails you can pass "--log-level=debug" to `parseArguments` for more detailed logging. For more detailed logging up until the `parseArguments` call use `testutil.InitTest(t, slog.LevelDebug)`

## Making a Release

Follow the steps in golang docs [Publishing a module](https://go.dev/doc/modules/publishing):

```bash
# Update the stable version so that it is equal to current version [util/stable_version.txt], merge changes, update local, and then:
git checkout main
go mod tidy
make test
# Make sure all changes merged into main, git status and sd log should be empty.
# Otherwise save your changes, "git reset --hard origin/main", create tag, then restore your changes
git status && sd log
export RELEASE_VERSION=`cat util/stable_version.txt`
git tag v$RELEASE_VERSION
git push origin v$RELEASE_VERSION
GOPROXY=proxy.golang.org go list -m github.com/joshallenit/gh-stacked-diff/v2@v$RELEASE_VERSION
# Update [util/current_version.txt]
```

For bubbles and bubbletea forks:

```bash
# Make labels in their repositories
git checkout main
git status && sd log
export RELEASE_VERSION=1.3.6
git tag v$RELEASE_VERSION
git push origin v$RELEASE_VERSION
# sd rebase-main if required
GOPROXY=proxy.golang.org go list -m github.com/joshallenit/bubbletea@v$RELEASE_VERSION
# Update version in go.mod replace
go mod tidy

# Same steps for:
GOPROXY=proxy.golang.org go list -m github.com/joshallenit/bubbles@v0.20.1
```


Once a tag is created [.github/workflows/release.yml] kicks off and creates the binaries for the release.

## Usage as a golang Library

Look at [main.go] for example usage.
