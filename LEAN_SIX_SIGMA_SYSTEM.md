# Path Home — Lean Six Sigma for the whole service

**Published by Genyxis Network. Version 0.9.4 (development).**

## Policy and purpose
The unit of improvement is a **broken process**, not a returning person. People are never assigned Lean, parole-compliance or recidivism scores. No device location, releases, legal situation, family relationships, or task status goes to Genyxis. Process improvements rely on independently collected **consented, de-identified and aggregated** partner reports; the apps work offline.

## Value-stream map

| Service value stream | Define: what matters to person | Measure: voluntary service-quality evidence | Analyze typical friction | Improve user-visible workflow | Control / audit |
|---|---|---|---|---|---|
| Release intake | Understand real conditions before planning | Time and misinterpretation to finish intake in usability testing | Too many fields, unfamiliar terms | 5 short screens, unknown dates allowed | Plain-language tests, signed-document verification label |
| Supervision dates | Never guess a required report or discharge | Tests that reject invented dates and times | Conflicting documents, missing time | Calendar entries require confirmed time | No auto-grant of legal permission; independent rule review |
| Legal library | Find the right primary law or case | Source-check coverage, broken links and supersession rate | Different jurisdictions, outdated cases | State filter, official source links, citations and warnings | Legal reviewer sign-off; automated link freshness, human citation review |
| Documents and ID | Obtain usable ID without unnecessary visits | Extra visits caused by missing proof, documents and fees | Circular ID requirements, wrong birth state | State-of-birth guidance, alternate documents, county-specific clerk | Every fee and rule carries source/date; no inferred eligibility |
| Money and waivers | Know actual gaps without hiding unknown costs | Unknown-fee rate, verified funding sources | False zero costs, programs out of money | Mark unknown separately, refer to local verified waiver programs | Quarterly fee spot-check; funding always confirm by phone |
| Housing | Reach the authorized county assessment entry point | False referrals and wrong-county routing | Generic shelter link, false program match | Match exact state+county to the HARA where verified | Source/area checks, recurring provider review |
| Transportation | Make a required appointment on time | Confirmed routes and booking-rule errors | Incorrect transit office or day/time | Phone and pre-trip questions for local transit | No implied excuse for missing supervision; time verified |
| Employment | Reach actual job help and training | Program-match accuracy | Generic job-search advice, eligibility misunderstanding | Exact county employment center and category filter | Validate service and funding eligibility and dates |
| Health/food/benefits | Find accurate eligibility and appeals | Failed referrals due to outdated requirements | Wrong program or deadline | State/county referral with national fallback when missing | Agency confirmations, escalation to legal aid where needed |
| Family preparation | Help safely without taking control | Consent test and revocation effectiveness | Helper overreach, premature sharing | Preparatory checklist, person-controlled claim and access | Least privilege, expiration, revocation tests |
| Stay on Track | Self-selected, supportive reminders | Opt-in comprehension, notification correctness in device tests | Excess alerts, stigmatizing language | Off by default, user-chosen places, gentle on-device prompt | Explicit consent, no company/location sharing, delete controls |
| Get Unstuck | A next step for an atypical barrier | Failed match and no-alternative rates in synthetic tests | Agency dead end, missing ID, no money | A barrier selection filters **actual verified county contacts** | Never claim funding, case acceptance, or parole approval |
| Accessibility | First-time users can finish key workflows | Moderated usability success across literacy/access needs | Unclear instructions, text size, hidden navigation | Plain labels, large touch targets, one action at a time | Device accessibility passes before releases |
| Security/privacy | User owns their records | Security test findings and review gate completion | Data leakage, insecure storage, permission creep | Windows DPAPI and Android Keystore encrypted local storage | Security review and OS runtime tests; no telemetry |
| Application delivery | Signed and accurate releases | Reproducible builds, version mismatches, crashes | Unbuilt Android APK, broken installers | Native Windows and Android code with known limitations | Signed builds, both-device tests, hashes and rollback |

## DMAIC operational cycle

1. **Define:** A named workflow and a human-understandable failure mode (e.g. 'a person makes two unnecessary trips to obtain a birth record'). Set a specific acceptance test before changing the app.
2. **Measure:** Use anonymous test cases and voluntary partner *process* statistics; do not gather real user profiles or device telemetry. Use small-cell suppression when reporting aggregated study outcomes.
3. **Analyze:** Apply 5 Whys, process maps and service-level FMEA to distinguish incorrect information, wrong catchment area, missing alternate-document route, scheduling mismatch, or program policy.
4. **Improve:** Fix the source record, app route, wording, alternative action or alert. Preserve the personal-data boundary and the exact Path Home logo and mission.
5. **Control:** Regression-test state/county restrictions, source age, permission handling, legal caution, accessibility, and privacy in Windows and Android builds. Keep a versioned audit trail and roll back unsafe records.

## Data quality contract for every local provider
Required fields: `state_code`, `county`, official `name`, contact `phone` or `website`, exact `scope`, `eligibility`, `before_you_go`, `if_denied`, `source_url`, and `verified_on`. A county/state match is **strict**; ZIP codes, GPS and adjacent counties are never guessed. A provider with no documented eligibility gets a confirmation warning. Duplicate or dead sources fail release review. Current funding, open-now and acceptance are **not** inferred. Re-verify official pages on a regular schedule and after reported closures or changes.

## Definition of done
- Release has working native Windows and Android builds tested on target devices.
- Unambiguous locality matching (including `Jackson County` as `Jackson`) and no cross-county false matches.
- Every local referral has a source record and an explicit eligibility caveat.
- User can start without a known county and receives a safe fallback, never a random office.
- Blocked task maps to relevant provider **category** and a next step; unresolved categories are transparent.
- No individual performance analytics, telemetry, compliance score, or company location storage.
- Written-source verification and human legal review of case-specific statements; previous findings retain effective dates.
- Accessibility and independent security QC completed before app-store deployment.

## Current evidence and limits
Current v0.9.4 contains **six researched Jackson County, Michigan example contacts**, verified on 2026-11-22 using official public agency pages. They are not live funding/eligibility confirmations. The other 50-state/ DC county-level directories have **not** been populated and cannot be advertised as complete. Nationwide full-text case-law ingestion, legal-review signoff and Android APK build are still open work. Nothing here demonstrates measured reductions in violations or improved human outcomes.
