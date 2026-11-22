# Contributing to Path Home

Path Home: Structural Reentry Assistance is published by Genyxis Network. The approved Path Home product logo and application icons must not be modified without explicit product approval.

## Lean Six Sigma quality gate

1. **Define:** State the user impact and the failing journey. Focus on release dates, legal sources, identity documents, fees, transport, local services, and privacy.
2. **Measure:** Write the expected result and a repeatable test **before** claiming a fix.
3. **Analyze:** Identify root cause, not symptoms. Check cross-screen state, offline behavior, permissions, encryption, and legal evidence.
4. **Improve:** Make the smallest complete correction that leaves adjacent workflows intact.
5. **Control:** Run the automated tests, check source provenance, include a regression test, and record outstanding risk.

## Required acceptance checks

- Windows Go tests/vet, Windows x64 cross-build.
- Native Android Gradle compile and Android Java rules tests on CI.
- Public source and branding integrity validation (`python qa/verify_release.py`).
- No personally identifying release records, precise locations, passwords, signing keys, or private API tokens in repository, test fixtures, analytics, or logs.
- All new legal content has source URL, relevant jurisdiction, effective/review dates, and appropriate review status.
- All provider referrals identify a real location, contact details, source and date, eligibility caveats and a fallback.
- A real device accessibility review is needed before public release; CI passing does *not* establish legal or safety readiness.

## Change discipline

Do not change Path Home branding or the five-paragraph mission without explicit permission. Don't add clinical claims or automated decisions about whether someone's conduct is permitted by parole conditions.

## Licensing

The project owner has not yet adopted an open-source license. Public visibility alone is not a grant to reuse, redistribute or modify the work. Do not add third-party files without recording their licenses.
