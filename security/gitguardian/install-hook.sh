#!/usr/bin/env bash
# GitGuardian ggshield — pre-commit secrets scanning beyond what gitleaks covers (200+ providers).
set -euo pipefail
pip install --user ggshield
ggshield install --mode local
ggshield install --mode global
echo "ggshield installed; set GITGUARDIAN_API_KEY in your shell profile"
