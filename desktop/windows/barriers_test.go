package main

import (
	"strings"
	"testing"
)

func TestBarrierNextActionAlwaysHelpful(t *testing.T) {
	for _, b := range []string{"Money or fees", "Missing ID or paperwork", "No transportation", "Program said no", "Instructions unclear", "Need safe housing", "Something else", ""} {
		text := barrierNextAction(b)
		if len(text) < 60 {
			t.Errorf("too little guidance for %q: %q", b, text)
		}
		for _, dangerous := range []string{"You are allowed", "You must travel", "guaranteed approval", "notify Genyxis"} {
			if strings.Contains(text, dangerous) {
				t.Errorf("unsafe instruction in %q", b)
			}
		}
	}
}
