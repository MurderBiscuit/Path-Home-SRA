package main

import "testing"

func TestLocalProviderJurisdiction(t *testing.T) {
	p := []localProvider{{ID: "j", State: "MI", County: "Jackson", Categories: []string{"housing"}, Source: "https://example.org", VerifiedOn: "2026-11-22"}, {ID: "b", State: "MI", County: "Barry", Categories: []string{"housing"}, Source: "https://example.org", VerifiedOn: "2026-11-22"}, {ID: "a", State: "AL", County: "Jackson", Categories: []string{"housing"}, Source: "https://example.org", VerifiedOn: "2026-11-22"}}
	if got := findLocalProviders(p, "MI", "Jackson", "housing"); len(got) != 1 || got[0].ID != "j" {
		t.Fatalf("cross-county leakage: %+v", got)
	}
	if got := findLocalProviders(p, "MI", "", "All"); len(got) != 0 {
		t.Fatalf("empty county may not show local programs")
	}
	if got := findLocalProviders(p, "MI", "Jackson", "documents"); len(got) != 0 {
		t.Fatalf("category mismatch")
	}
}
func TestBarrierCategory(t *testing.T) {
	if barrierHelpCategory("No transportation") != "transportation" {
		t.Fatal("wrong route")
	}
}
