package main

import "testing"

func TestNoBase(t *testing.T) {
	for base, want := range map[string]bool{
		"": true,
		"0000000000000000000000000000000000000000": true,
		"abc1234":     false,
		"origin/main": false,
	} {
		if got := noBase(base); got != want {
			t.Errorf("noBase(%q) = %v, want %v", base, got, want)
		}
	}
}
