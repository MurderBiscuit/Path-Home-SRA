# Structural Reentry Assistance — process risk review

This is a lightweight Failure Mode and Effects Analysis (FMEA) **for service design**, not a risk score on individuals. Prioritize severity of potential harm and how reliably the app can prevent it. Assess with lived-experience reviewers, legal aid and local service organizations before public release.

| Process breakdown | Potential harm | Existing or proposed error-proofing | Residual verification |
|---|---|---|---|
| Reporting date guessed or missing | Missed mandatory appointment | Explicit unknown values; require confirmed calendar times; call issuing office if unclear | Real-user testing and signed conditions |
| Unclear parole travel or housing rule | Unintentional violation | No automatic permissions; show verified primary sources and prompt officer/attorney confirmation | State-by-state legal review |
| Birth-certificate request requires ID the user lacks | Circular eligibility barrier | “Get Unstuck” shows alternative proof/referral path | Confirm eligibility rules with issuing agency |
| Document fee exceeds cash available | Unmet basic ID needs | Estimate costs, show unknowns, waivers, and assistance paths | Confirm live fees and waiver availability |
| Transportation agency cannot serve address or schedule | Missed essential appointment | Reachable local transit/211 referral and pre-trip confirmation | Validate transport contact details |
| Reentry program rejects an application | Repeated unsuccessful referrals | Ask for denial reason, appeal/alternative proof, referral elsewhere | Partner follow-up (voluntary only) |
| Notification reveals parole context on shared device | Privacy violation, stigma, safety concerns | Default discreet language; minimal on-device notifications | Actual device lock-screen accessibility/security review |
| Legal reference outdated or overturned | Misleading guidance | Source links, dated reviews, warnings | Qualified review, later-case validity checks |
| Screen is confusing to first-time phone users | Abandonment or missed step | Large labels, one question at a time, “Get Unstuck” path, accessibility tests | Moderated real-user tests |

**No user profiling:** Do not calculate a “failure risk,” “compliance score,” or individual recidivism prediction. Keep personal goals and barrier choices encrypted on-device. Use aggregate partner process audits and voluntary user studies, with explicit consent, for continuous improvement.

## v0.9.4: County-specific controls
Local help now uses exact state/county matches and service category routing. No current county -> prompt to choose county + 211 fallback. Program list must show source dates, 'before you go', and 'if denied'; test against cross-county referrals.
