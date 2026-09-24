#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
"${repo_root}/hack/update-codegen.sh"
cd "${repo_root}"
git diff --exit-code -- generated
