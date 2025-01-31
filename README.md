# Developer Scripts for Stacked Diff Workflow

These scripts make it easier to build from the command line and to create and update PR's with Github. They facilitates a [stacked diff workflow](https://kastiglione.github.io/git/2020/09/11/git-stacked-commits.html), where you always commit on `main` branch and have can have multiple streams of work all on `main`.

Join the discussion in [#devel-stacked-diff-workflow](https://slack-pde.slack.com/archives/C03V94N2A84)

Note: these scripts do *not* facilitate Stacked *Pull Requests*. Github does some things that add friction to using Stacked PR's, even with support from third party software. For example, after merging one of the PR's in the stack, the other PR's will require a re-review. Instead, it's recommended to organize your PR's, as much as reasonably possible, so that they can be all be rebased against main at the same time. When there are dependencies, wait for dependant PR to be merged before putting up the next one. You may find that often you are still working on the next commit while the other is being reviewed/merged.

## TL;DR

Using a stacked diff workflow like this allows you to work on separate streams of work without changing branches. It also makes creating and updating Pull Requests faster as you do not have to manually switch branches.

## Installation

Everything is terminal based, so do yourself a favor and install [iTerm](https://iterm2.com/) and [zsh](https://ohmyz.sh/) to make your life better.

Clone the repository or download the [latest release](https://github.com/tinyspeck/stacked-diff-workflow/releases), and then:

```bash
# Install Github CLI
brew install gh 
# Setup login for Github CLI
gh auth login 
# Add the /bin directory to your PATH. 
# Replace the directory below to wherever you cloned the repository or unzipped the release.
# For example if using zsh and cloned in your home directory:
echo "export PATH=\$PATH:\$HOME/stacked-diff-workflow/bin" >> ~/.zshrc
source ~/.zshrc
```

## Scripts

### For Stacked Diff Workflow

*Note: the scripts attempts to use `main` as the main branch, and if that branch does not exist it looks for `master`.*

#### Script: git-checkout

`git-checkout <commit hash or pr number>`

Checkout the feature branch associated with a given PR or commit. For when you want to checkout the feature branch to rebase with origin/main, merge with origin/main, or for any other reason. After modifying the feature branch use `replac-commit` or `replace-head` to sync local `main`.

#### Script: gitlog

`gitlog`

Abbreviated git log that only shows what has changed, useful for copying commit hashes. A ✅ means that there is a PR associated with the commit (actually means there is a branch, but having a branch means there is a PR when using this workflow).

<img width="663" alt="image" src="https://user-images.githubusercontent.com/79605685/210386995-9c3e7179-24ed-4d59-9b3e-2b3b34aa6ccc.png">

#### Script: new-pr

Create a new PR with a cherry-pick of the given commit hash

```
new-pr [flags] [commit hash to make PR for (default is top commit on main)]
Flags:
  -base string
    	Base branch for Pull Request (default "main")
  -draft
    	Whether to create the PR as draft (default true)
  -feature-flag string
    	Value for FEATURE_FLAG in PR description (default "None")
  -logFlags int
    	Log flags, see https://pkg.go.dev/log#pkg-constants
```

###### Note on Commit Messages

Keep your commit summary to a [reasonable length](https://www.midori-global.com/blog/2018/04/02/git-50-72-rule). The commit summary is used as the branch name. To add more detail use the [commit description](https://stackoverflow.com/questions/40505643/how-to-do-a-git-commit-with-a-subject-line-and-message-body/40506149#40506149).

###### Ticket Number

If you prefix the Jira ticket to the git commit summary then `newpr` will populate the `Ticket` section of the PR description.

For example:
`CONV-9999 Add new feature`

###### Templates

The Pull Request Title, Body (aka Description), and Branch Name are created from [golang templates](https://pkg.go.dev/text/template). The defaults are:

- [branch-name.template](src/config/branch-name.template)
- [pr-description.template](src/config/pr-description.template)
- [pr-title.template](src/config/pr-title.template)

The [possible values](config/templates) for the templates are:

- **TicketNumber** - Jira ticket as parsed from the commit summary
- **Username** -  Name as parsed from git config email
- **CommitBody** - Body of the commit message
- **CommitSummary** - Summary line of the commit message
- **CommitSummaryCleaned** - Summary line of the commit message without spaces or special characters
- **CommitSummaryWithoutTicket** - Summary line of the commit message without the prefix of the ticket number
- **FeatureFlag** - Name of feature flag given from command line
- **CodeOwners** - List of matching files and owners from the CODEOWNERS of the repository, if any

To change a template, copy the default from [src/config/](src/config/) into `~/.stacked-diff-workflow/` and modify.

<img width="938" alt="image" src="https://user-images.githubusercontent.com/79605685/210406914-9b43f0e0-ac11-498f-bdd7-5a48e07dcbc0.png">

#### Script: replace-commit

`replace-commit <commit hash or pr number>`

Reset the main branch with the squashed contents of the given commits associated branch. Sometimes you might want to switch to a feature branch and make changes to it (rebase, amend). With this script you can then ensure that your `main` branch is up to date.

#### Script: replace-head

`replace-head`

Use during rebase of main branch to use the contents of a feature branch that already fixed the merge conflicts.

#### Script: update-pr

```bash
update-pr <commitHash or pullRequestNumber> [fixup commit (defaults to top commit)] [other fixup commit...]
Flags:
  -logFlags int
    	Log flags, see https://pkg.go.dev/log#pkg-constants
```

Add one or more commits to a PR.

### To Help You Build

*Note: Only [Android](https://github.com/tinyspeck/slack-android-ng) build scripts are currently supported.*

#### Script: assemble-app

`assemble-app`

Calls `./gradlew assembleInternalDebug` and build tests. Use "-s" (silent) flag to not use voice (`say`) to announce success/failure. Any options after, or in-lieu of, "-s" will be passed to the `./gradle` command, for example `--rerun-tasks`.

#### Script: install-app

`install-app`

Calls `./gradlew assembleInternalDebug` and install on real device. Use "-s" (silent) flag to not use voice (`say`) to announce success/failure. Any options after, or in-lieu of, "-s" will be passed to the `./gradle` command, for example `--rerun-tasks`.

#### Script: install-apk

`install-apk`

Installs the already compiled APK on a real devices. Useful for after you have run `install-app` but forgot to plugin in your phone 😄. It's faster than running `install-app` again as it doesn't run gradle.

### To Help with Github

#### Script: add-reviewers

```
add-reviewers <pullRequestNumber or commitHash>
Flags:
  -poll-frequency duration
    	Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration (default 5m0s)
  -reviewers string
    	Comma-separated list of Github usernames to add as reviewers
  -silent
    	Whether to use voice output
  -when-checks-pass
    	Poll until all checks pass before adding reviewers (default true)
```

The `add-reviewers` command will mark your Draft PR as "Ready for Review" and automatically add reviewers that are specified in the PR_REVIEWERS environment variable.
You can specify more than one reviewer using a comma-delimited string.

```bash
export PR_REVIEWERS=first-user,second-user,third-user
```

Add this to your shell rc file (`~/.zshrc` or `~/.bashrc`) and run `source <rc-file>`

<img width="904" alt="image" src="https://user-images.githubusercontent.com/79605685/210428712-bcea3ce7-e70f-4982-aa54-48e166221a1d.png">

#### Script: git-merge-pr

`git-merge-pr <pull request number>`

Add the given PR to the merge queue

#### Script: git-prs

`git-prs`

Lists all of your open PRs. Useful for copying PR numbers.

<img width="904" alt="image" src="https://github.com/tinyspeck/stacked-diff-workflow/assets/79605685/7e7a5708-58dc-4060-96b9-89615a86c009">

## Example Workflow

### Creating and Updating PRs

Use **new-pr** and **update-pr** to create and update PR's while always staying on `main` branch.

### To Update Main

Once a PR has been merged, just rebase main normally. The local PR commit will be replaced by the one that Github created when squasing and merging.

```bash
git fetch && git rebase origin/main
```

If you run into conflicts with a commit that has already been merged you can just ignore it. This can happen, for example, if a change was made on github.com and it is not reflected in your local commit. Obviously, only do this if the PR has actually already been merged into main! The error message from rebase will let you know which commit has conflicts.

```bash
git reset --hard head && git rebase --continue
```

#### To Fix Merge Conflicts

##### Easy Flow

If you just are rebasing with `main` and the commit with merge conflict has already been **merged**, then the process is simpler.

1. Fix Merge Conflict

```bash
# switch to feature branch that has a merge conflict
git-checkout <commit hash or PR number> 
git fetch && git merge origin/main
# ... and address any merge conflicts
# Update your PR
git push origin/xxx 
```

2\. Merge PR via Github

3\. Update your Main Branch

```bash
# Checkout main if not already there:
git switch main
# Rebase as normal:
git fetch && git rebase origin/main
# rebase complains about a merge conflict from the merged PR...
# Essentially delete the commit locally, it's already been merged so there is no need to fix it again:
git reset --hard HEAD
# Then continue:
git rebase --continue
```

##### Longer Running Flow

If you want to update your main branch *before* you merge your PR, you can use **replace-head** to keep your local `main` up to date.

```bash
# switch to feature branch that has a merge conflict
git-checkout <commit hash or PR number> 
# rebase or merge
git fetch && git merge origin/main
# ... and address any merge conflicts
# Update your PR
git push origin/xxx 
# Rebase your local main branch.
git switch main
git rebase origin/main
# hit same merge conflicts, use the replace-merge script to copy the fixes you just made
replace-head
# continue with the rebase
git add . && git rebase --continue
# All done... now both the feature branch and your local main are rebased with main, and the merge conflicts only had to be fixed once
```

# How To Build Scripts

The scripts are written in golang and bash.

1. Install [golang](https://go.dev/dl/).
2. Install make. This is already installed on Mac, but to instructions for windows are [here](https://leangaurav.medium.com/how-to-setup-install-gnu-make-on-windows-324480f1da69).

Then run:

```bash
make build
```


	// so I could do something like this:
	// https://github.com/Nutlope/aicommits
	// to create an AI git commit message

	/*
					https://www.hatica.io/blog/ai-commit-tools/
				https://github.com/kamushadenes/chloe/blob/main/.github/scripts/release-notes.py
		the commit messages don't look that great, probably not the best idea
		but splitting commits into smaller ones based on compilability could be good. Maybe figuring out which dependencies are declared in which file and then which dependencies are used in each file? That could work


				I could actually try these commit messages to see how well they work or don't work
				Other ideas:
					- show git log very fast and then check for branches async using ANSI codes
					- output the latest head commit so that a rollback can be done "git reset --hard xxx", but that doesn't keep track of what it does to branches which could screw up a PR, undo? that could be useful.
					- Include PR link instead of just a :green-check: beside PR's, there might be a way to http link from terminal. Too long to do it for every branch, would require the ANSI backspace idea. Or it could run in the background and update it next? Not sure if Go can launch background tasks and terminate, but probably? Or do it with a visual gui? wow, that would be something. https://github.com/charmbracelet/bubbletea
					- Next `gh pr view 83824 --json latestReviews` and ensure developer is already not approved so that the review is not dismissed
					- show the output of git commands in a tabbed window that uses ANSI escape codes to move around the screen
					- better error handling so that it reverted on error rather then leaving in an indeterminite state... but wouldn't this mean that I have to save error codes so they can be reported upstream?

POLISH
- update README with latest changes
- code review, review/organize/document code
- standardize all test so they are using assert* and named properly
- test for error conditions (ugh not fun) + but could lead to better error handling (rollback)
- get rid of the check for master when / vs - in branch name
					
NOT WORTH DOING


						*/
