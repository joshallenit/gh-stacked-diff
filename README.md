# Developer Scripts

These scripts make it easier to build from the command line and to create and update PR's. They facilitates a [stacked diff workflow](https://kastiglione.github.io/git/2020/09/11/git-stacked-commits.html), where you always commit on `main` branch and have can have multiple streams of work all on `main`.

### TL;DR

Using a stacked diff workflow like this allows you to work on separate streams of work without changing branches.

## Build Scripts

- installapp: assembleInternalDebug and install on device
- assemble: assembleInternalDebug and build tests


## Working with Pull Requests

- **gitlog**: Abbreviate git log that only shows what has changed, useful for copying commit hashes.
- **git-newpr**: Create a new PR with a cherry-pick of the given commit hash
- **git-updatepr**: Add the topmost commit to the PR with the given commit hash
- **git-reset-main**: Reset the main branch with the squashed contents of the given commits associated branch. Sometimes you might want to switch to a feature branch and make changes to it (rebase, amend). With this script you can then ensure that your `main` branch is up to date.
- **git-prs**: Lists all of your open PRs. Useful for copying PR numbers.
- **git-review**: Update the given PR as "Ready for Review" and automatically add reviewers listed in your PR_REVIEWERS environment variable.

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
echo "export PR_REVIEWERS=first-user,second-user,third-user" >> ~/.zshrc # List regular PR reviewers for the git-review command
```

## When Creating PRs

If you prefix the Jira ticket to the git commit summary then `git-newpr` will populate the `Ticket` section of the PR description.

For example:
`CONV-9999 Add new feature`

## Automatically adding PR Reviewers

The `git-review` command will mark your Draft PR as "Ready for Review" and automatically add reviewers that are specified in the PR_REVIEWERS environment variable.
You can specify more than one reviewer using a comma-delimited string.

```bash
export PR_REVIEWERS=first-user,second-user,third-user
```

Add this to your shell rc file (`~/.zshrc` or `~/.bashrc`) and run `source <rc-file>`

## Example Workflow

### To Update Main

```bash
git fetch && git rebase origin/main
```
