package main

import "testing"

func TestQuestionTopic(t *testing.T) {
	tests := map[string]string{
		"Can I move to a new address?":   "housing",
		"Can I live with my sister?":     "housing",
		"Can I travel out of state?":     "travel",
		"Can I go outside my county?":    "travel",
		"Can I vote?":                    "voting",
		"How do I report to my officer?": "reporting",
		"Can I get a job?":               "employment",
		"Am I allowed a new phone?":      "",
		"Can I use a network?":           "",
		"Where do I get paperwork?":      "",
	}
	for q, want := range tests {
		if got := questionTopic(q); got != want {
			t.Errorf("%q => %q; want %q", q, got, want)
		}
	}
}
func TestDatesAndMoney(t *testing.T) {
	if dateOrderValid("2028-01-20", "2027-01-20") {
		t.Fatal("backwards date order accepted")
	}
	if dateOrderValid("2027-02-30", "2027-03-01") || dateOrderValid("2027-01-01", "2027-99-99") {
		t.Fatal("invalid calendar dates accepted")
	}
	if !dateOrderValid("", "2027-01-20") {
		t.Fatal("unknown start blocked")
	}
	for _, amount := range []string{"-1", "NaN", "Inf", "one hundred", "10000001", "$1$0"} {
		if validateMoneyField(amount) == nil {
			t.Errorf("bad amount %q accepted", amount)
		}
	}
	for _, amount := range []string{"", "0", "12.25", "$15"} {
		if validateMoneyField(amount) != nil {
			t.Errorf("valid %q rejected", amount)
		}
	}
}

func TestFederalSupervisionIsExplicit(t *testing.T) {
	if !federalSupervision("Federal supervised release") {
		t.Fatal("federal supervision undetected")
	}
	if federalSupervision("state parole") {
		t.Fatal("state parole received federal treatment")
	}
	if federalSupervision("not federal") {
		t.Fatal("negative should not equal federal")
	}
}
