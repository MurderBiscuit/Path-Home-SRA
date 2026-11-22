#!/usr/bin/env bash
set -euo pipefail
if [[ $# -ne 1 || ! "$1" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]]; then
  echo 'Usage: bash tools/publish_public_repo.sh OWNER/path-home-sra' >&2
  echo 'Target owner and repository name must be explicit.' >&2
  exit 2
fi
TARGET="$1"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
command -v gh >/dev/null || { echo 'Install GitHub CLI (gh) first.' >&2; exit 1; }
command -v git >/dev/null || { echo 'Install git first.' >&2; exit 1; }
gh auth status >/dev/null || { echo 'Sign into GitHub with gh auth login first.' >&2; exit 1; }
if git remote get-url origin >/dev/null 2>&1; then
  echo 'This working copy already has an origin; inspect it rather than republish blindly.' >&2; exit 1
fi
if [[ ! -d .git ]]; then git init -b main; fi
# QA requires icon-embedded Windows executable, but .gitignore keeps unsigned exe out of GitHub.
if [[ -f desktop/windows/PathHome_Windows_x64.exe ]]; then python qa/verify_release.py; fi
git add -A
if git diff --cached --name-only | grep -Ei '(\.env|\.pem|\.p12|\.pfx|\.key|\.jks|\.keystore|plan\.json|plan\.dpapi|user_profiles|private_plan|\.exe$)'; then
  echo 'Blocked: private/sensitive or unsigned binary file staged.' >&2; exit 1
fi
if git diff --cached --quiet; then echo 'No files staged to publish'; exit 1; fi
git commit -m 'Initialize Path Home SRA public source and quality gates'
gh repo create "$TARGET" --public --source . --remote origin --push
printf 'Published: https://github.com/%s\n' "$TARGET"
