package main

import "strings"

// Only match an exact verified county and state. Never infer that a nearby
// or statewide office serves a person, and never invent eligibility.
type localProvider struct {
	ID          string   `json:"id"`
	State       string   `json:"state_code"`
	County      string   `json:"county"`
	Name        string   `json:"name"`
	Categories  []string `json:"categories"`
	Address     string   `json:"address"`
	Phone       string   `json:"phone"`
	Website     string   `json:"website"`
	Source      string   `json:"source_url"`
	VerifiedOn  string   `json:"verified_on"`
	Scope       string   `json:"scope"`
	Eligibility string   `json:"eligibility"`
	BeforeYouGo string   `json:"before_you_go"`
	IfDenied    string   `json:"if_denied"`
}

func findLocalProviders(all []localProvider, state, county, need string) []localProvider {
	state = strings.ToUpper(strings.TrimSpace(state))
	county = strings.TrimSpace(county)
	if strings.HasSuffix(strings.ToLower(county), " county") {
		county = strings.TrimSpace(county[:len(county)-7])
	}
	if state == "" || county == "" {
		return nil
	}
	out := make([]localProvider, 0)
	for _, p := range all {
		if strings.ToUpper(p.State) != state || !strings.EqualFold(strings.TrimSpace(p.County), county) {
			continue
		}
		if need != "" && need != "All" {
			found := false
			for _, c := range p.Categories {
				if strings.EqualFold(c, need) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if !strings.HasPrefix(p.Source, "https://") || p.VerifiedOn == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
func barrierHelpCategory(v string) string {
	switch v {
	case "Money or fees":
		return "money"
	case "Missing ID or paperwork":
		return "documents"
	case "No transportation":
		return "transportation"
	case "Instructions unclear":
		return "legal"
	case "Need safe housing":
		return "housing"
	case "Program said no":
		return "legal"
	default:
		return "All"
	}
}
