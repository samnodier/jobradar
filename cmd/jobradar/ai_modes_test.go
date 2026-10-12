package main

import "testing"

func TestPreparationModes(t *testing.T) {
	for _, mode := range []string{"", "evaluate", "resume", "cover-letter", "interview", "outreach"} {
		if prompt, ok := aiPreparationPrompt(mode); !ok || prompt == "" {
			t.Fatalf("missing prompt for %q", mode)
		}
	}
	if _, ok := aiPreparationPrompt("invent-credentials"); ok {
		t.Fatal("unknown modes must fail")
	}
}
