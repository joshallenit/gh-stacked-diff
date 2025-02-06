# Change Log

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-02-28

### Added

- Ability to use log list index for a commit indicator. Avoids having to copy & paste git hashes or PR numbers.
- Ability to add reviewers from `new` and `update` commands. 
- `sd log` now also displays commits on associated branches.
- Ability to set log level via `sd` flag "--log-level".
- Unit tests

### Changed

- `sd rebase-main` now outputs in real time instead of only when rebase ends.
- Moved all scripts to subcommands of a new `sd` executable.
- Converted all logs to use slog (logs at DEBUG, INFO, or ERROR levels) so that the log level can be changed to help with debugging. 
- Renamed replace-head to replace-conflicts
- `sd log` was made faster by running some git commands once, instead of for each commit.
- `replace-conflicts` now asks for confirmation (also has a confirm flag).

### Fixed

- Can now use `new` for root commit.
- More reliable getMainBranch()
- More reliable help and command line parsing error messages
