package main

// barrierNextAction provides general, non-legal next steps. It never decides supervision permissions.
func barrierNextAction(barrier string) string {
	switch barrier {
	case "Money or fees":
		return "Ask the issuing agency about a fee waiver. If one is not available, ask 211 or a local reentry organization about help with document costs."
	case "Missing ID or paperwork":
		return "Ask the agency exactly what documents it accepts and whether other proof can work. Start with birth records, then Social Security, then state ID."
	case "No transportation":
		return "Check local transit and 211 for a confirmed ride. If a required appointment is at risk, contact the supervising office promptly for instructions."
	case "Program said no":
		return "Ask for the specific eligibility reason, any appeal process, and what other proof is accepted. Ask for a referral to another program."
	case "Instructions unclear":
		return "Compare the instructions with your signed documents. Contact the issuing agency, your attorney, or legal aid for clarification before acting."
	case "Need safe housing":
		return "Ask 211 or a reentry housing service about available options. If you have supervision, confirm the proposed address with your supervising office."
	case "Something else":
		return "Write down the immediate problem and who can answer it. Ask the first agency you contact for a specific referral if they cannot help."
	default:
		return "Choose one reason above to see a practical next step. You can always find local help without saving anything."
	}
}
