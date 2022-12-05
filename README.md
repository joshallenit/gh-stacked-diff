# Developer Scripts

These scripts make it easier to build from the command line and to create and update PR's. They facilitates a [stacked diff workflow](https://kastiglione.github.io/git/2020/09/11/git-stacked-commits.html), where you always commit on `main` branch and have can have multiple streams of work all on `main`.

Join the discussion in [#devel-stacked-diff-workflow](https://slack-pde.slack.com/archives/C03V94N2A84)

### TL;DR

Using a stacked diff workflow like this allows you to work on separate streams of work without changing branches.

## Build Scripts

- installapp: assembleInternalDebug and install on device
- assemble: assembleInternalDebug and build tests


## Working with Pull Requests

- **gitlog**: Abbreviated git log that only shows what has changed, useful for copying commit hashes.
- **git-newpr**: Create a new PR with a cherry-pick of the given commit hash
- **git-updatepr**: Add the topmost commit to the PR with the given commit hash
- **git-reset-main**: Reset the main branch with the squashed contents of the given commits associated branch. Sometimes you might want to switch to a feature branch and make changes to it (rebase, amend). With this script you can then ensure that your `main` branch is up to date.
- **git-prs**: Lists all of your open PRs. Useful for copying PR numbers.
- **git-review**: Update the given PR as "Ready for Review" and automatically add reviewers listed in your PR_REVIEWERS environment variable.
- **git-merge-pr**: Add the given PR to the merge queue
- **git-checkout**: Checkout the feature branch associated with a given PR or commit. For when you want to checkout the feature branch to add commits or rebase, (instead of `git-update-pr`), and then use `git-reset-main` to sync `main`

## Working with these scripts

- create-symlinks.sh - Create symlinks for these scripts

## Installation

Clone repository and then:

```bash
# Install Github CLI
brew install gh 
# Setup login for Github CLI
gh auth login 
# Go Lang is used by git-reset-main
brew install go 
# Install pcregrep
brew install pcre 
# cd where you clone the repository
cd development-scripts 
# Create symlinks for the scripts
./create-symlinks.sh 
# Setup git-review with the your regular reviewers
echo "export PR_REVIEWERS=first-user,second-user,third-user" >> ~/.zshrc
```

To avoid having some commands show their output in full-screen `less` (or your default pager), define these args in your LESS environment variable:

```bash
export LESS=-FRX
```

Add to your `.zshrc` file:

```bash
echo "export LESS=-FRX" >> `~/.zshrc`
```

Don't forget to run `source`!

```bash
source ~/.zshrc
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
