# Path Home — Lean Six Sigma QC delta: v0.9.6

## Define

Reduce avoidable release and legal-guidance errors in Windows and Android while preserving all Path Home branding, publisher credits, private local storage and mission text.

## Measure: defects reproduced

| Defect | Severity | Root cause | Correction | Verification |
|---|---|---|---|---|
| Duplicate `onBackPressed()` implementation in Android | Build blocker | Source-only checks did not compile the full Android Activity | Removed duplicate; added duplicate-signature gate and GitHub full Android compile job | Local duplicate-signature gate passes; full SDK build still pending |
| Windows rebuild lost the approved app icon | Installation branding defect | Icon is embedded after Go cross-compile rather than during build | Added reusable `tools/embed_windows_icon.py` and CI step, validated `.rsrc` | Windows PE resource check passes |
| QA depended on a previous local directory | Public repository build blocker | Previous baseline path was not standalone | Replaced dependency with brand SHA-256 manifest | QA gate passes using standalone folder |
| Dollar input with embedded `$` was accepted | Budget input defect | Broad string replacement deleted all dollar signs | Only a single leading currency mark permitted | Go and Java regression tests pass |
| Nonexistent supervision date could pass ordering test | Planning/data validation defect | String comparisons without calendar parsing | Parse dates as real calendar dates, keep unknown dates allowed | Go and Java negative tests pass |
| `network` and `paperwork` were classified as employment | Legal guidance category defect | Substring `work` matched unrelated words | Whole-word employment terms and test cases | Go and Java tests pass |
| CI initially expected checked-in unsigned EXE | CI design defect | Ignore rule intentionally excludes binaries | Build Windows binary, attach icon, **then** run QA | Workflow source audited; GitHub execution pending |

## Analyze and control

- Every change must preserve a real first-time user's path to local providers, Law Library, document help, supervision plan and privacy.
- Tests must distinguish **unit-tested**, **compiled**, **device-tested**, and **legally approved**. None are interchangeable.
- Do not publish personally identifiable release documents or background locations. Public data assets are not case-specific legal advice.
- CI fails when legal rules lose citations, brand hashes change unexpectedly, or Android resource references are missing.

## Release-blocking work still outstanding

1. Full Android SDK compilation and physical-device accessibility tests.
2. Signed installers and update channels.
3. Independent legal review of case law and supervision law, and ongoing provider eligibility checks.
4. Privacy and security assessment, realistic failure recovery, and consent-based family handoff.
5. Real Windows hardware UI and assistive-tech testing.

**Publisher: Genyxis Network.** Official product: **Path Home: Structural Reentry Assistance**.
