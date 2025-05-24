package interactive

import (
	"strings"

	"slices"

	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

/*
Example output of gh api collaborators:
gh api repos/joshallenit/gh-stacked-diff/collaborators
[

	{
	  "login": "xxx",
	  "id": 4293001,
	  "node_id": "MDQ6VXNlcjQyOTMwMDE=",
	  "avatar_url": "https://avatars.githubusercontent.com/u/4293001?v=4",
	  "gravatar_id": "",
	  "url": "https://api.github.com/users/joshallenit",
	  "html_url": "https://github.com/joshallenit",
	  "followers_url": "https://api.github.com/users/joshallenit/followers",
	  "following_url": "https://api.github.com/users/joshallenit/following{/other_user}",
	  "gists_url": "https://api.github.com/users/joshallenit/gists{/gist_id}",
	  "starred_url": "https://api.github.com/users/joshallenit/starred{/owner}{/repo}",
	  "subscriptions_url": "https://api.github.com/users/joshallenit/subscriptions",
	  "organizations_url": "https://api.github.com/users/joshallenit/orgs",
	  "repos_url": "https://api.github.com/users/joshallenit/repos",
	  "events_url": "https://api.github.com/users/joshallenit/events{/privacy}",
	  "received_events_url": "https://api.github.com/users/joshallenit/received_events",
	  "type": "User",
	  "user_view_type": "public",
	  "site_admin": false,
	  "permissions": {
	    "admin": true,
	    "maintain": true,
	    "push": true,
	    "triage": true,
	    "pull": true
	  },
	  "role_name": "admin"
	},
	{
	  "login": "xxxx",
	  "id": 79605685,
	  "node_id": "MDQ6VXNlcjc5NjA1Njg1",
	  "avatar_url": "https://avatars.githubusercontent.com/u/79605685?v=4",
	  "gravatar_id": "",
	  "url": "https://api.github.com/users/slack-jallen",
	  "html_url": "https://github.com/slack-jallen",
	  "followers_url": "https://api.github.com/users/slack-jallen/followers",
	  "following_url": "https://api.github.com/users/slack-jallen/following{/other_user}",
	  "gists_url": "https://api.github.com/users/slack-jallen/gists{/gist_id}",
	  "starred_url": "https://api.github.com/users/slack-jallen/starred{/owner}{/repo}",
	  "subscriptions_url": "https://api.github.com/users/slack-jallen/subscriptions",
	  "organizations_url": "https://api.github.com/users/slack-jallen/orgs",
	  "repos_url": "https://api.github.com/users/slack-jallen/repos",
	  "events_url": "https://api.github.com/users/slack-jallen/events{/privacy}",
	  "received_events_url": "https://api.github.com/users/slack-jallen/received_events",
	  "type": "User",
	  "user_view_type": "public",
	  "site_admin": false,
	  "permissions": {
	    "admin": false,
	    "maintain": false,
	    "push": true,
	    "triage": true,
	    "pull": true
	  },
	  "role_name": "write"
	}

]

Example output from: gh repo view --json nameWithOwner

	{
	  "nameWithOwner": "joshallenit/gh-stacked-diff"
	}
*/
func getAllCollaborators() []string {
	jq := ".[] | .login"
	out := util.ExecuteOrDie(util.ExecuteOptions{},
		"gh", "api", "repos/"+util.GetRepoNameWithOwner()+"/collaborators", "--paginate", "--cache", "6h", "--jq", jq)
	collaborators := strings.Fields(out)
	collaborators = removeCurrentUser(collaborators)
	slices.Sort(collaborators)
	return collaborators
}

func removeCurrentUser(users []string) []string {
	return slices.DeleteFunc(users, func(next string) bool {
		return next == util.GetLoggedInUsername()
	})
}
