# Developer Scripts

These scripts make it easier to build from the command line and to create and update PR's. They facilitates a [stacked diff workflow](https://kastiglione.github.io/git/2020/09/11/git-stacked-commits.html), where you always commit on `main` branch and have can have multiple streams of work on all `main`.

## Build Scripts

- installapp: assembleInternalDebug and install on device
- assemble: assembleInternalDebug and build tests


## Working with Pull Requests

- gitlog: Abbreviate git log that only shows what has changed, useful for copying commit hashes.
- git-newpr: Create a new PR with a cherry-pick of the given commit hash
- git-updatepr: Add the topmost commit to the PR with the given commit hash
- git-reset-main: Reset the main branch with the squashed contents of the given commits associated branch.

## Working with these scripts

- create-symlinks.sh - Create symlinks for these scripts

## Installation

Clone repository and then:

```bash
brew install gh # Install Github CLI
gh auth login # Setup login for Github CLI
brew install go # Go Lang is used by git-reset-main
cd development-scripts # or wherever you cloned repo
./create-symlinks.sh # Create symlinks for the scripts
```
