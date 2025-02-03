# How To Build Scripts

The scripts are written in golang and bash.

1. Install [golang](https://go.dev/dl/).
2. Install make. This is already installed on Mac, but to instructions for windows are [here](https://leangaurav.medium.com/how-to-setup-install-gnu-make-on-windows-324480f1da69).

Then run:

```bash
make build
```

# Code Organization

The main entry point to the CLI is [src/go/sd/sd_main.go]. 
