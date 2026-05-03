package main

import (
	"strings"
	"testing"
)

func TestUsageMentionsAllSubcommands(t *testing.T) {
	for _, sub := range []string{"new", "find", "ls", "help"} {
		if !strings.Contains(usage, sub) {
			t.Errorf("usage missing subcommand %q", sub)
		}
	}
}
