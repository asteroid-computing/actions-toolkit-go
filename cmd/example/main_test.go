package main

import "testing"

func TestGreet(t *testing.T) {
	tests := []struct {
		who, actor, repo, want string
	}{
		{"", "", "", "Hello, world! Run by someone in a local checkout."},
		{"Mona", "octocat", "owner/repo", "Hello, Mona! Run by octocat in owner/repo."},
	}
	for _, tt := range tests {
		if got := greet(tt.who, tt.actor, tt.repo); got != tt.want {
			t.Errorf("greet(%q, %q, %q) = %q, want %q", tt.who, tt.actor, tt.repo, got, tt.want)
		}
	}
}
