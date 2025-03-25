# Stacked Diff Workflow

Using a [stacked diff workflow](https://newsletter.pragmaticengineer.com/p/stacked-diffs) allows you to break down a pull request into several smaller PRs. It also allows you to work on separate streams of work without the overhead of changing branches. Once you experience the efficiency of stacked diffs you can't imagine going back to your old workflow.

This project is a Command Line Interface that manages git commits and branches to allow you to quickly use a stacked diff workflow. It uses the Github CLI to create pull requests and add reviewers once PR checks have passed.

## Installation

### Installation as Github CLI Plugin

#### Mac

*Optional: As this is a CLI, do yourself a favor and install [iTerm](https://iterm2.com/) and [zsh](https://ohmyz.sh/), as they make working from the command line more pleasant.*

```bash
# Install Github CLI. 
brew install gh 
# Setup login for Github CLI
gh auth login
# Install plugin
gh extensions install joshallenit/gh-stacked-diff 
# Add a shell alias to make it faster to use. 
# For example if using zsh:
echo "alias sd='gh stacked-diff'" >> ~/.zshrc
source ~/.zshrc
```

#### Windows

1. Install [Git and Git Bash](https://gitforwindows.org/)
2. Install [Github CLI](https://cli.github.com/). Winget is possible: `winget install --id GitHub.cli`
3. Authenticate gh and install plugin:
      ```bash
      gh auth login 
      # Install plugin
      gh extensions install joshallenit/gh-stacked-diff 
      # Add a shell alias to make it faster to use. 
      # For example if using Git Bash:
      echo "alias sd='gh stacked-diff'" >> ~/.bashrc
      source ~/.bashrc
      ```

### Usage as a golang Library

The code can also be used as a go library within your own go application. See the [Developer Guide](DEVELOPER_GUIDE.md#usage-as-a-golang-library) for more info.

```bash
go get github.com/joshallenit/gh-stacked-diff/v2@v2.0.8
```

## Command Line Interface

```bash
usage: sd [top-level-flags] <command> [<args>]`

Possible commands are:

   add-reviewers       Add reviewers to Pull Request on Github once its checks have passed
   branch-name         Outputs branch name of commit
   checkout            Checks out branch associated with commit indicator
   code-owners         Outputs code owners for all of the changes in branch
   log                 Displays git log of your changes
   new                 Create a new pull request from a commit on main
   prs                 Lists all Pull Requests you have open.
   rebase-main         Bring your main branch up to date with remote
   replace-commit      Replaces a commit on main branch with its associated branch
   replace-conflicts   For failed rebase: replace changes with its associated branch
   update              Add commits from main to an existing PR
   wait-for-merge      Waits for a pull request to be merged

To learn more about a command use: sd <command> --help

flags:

  -log-level string
        Possible log levels:
           debug
           info
           warn
           error
        Default is info, except on commands that are for output purposes,
        (namely branch-name and log), which have a default of error.
```

### Basic Commands

#### log

Displays summary of the git commits on current branch that are not in the remote branch.

Useful to view list indexes, or copy commit hashes, to use for the commitIndicator required by other commands.

A ✅ means that there is a PR associated with the commit (actually it means there is a branch, but having a branch means there is a PR when using this workflow). If there is more than one commit on the associated branch, those commits are also listed (indented under the their associated commit summary).

```bash
usage: sd log
```

<img width="663" alt="image" src="https://user-images.githubusercontent.com/79605685/210386995-9c3e7179-24ed-4d59-9b3e-2b3b34aa6ccc.png">

#### new

Create a new PR with a cherry-pick of the given commit indicator.

This command first creates an associated branch, (with a name based on the commit summary), and then uses Github CLI to create a PR.

Can also add reviewers once PR checks have passed, see "--reviewers" flag.

```bash
usage: sd new [flags] [commitIndicator (default is HEAD commit on main)]

Ticket Number:

If you prefix a (Jira-like formatted) ticket number to the git commit
summary then the "Ticket" section of the PR description will be
populated with it.

For example:

"CONV-9999 Add new feature"

Templates:

The Pull Request Title, Body (aka Description), and Branch Name are
created from golang templates.

The default templates are:

   branch-name.template:      templates/config/branch-name.template
   pr-description.template:   templates/config/pr-description.template
   pr-title.template:         templates/config/pr-title.template

To change a template, copy the default from templates/config/ into
~/.gh-stacked-diff/ and modify contents.

The possible values for the templates are:

   CommitBody                   Body of the commit message
   CommitSummary                Summary line of the commit message
   CommitSummaryCleaned         Summary line of the commit message without
                                spaces or special characters
   CommitSummaryWithoutTicket   Summary line of the commit message without
                                the prefix of the ticket number
   FeatureFlag                  Value passed to feature-flag flag
   TicketNumber                 Jira ticket as parsed from the commit summary
   Username                     Name as parsed from git config email.
   UsernameCleaned              Username with dots (.) converted to dashes (-).

flags:

  -base string
        Base branch for Pull Request (default "main")
  -draft
        Whether to create the PR as draft (default true)
  -feature-flag string
        Value for FEATURE_FLAG in PR description
  -indicator string
        Indicator type to use to interpret commitIndicator:
           commit   a commit hash, can be abbreviated,
           pr       a github Pull Request number,
           list     the order of commit listed in the git log, as indicated
                    by "sd log"
           guess    the command will guess the indicator type:
              Number between 0 and 99:       list
              Number between 100 and 999999: pr
              Otherwise:                     commit
         (default "guess")
  -min-checks int
        Minimum number of checks to wait for before verifying that checks
        have passed before adding reviewers. It takes some time for checks
        to be added to a PR by Github, and if you add-reviewers too soon it
        will think that they have all passed. (default 4)
  -reviewers string
        Comma-separated list of Github usernames to add as reviewers once
        checks have passed.
```

<img width="938" alt="image" src="https://user-images.githubusercontent.com/79605685/210406914-9b43f0e0-ac11-498f-bdd7-5a48e07dcbc0.png">

###### Note on Commit Messages

Keep your commit summary to a [reasonable length](https://www.midori-global.com/blog/2018/04/02/git-50-72-rule). The commit summary is used as the branch name. To add more detail use the [commit description](https://stackoverflow.com/questions/40505643/how-to-do-a-git-commit-with-a-subject-line-and-message-body/40506149#40506149). The
created branch name is truncated to 120 chars as Github has problems with very long
branch names.


#### update

Add commits from local main branch to an existing PR.

Can also add reviewers once PR checks have passed, see "--reviewers" flag.

```bash
usage: sd update [flags] [PR commitIndicator [fixup commitIndicator (defaults to head commit) [fixup commitIndicator...]]]

If commitIndicator are missing then you will be prompted to select commits:

   [enter]    confirms selection
   [space]    adds to selection when selecting commits to add
   [up,k]     moves cursor up
   [down,j]   moves cursor down
   [q,esc]    cancels

flags:

  -indicator string
        Indicator type to use to interpret commitIndicator:
           commit   a commit hash, can be abbreviated,
           pr       a github Pull Request number,
           list     the order of commit listed in the git log, as indicated
                    by "sd log"
           guess    the command will guess the indicator type:
              Number between 0 and 99:       list
              Number between 100 and 999999: pr
              Otherwise:                     commit
         (default "guess")
  -min-checks int
        Minimum number of checks to wait for before verifying that checks
        have passed before adding reviewers. It takes some time for checks
        to be added to a PR by Github, and if you add-reviewers too soon it
        will think that they have all passed. (default 4)
  -reviewers string
        Comma-separated list of Github usernames to add as reviewers once
        checks have passed.
```

#### add-reviewers

Add reviewers to Pull Request on Github once its checks have passed.

If PR is marked as a Draft, it is first marked as "Ready for Review".

```bash
usage: sd add-reviewers [flags] [commitIndicator [commitIndicator]...]

flags:

  -indicator string
        Indicator type to use to interpret commitIndicator:
           commit   a commit hash, can be abbreviated,
           pr       a github Pull Request number,
           list     the order of commit listed in the git log, as indicated
                    by "sd log"
           guess    the command will guess the indicator type:
              Number between 0 and 99:       list
              Number between 100 and 999999: pr
              Otherwise:                     commit
         (default "guess")
  -min-checks int
        Minimum number of checks to wait for before verifying that checks
        have passed before adding reviewers. It takes some time for checks
        to be added to a PR by Github, and if you add-reviewers too soon it
        will think that they have all passed. (default 4)
  -poll-frequency duration
        Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration (default 30s)
  -reviewers string
        Comma-separated list of Github usernames to add as reviewers once
        checks have passed.
        Falls back to PR_REVIEWERS environment variable.
  -when-checks-pass
        Poll until all checks pass before adding reviewers (default true)
```

<img width="904" alt="image" src="https://user-images.githubusercontent.com/79605685/210428712-bcea3ce7-e70f-4982-aa54-48e166221a1d.png">

###### Reviewers

You can specify more than one reviewer using a comma-delimited string.

To use the environment variable instead of the "--reviewers" flag:

```bash
export PR_REVIEWERS=first-user,second-user,third-user
```

Add this to your shell rc file (`~/.zshrc` or `~/.bashrc`) and run `source <rc-file>`

### Commands for Rebasing and Fixing Merge Conflicts

#### rebase-main

Rebase with origin/main, dropping any commits who's associated branches have been merged.

This avoids having to manually call "git reset --hard head" whenever you have merge conflicts with a commit that has already been merged but has slight variation with local main because, for example, a change was made with the Github Web UI.

```bash
usage: sd rebase-main
```

#### checkout

Checks out the branch associated with commit indicator.

For when you want to merge only the branch with with origin/main, rather than your entire local main branch, verify why CI is failing on that particular branch, or for any other reason.

After modifying the branch you can use "sd replace-commit" to sync local main.

```bash
usage: sd checkout [flags] <commitIndicator>

flags:

  -indicator string
        Indicator type to use to interpret commitIndicator:
           commit   a commit hash, can be abbreviated,
           pr       a github Pull Request number,
           list     the order of commit listed in the git log, as indicated by "sd log"
           guess    the command will guess the indicator type:
              Number between 0 and 99:       list
              Number between 100 and 999999: pr
              Otherwise:                     commit
         (default "guess")
```

#### replace-commit

Replaces a commit on main branch with the squashed contents of its associated branch.

This is useful when you make changes within a branch, for example to fix a problem found on CI, and want to bring the changes over to your local main branch.

```bash
usage: sd replace-commit [flags] <commitIndicator>

flags:

  -indicator string
        Indicator type to use to interpret commitIndicator:
           commit   a commit hash, can be abbreviated,
           pr       a github Pull Request number,
           list     the order of commit listed in the git log, as indicated
                    by "sd log"
           guess    the command will guess the indicator type:
              Number between 0 and 99:       list
              Number between 100 and 999999: pr
              Otherwise:                     commit
         (default "guess")
```

#### replace-conflicts

During a rebase that failed because of merge conflicts, replace the current uncommitted changes (merge conflicts), with the contents (diff between origin/main and HEAD) of its associated branch.

```bash
usage: sd replace-conflicts

flags:

  -confirm
        Whether to automatically confirm to do this rather than ask for y/n input
```

### Commands for Custom Scripting

#### branch-name

Outputs the branch name for a given commit indicator. Useful for your own custom scripting.

```bash
usage: sd branch-name [flags] <commitIndicator>

flags:

  -indicator string
        Indicator type to use to interpret commitIndicator:
           commit   a commit hash, can be abbreviated,
           pr       a github Pull Request number,
           list     the order of commit listed in the git log, as indicated
                    by "sd log"
           guess    the command will guess the indicator type:
              Number between 0 and 99:       list
              Number between 100 and 999999: pr
              Otherwise:                     commit
         (default "guess")
```

#### wait-for-merge

Waits for a pull request to be merged. Polls PR every 30 seconds.

Useful for your own custom scripting.

```bash
usage: sd main [flags] <commit hash or pull request number>

flags:

  -indicator string
        Indicator type to use to interpret commitIndicator:
           commit   a commit hash, can be abbreviated,
           pr       a github Pull Request number,
           list     the order of commit listed in the git log, as indicated
                    by "sd log"
           guess    the command will guess the indicator type:
              Number between 0 and 99:       list
              Number between 100 and 999999: pr
              Otherwise:                     commit
         (default "guess")
```

### Other Commands

#### code-owners

Outputs code owners for each file that has been modified in the current local branch when compared to the remote main branch

```bash
usage: sd code-owners
```

#### prs

Lists all of your open PRs. Useful for copying PR numbers.

```bash
usage: sd prs
```

## Example Workflow

### Creating and Updating PRs

Use **sd new** and **sd update** to create and update PR's while always staying on `main` branch.

### To Update Main

*Note: This process is automated by the `sd rebase-main` command. There is no need to follow these steps manually.*

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
      sd checkout <commitIndicator> 
      git fetch && git merge origin/main
      # ... and address any merge conflicts
      # Update your PR
      git push origin/xxx 
      ```
2. Merge PR via Github
3. [Update your Main Branch](#to-update-main)

##### Advanced Flow

If you want to update your main branch *before* you merge your PR, you can use **replace-conflicts** to keep your local `main` up to date.

```bash
# switch to feature branch that has a merge conflict
sd checkout <commitIndicator> 
# rebase or merge
git fetch && git merge origin/main
# ... and address any merge conflicts
# Update your PR
git push origin/xxx 
# Rebase your local main branch.
git switch main
git rebase origin/main
# hit same merge conflicts, use replace-head to copy the fixes you just made
replace-head <commitIndicator>
# continue with the rebase
git add . && git rebase --continue
# All done... now both the feature branch and your local main are rebased with main, 
# and the merge conflicts only had to be fixed once
```

## Building Source and Contributing

See the [Developer Guide](DEVELOPER_GUIDE.md), which includes instructions on how to build the source, as well as an overview of the code.

## Stacked Pull Requests?

Note: these scripts do *not* facilitate Stacked *Pull Requests*. Github does some things that add friction to using Stacked PR's, even with support from third party software. For example, after merging one of the PR's in the stack, the other PR's will require a re-review. Instead of Stacked PRs, it's recommended to organize your PR's, as much as reasonably possible, so that they can be all be rebased against main at the same time. When there are dependencies, wait for dependant PR to be merged before putting up the next one. You may find that often you are still working on the next commit while the other is being reviewed/merged.

## Acknowledgments

- Thanks to [Dave Lee](https://x.com/kastiglione) for publishing [this article](https://kastiglione.github.io/git/2020/09/11/git-stacked-commits.html) that inspired the first version of the scripts.  

- Thanks to the Github team for creating [their CLI](https://cli.github.com/) that is leveraged here.

## Version Compatibility

| Stacked Diff version | gh CLI versions tested | git versions tested |
| -------------------- | ---------------------- | ------------------- |
| [2.0.0](CHANGELOG.md#200---2025-02-28) | 2.66.1, 2.64.0 | 2.47.1, 2.48.1 |
