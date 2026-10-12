package main

import (
	"slices"
	"testing"
)

func TestParseComponents(t *testing.T) {
	got, err := parseComponents(" example , ,example")
	if err != nil || !slices.Equal(got, []string{"example"}) {
		t.Errorf("parseComponents = %q, %v", got, err)
	}
	if got, err := parseComponents(""); err != nil || len(got) != 0 {
		t.Errorf("empty input = %q, %v; want nothing", got, err)
	}
	if _, err := parseComponents("example,nope"); err == nil {
		t.Error("unknown component accepted")
	}
}
