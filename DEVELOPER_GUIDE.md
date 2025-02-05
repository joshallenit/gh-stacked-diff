# Contributing Developer Guide

## How To Build Scripts

The scripts are written in golang and bash.

1. Install [golang](https://go.dev/dl/).
2. Install make. This is already installed on Mac, but to instructions for windows are [here](https://leangaurav.medium.com/how-to-setup-install-gnu-make-on-windows-324480f1da69).

Then run:

```bash
make build
```

Binaries are created in `./bin`.

## Code Organization

The main entry point to the Stacked Diff Workflow CLI ("sd") is [src/go/sd/sd_main.go].

## How to Make a Release

1. Set releaseVersion in [project.properties](project.properties).
2. On each platform (Windows and Mac) run `make release`, setting the PLATFORM environment variable accordingly.
```bash
# On a Windows machine
export PLATFORM=windows; make release
# On a Mac machine
export PLATFORM=mac; make release
```
