# Release gates and known limits (2026-11-22)

**This is a development source bundle, not a public-facing release.**

| Gate | Status | Acceptance standard |
|---|---|---|
| Windows unit tests/vet/cross-build | Validated locally | All checks green |
| Android business rules | Validated locally | Pure Java tests green |
| Android full APK compilation | Pending CI / Android SDK | `gradle assembleDebug` green |
| Signed Android and Windows builds | Pending | Release signing + verification |
| Device usability and accessibility | Pending | TalkBack, keyboard/screen-reader, font scaling, contrast, rotation, small screens |
| Privacy/security review | Pending | Storage, permission, location, recovery and threat-model review |
| Case-law correctness, parole laws and local eligibility | Pending legal/provider review | Validated current sources, citations, cautions |
| Notifications and background reminders | Pending | Consent, lifecycle, permission and device test evidence |
| Family account handoff | Not implemented in current offline native builds | Secure consent-based integration tests |
| Nationwide providers | Not complete | State/county source checks and fallbacks |

**Safety priority:** Never automatically grant travel, housing, contact, driving, or other parole permissions based solely on general rules. Users must verify signed orders and relevant agency instructions. No publisher telemetry includes private user profiles or location.
