# Path Home — Lean Six Sigma continuous-improvement playbook

**Owner:** Genyxis Network · **Scope:** improve systems and access, never rate the worth or conduct of returning citizens.

## Why Lean Six Sigma belongs here

Reentry often involves repeated appointments, missing documents, conflicting instructions, eligibility dead ends, waiting, unnecessary trips, and uncertain handoffs. Those are **process defects and barriers**. The person encountering them is not a defect. The application uses simple language; DMAIC terminology is for the product and partner teams.

## DMAIC, applied to real reentry barriers

| Phase | Operational question | Deliverable | Guardrail |
|---|---|---|---|
| Define | Where are people losing time, money or access? | Map release-day reporting, ID acquisition, housing, benefits and referrals | Let affected people define the problems; don't assume all conditions are alike |
| Measure | How often do systems cause extra visits, failed referrals, unknown fees or missed information? | Count *process* friction, not a person's worth | No user telemetry; use de-identified voluntary partner service audits only |
| Analyze | What caused the obstacle: missing eligibility proof, unclear instructions, office hours, transportation, fee, referral failure? | Pareto chart of structural barriers, 5 Whys discussion with frontline staff and returning citizens | Never treat poverty or an individual's conviction as the root cause |
| Improve | What simpler path would actually work? | Plain-language requirements; verified alternatives; one-call referrals; offline reminders and locally saved next steps | Confirm legal conditions with qualified sources and supervising agencies |
| Control | Do improvements stay useful as laws, fees and offices change? | Source freshness reviews, provider verification, accessibility and usability tests, documented change control | Human review before showing law as definitive; rollback bad updates |

## First feature: I'm stuck — help me find a way (v0.9.3)

A single voluntary screen in Windows and Android. People can describe a goal, choose a barrier, and see **one practical next step**. Options include document fees, missing ID, transportation, program denial, unclear instructions, safe housing, and other situations. It links back to local-help resources. Choices and free-text goals are saved **only in the locally encrypted profile**. No uploads, score, tracking pixel, provider alerts or punitive dashboard.

## Suggested service-process indicators (not collected automatically)

- **First-contact resolution:** percentage of partner interactions resolved or referred to a specific reachable provider on the first contact.
- **Unnecessary trips:** average repeated office visits to complete an ID or housing intake because requirements were missing or unclear.
- **Document lead time:** days from eligible request to actual receipt of a birth record or photo ID, reported only as de-identified program aggregates.
- **Referral completion:** whether a referral reached a provider that actually accepted the eligibility conditions, not merely that a link was clicked.
- **Legal source freshness:** percentage of published legal summaries with verified official sources and dates still within policy review windows.
- **Accessibility:** task completion and comprehension in moderated, voluntary user studies with people inexperienced with phones, with large-print and low-literacy variants.

**Do not infer parole compliance, recidivism risk, or safety from app usage.** Do not send individual locations, case details, timelines, barrier selections or family data to Genyxis. If a partner wants impact evidence, publish aggregate outcomes from separate voluntary studies, with informed consent and small-cell suppression.

## Pilot improvement cycles

1. Michigan/Jackson pilot: map the document and first-reporting workflows with local organizations and returning residents. Verify every office, fee, and alternate-document route; establish a baseline from *partner operations*, not user devices.
2. Run one narrow improvement experiment at a time (e.g. offer a fee-waiver pathway before telling a person to save money). Test a 10-minute usability scenario and ask the participant whether the instructions make sense.
3. Review for unintended harm: wrong legal assumption, missing resource, misleading eligibility, inaccessible button, excessive notifications. Fix the process, then check again.
4. Expand by jurisdiction only after sources and practical alternatives have been verified locally; don't claim all 52 tracks have comprehensive up-to-date case law.

## Release gates

Do not release as legal guidance until legal and privacy review, Android compilation, real-device accessibility tests, Windows runtime tests, and signed installer/update verification. The barrier helper offers general steps and never grants permission to violate an individual's conditions.
