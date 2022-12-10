# Developer Scripts for Stacked Diff Workflow

These scripts make it easier to build from the command line and to create and update PR's. They facilitates a [stacked diff workflow](https://kastiglione.github.io/git/2020/09/11/git-stacked-commits.html), where you always commit on `main` branch and have can have multiple streams of work all on `main`.

Join the discussion in [#devel-stacked-diff-workflow](https://slack-pde.slack.com/archives/C03V94N2A84)

## TL;DR

Using a stacked diff workflow like this allows you to work on separate streams of work without changing branches.

## Scripts

### For Stacked Diff Workflow

#### git-checkout

`git-checkout <commit hash or pr number>`

Checkout the feature branch associated with a given PR or commit. For when you want to checkout the feature branch to rebase with origin/main, merge with origin/main, or for any other reason. After modifying the feature branch use `replac-commit` or `replace-head` to sync local `main`.

#### git-updatepr

`git-updatepr <commit hash or pr number>`

Add the topmost commit to the PR with the given commit hash

#### gitlog

`gitlog`

Abbreviated git log that only shows what has changed, useful for copying commit hashes.

#### new-pr

Create a new PR with a cherry-pick of the given commit hash

```
new-pr <commitHash>
Usage of new-pr:
  -draft
    	Whether to create the PR as draft (default true)
```

##### Ticket Number

If you prefix the Jira ticket to the git commit summary then `newpr` will populate the `Ticket` section of the PR description.

For example:
`CONV-9999 Add new feature`

##### Templates

The Pull Request Title, Body (aka Description), and Branch Name are created from [golang templates](https://pkg.go.dev/text/template). The defaults are:

- [branch-name.template](cmd/config/branch-name.template)
- [pr-description.template](cmd/config/pr-description.template)
- [pr-title.template](cmd/config/pr-title.template)

The [possible values](config/templates) for the templates are:

- **TicketNumber** - Jira ticket as parsed from the commit summary
- **Username** -  Name as parsed from git config email
- **CommitBody** - Body of the commit message
- **CommitSummary** - Summary line of the commit message
- **CommitSummaryCleaned** - Summary line of the commit message without spaces or special characters
- **CommitSummaryWithoutTicket** - Summary line of the commit message without the prefix of the ticket number

To change a template, copy the default from [cmd/config/](cmd/config/) into `~/.stacked-diff-workflow/` and modify.

#### replace-commit

`replace-commit <commit hash or pr number>`

Reset the main branch with the squashed contents of the given commits associated branch. Sometimes you might want to switch to a feature branch and make changes to it (rebase, amend). With this script you can then ensure that your `main` branch is up to date.

#### replace-head

`replace-head`

Use during rebase of main branch to use the contents of a feature branch that already fixed the merge conflicts.

### To Help You Build

#### assemble-app

`assemble-app`

Calls `./gradlew assembleInternalDebug` and build tests. Use "-s" (silent) flag to not use voice (`say`) to announce success/failure.

#### install-app

`install-app`

Calls `./gradlew assembleInternalDebug` and install on real device. Use "-s" (silent) flag to not use voice (`say`) to announce success/failure.

### To Help with Github

#### add-reviewers

`add-reviewers <pull request number>`

The `add-reviewers` command will mark your Draft PR as "Ready for Review" and automatically add reviewers that are specified in the PR_REVIEWERS environment variable.
You can specify more than one reviewer using a comma-delimited string.

```bash
export PR_REVIEWERS=first-user,second-user,third-user
```

Add this to your shell rc file (`~/.zshrc` or `~/.bashrc`) and run `source <rc-file>`

```
Usage of add-reviewers:
  -poll-frequency duration
    	Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration (default 5m0s)
  -reviewers string
    	Comma-separated list of Github usernames to add as reviewers
  -silent
    	Whether to use voice output
  -when-checks-pass
    	Poll until all checks pass before adding reviewers (default true)
  <pullRequestNumber>
```

#### git-merge-pr

`git-merge-pr <pull request number>`

Add the given PR to the merge queue

#### git-prs

`git-prs`

Lists all of your open PRs. Useful for copying PR numbers.

## Installation

Clone repository and then:

```bash
# Install Github CLI
brew install gh 
# Setup login for Github CLI
gh auth login 
# Add the /bin directory to your PATH. Replace the directory below to wherever you cloned the repository.
# For example if using zsh and cloned in your home directory:
echo "export PATH=\$PATH:\$HOME/stacked-diff-workflow/bin" >> ~/.zshrc
source ~/.zshrc
```

## Example Workflow

### Creating and Updating PRs

Use **new-pr** and **git-update-pr** to create and update PR's while always staying on `main` branch.

### To Update Main

Once a PR has been merged, just rebase main normally. The local PR commit will be replaced by the one that Github created when squasing and merging.

```bash
git fetch && git rebase origin/main
```

### To Fix a Merge Conflict

If you have a merge conflict on your PR, you can use **replace-head** to keep your local `main` up to date.

```bash
# switch to feature branch that has a merge conflict
git-checkout <commit hash or PR number> 
# rebase or merge
git fetch && git rebase origin/main
# ... and address any merge conflicts
# Update your PR
git push origin/xxx 
git switch main
git rebase origin/main
# hit same merge conflicts, use the replace-merge script to copy the fixes you just made
replace-head
# continue with the rebase
git add . && git rebase --continue
# All done... now both the feature branch and your local main are rebased with main, and the merge conflicts only had to be fixed once
```

## Migration from create-symlinks

The pre 1.0 version of these scripts use symlinks. They are no longer required with 1.0+.

```
rm /usr/local/bin/full-assemble
rm /usr/local/bin/assemble
rm /usr/local/bin/git-updatepr
rm /usr/local/bin/git-newpr
rm /usr/local/bin/installapp
rm /usr/local/bin/gitlog
rm /usr/local/bin/git-amendpr
rm /usr/local/bin/git-reset-main
rm /usr/local/bin/git-prs
rm /usr/local/bin/git-review
rm /usr/local/bin/git-merge-pr
rm /usr/local/bin/squash-commits.go
rm /usr/local/bin/git-get-commit-branch
rm /usr/local/bin/git-checkout
```

## Building Scripts Yourself

If you want to build the scripts yourself for some reason, perhaps to try out a change, use `make build` from the project directory. The go output binaries are saved under [bin](bin)
