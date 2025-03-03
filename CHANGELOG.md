# Change Log

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.8](https://github.com/joshallenit/gh-stacked-diff/compare/v2.0.0...v2.0.8) - 2025-03-03

### Added

- Build actions so that this can be used as a Github CLI plugin.

### Changes

- Reorganized code so that project can be used as a Github CLI plugin and go library.
- Removed scripts under /src/bash that were deprecated.
- Renamed repository from stacked-diff-workflow to gh-stacked-diff. The gh- prefix is required for it to work as a Github CLI plugin.
- More reliable rollback for "new" and "update" command when there is a problem.

## [2.0.0](https://github.com/joshallenit/gh-stacked-diff/compare/v1.3.0...v2.0.0) - 2025-02-16

### Added

- Ability to use log list index for a commit indicator. Avoids having to copy & paste git hashes or PR numbers.
- Ability to add reviewers from `new` and `update` commands. 
- `sd log` now also displays commits on associated branches.
- Ability to set log level via `sd` flag "--log-level".
- Unit tests.

### Changed

- `sd rebase-main` now outputs in real time instead of only when rebase ends.
- Moved all scripts to subcommands of a new `sd` executable.
- Converted all logs to use slog (logs at DEBUG, INFO, or ERROR levels) so that the log level can be changed to help with debugging. 
- Renamed replace-head to replace-conflicts
- `sd log` was made faster by running some git commands once, instead of for each commit.
- `sd replace-conflicts` now asks for confirmation (also has a confirm flag).
- `sd rebase-main` now deletes branches that have been merged.

### Fixed

- More reliable `getMainBranch`.
- More reliable help and command line parsing error messages.
- More reliable `sd rebase-main` by using Github CLI to check for merged branches.
