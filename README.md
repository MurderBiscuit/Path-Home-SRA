# Path Home: Structural Reentry Assistance — v0.9.6 (release-gate development build)

**Published by Genyxis Network.** Path Home's approved product name, logo, installation icons, navigation icons, colors and five-paragraph mission are unchanged.

Path Home helps people returning home from incarceration understand their **next steps, supervision dates, legal research sources, documents, expenses, local assistance, and progress**, with an optional family preparation checklist and a private Stay on Track reminder. It is built as separate native Windows and Android projects, **not as a website shortcut**.

## What changed in the full-process Lean Six Sigma QC pass

This version also fixes a duplicate Android `onBackPressed()` override that would otherwise prevent a full APK compilation, and adds CI quality gates to catch build errors automatically. The previous version resolves actual errors found in navigation, local referral state changes, legal question categorization, federal supervision references, encrypted profile recovery, amount validation, redundant buttons, and Android field-write responsiveness. See **`PROCESS_FLOW_QC.md`** for the end-to-end flow matrix, concrete fixes, tests, remaining blockers, and release gates.

The Law Library remains prominent on the home screen and Android bottom navigation. Legal research is **general information, not case-specific permission or legal advice**. References may be incomplete or superseded; confirm current orders and instructions with the relevant agency, attorney, or court.

## Packages

- `desktop/windows/PathHome_Windows_x64.exe`: native Windows GUI developer executable, locally encrypted with Windows DPAPI; not digitally signed and not yet tested on real Windows hardware.
- `android/`: native Android Studio project, local Android Keystore encryption; **no APK included** (Android SDK not available in this build environment).
- `shared/reentry-public.json`: bundled research records; 50 states, DC and federal legal tracks are represented, **not** a verified complete law database.
- `LOCAL_DIRECTORY_QC.md`: six Jackson County, Michigan pilot referrals only. Elsewhere, the app uses explicit no-match handling and 211 referral instead of guessing offices.
- `MISSION.md`, `PUBLISHER.md`, `branding/`: existing approved identity and mission, preserved.
- `LEAN_SIX_SIGMA_SYSTEM.md`, `PROCESS_FLOW_QC.md`: operating controls, FMEA-style failure analysis and regression gates.

## Testing

From the project root:

```bash
cd desktop/windows
go test barriers.go barriers_test.go localproviders.go localproviders_test.go workflow.go workflow_test.go
GOOS=windows GOARCH=amd64 go vet ./...
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags '-H windowsgui' -o PathHome_Windows_x64.exe .
cd ../..
python tools/embed_windows_icon.py --exe desktop/windows/PathHome_Windows_x64.exe --ico branding/PathHome.ico
python qa/verify_release.py
```

On a development PC with Java installed:

```bash
javac -d /tmp/pathhome-javatest android/app/src/main/java/org/pathhome/sra/WorkflowRules.java android/qa/WorkflowRulesTest.java
java -cp /tmp/pathhome-javatest org.pathhome.sra.WorkflowRulesTest
```

For full Android builds, install the Android SDK/Android Studio and run Gradle there. Do not claim this ZIP is an installable Android APK.

## Privacy and publication

Personal information remains on the device under OS-backed encryption. Location is optional and only used on the phone for self-chosen check-ins; it is not sent to Genyxis. No personal account records or user data are included in the source package. Current offline builds do not sync or share profiles across devices.

**Development status:** not for public release or operational legal compliance without independent legal, security, device and accessibility testing. **No public GitHub repository was created as part of this package.**

## Continuous quality control

Run `python qa/verify_release.py` and see `RELEASE_GATES.md` and `.github/workflows/quality.yml`. The GitHub Actions workflow compiles Android against the Android SDK so duplicate-method and missing-resource errors cannot slip through source-only tests.

**Important:** A plain Go rebuild does not automatically embed the Windows app icon. Always run `tools/embed_windows_icon.py` after creating the executable. The quality gate checks this and the GitHub workflow performs it automatically.

## Publish in an authorized, standalone GitHub repository

The connected GitHub integration in this environment has **no repository-creation action**, so a public GitHub URL is not claimed. The script `tools/publish_public_repo.sh` will create a new public repository under the explicitly supplied `OWNER/repository` after authentication via the GitHub CLI. Use a Genyxis-controlled GitHub namespace; the script never picks an organization automatically. It excludes unsigned Windows builds and checks for private files before committing.

Do not publish from a copy containing real user profiles, precise location records or private signing keys.