package actioncontext_test

import (
	"fmt"

	"github.com/google/go-github/v92/github"

	"github.com/asteroid-computing/actions-toolkit-go/actioncontext"
)

// Example_pullRequest demonstrates constructing a local ActionContext with a
// pull_request payload and using the typed PullRequestEvent accessor.
func Example_pullRequest() {
	ac := &actioncontext.ActionContext{
		EventName:    "pull_request",
		EventPayload: &github.PullRequestEvent{Number: new(42)},
	}

	pr, ok := ac.PullRequestEvent()
	fmt.Println(pr.GetNumber(), ok)
	// Output:
	// 42 true
}
