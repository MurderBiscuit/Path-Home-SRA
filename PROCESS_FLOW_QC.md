# Path Home v0.9.5 — End-to-end process-flow and regression QC

**Published by Genyxis Network.** **Development build.** 2026-11-22.

This release applies Lean Six Sigma across the system: **Define** the user outcome, **Measure** an objective test, **Analyze** failure causes, **Improve** the user flow, and **Control** with a permanent regression gate. The improvement unit is the *process*, not the person. No user surveillance, personal performance score, profile upload, or company collection of release conditions.

## User journey / failure response matrix

| Journey | What the person needs | Acceptance check / safe recovery | Status |
|---|---|---|---|
| Open app | Find the next action, Law Library, and help | Home, My Plan, Law Library, Ask, Help readily available | Source verified; on-device usability pending |
| Choose location | Correct referrals, no unrelated offices | State and county explicitly entered; changing state clears Android county; no random provider fallback | Static/logic verified; on-device pending |
| Choose supervision type | Avoid mixing state and federal legal rules | Residential state remains separate from federal supervision; `FED` records appear only with federal supervised release explicitly selected | Pure logic + structural tests passed |
| Enter release dates | Support period, first report and expected supervision end | Unknown dates allowed; reversed start/end rejected before next step; never treat estimated end as discharge | Pure logic + source checks passed |
| Calendar | Correct date/time without guessing | Date and confirmed 24-hour local time required before phone calendar handoff; OS calendar must be opened | Logic/static verified; live phone pending |
| Documents / money | Avoid an invented "enough money" conclusion | Invalid amounts identified; missing document fees cause **partial** estimate label; waivers and funding never guaranteed | Pure logic + source checks passed |
| Local assistance | One source-checked county provider and a backup | Only exact selected state+county; call-first instructions, eligibility and official source; 211 fallback for unresearched counties | Dataset + source checks passed; provider re-verification needed |
| Program denial | One useful next action | Barrier category maps to local provider category, or truthful no-match fallback | Go unit tests passed; device journeys pending |
| Legal question | Relevant, cited research, not invented permission | Unmatched questions fail closed; combined travel/housing etc. flagged as multi-topic; federal only when selected | Go and JVM rule tests passed; legal review outstanding |
| Law Library | Fast access to real case-law sources | Main navigation link; state and explicitly selected federal research labels; legal precedent warnings | Source inspected; no offline case-law database yet |
| Family preparation | Simple pre-release checklist | Checkboxes local to user; actual consent-based remote handoff clearly marked unavailable | Static verified; live family service pending |
| Stay on Track | Voluntary, private reminder | Android prompts only on manual/app-open checks; no background tracking or reports; Windows manual only | Source inspected; GPS/haptics device tests pending |
| Privacy / damaged storage | Preserve, never silently destroy data | Windows DPAPI save errors prevent navigation/close; Android decrypt errors block writes and ask before erasure | Go checks + Android structural checks; live OS failure tests pending |
| Navigation / multitap | Smooth repeatable screens | Android question results have a dedicated screen; back returns Home; on-rotation page restored; field writes batched and flushed before navigation/onPause | Source verified; Android instrumented tests pending |
| Identity / mission | Consistent approved branding | Product logo/icons unchanged from 0.9.4; mission text identical across Windows & Android; Genyxis remains publisher label | SHA-256 checks passed |

## Verified issues resolved in v0.9.5

1. Prevented Android help screen from drawing a duplicate 211 fallback button.
2. Corrected state switching to clear both the stored and visibly displayed county.
3. Fixed questions such as **"Can I move to a new address?"** being treated as county travel.
4. Prevented repeated searches from stacking duplicate legal-result cards.
5. Protected unreadable Android-encrypted profiles from automatic overwrite; user must explicitly approve erasure.
6. Reduced Android text-entry lag by batching encrypted writes, flushing before navigation and on app pause.
7. Prevented Windows from navigating or closing silently when the encrypted plan cannot be saved.
8. Rejected malformed, negative, NaN and infinite money inputs rather than converting them to zero without warning.
9. Blocked reversed supervision date order, while permitting unknown dates.
10. Kept incomplete document-fee budgets explicitly partial.
11. Fixed federal code **`FED`** being incorrectly checked as `US`, separating residence from supervision system.
12. Required distinct review of combined legal topics (such as housing plus county travel).
13. Made the Android date picker reopen on the previously selected date instead of unexpectedly resetting to today.
14. Marked the fifth intake step and linked the user back to **My Progress**.
15. Kept Windows launcher/application icon resources consistent with the approved Path Home brand.

## Automated checks run

- 28/28 package/data/privacy/branding/source **static QC assertions** passed (`python qa/verify_release.py`).
- 6/6 **Go unit tests** passed, including exact locality matching, barriers, dates, money, federal selection, and question categorization (`go test` pure modules).
- Windows code passed **`GOOS=windows GOARCH=amd64 go vet`** and **cross-compilation** as a native Windows GUI executable; seven Path Home icon sizes embedded into PE resources.
- Android pure Java workflow rule tests passed with desktop JDK (`javac`/`java`); full native Android compilation is **not** possible in this environment without the Android SDK. Missing Android dependencies are expected when parsing the Activity with plain `javac`.
- Mission and all product-branded images match v0.9.4 byte-for-byte; the public legal database is identical in both app projects.

## Open release blockers (not hidden or declared fixed)

**Not production or app-store ready** until: Android SDK Gradle build succeeds; Android physical-device/e2e/accessibility tests pass; Windows executable is run on supported Windows machines at different display scaling, keyboard navigation and screen sizes; digitally signed builds and independent updates are tested; background notifications/reminders are implemented and tested; full case-law collection, legal-review approvals, and official data-refresh process are delivered; the local assistance directory covers counties beyond the six Jackson pilot records; and consent-based cross-device family handoff is deployed and reviewed. The Windows executable is a developer build and is **not digitally signed**.

**Do not advertise** nationwide county-level referrals, automatic missed-appointment alerts, background GPS monitoring, production legal advice, or live family accounts in the present release. The model deliberately fails safely when a record is missing.

## Lean Six Sigma control rule

For every release, test the 15 journeys above, record failed flows and corrected causes, and require: no wrong-jurisdiction legal claims, no guessed deadlines or assistance providers, no default upload of private data, no unconsented location tracking, and no silent data loss. Use anonymous/synthetic test fixtures; aggregate only voluntary process metrics from providers, never private individual records.
