#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

# The tools.go dependency pins code-generator alongside client-go in go.mod.
# Install from that module checkout instead of resolving a second version tag.
unset KUBE_CODEGEN_TAG
export GOBIN="${repo_root}/.bin"
export GOCACHE="${GOCACHE:-${repo_root}/.cache}"
mkdir -p "${GOBIN}"

# A fresh checkout can have module metadata without the generator source.
go mod download k8s.io/code-generator
codegen_dir="$(go list -m -f '{{.Dir}}' k8s.io/code-generator)"
# shellcheck source=/dev/null
source "${codegen_dir}/kube_codegen.sh"

input_dir="${repo_root}/internal/codegen/input"
trap 'rm -rf "${input_dir}"' EXIT
"${repo_root}/hack/prepare-codegen.sh" "${input_dir}"

rm -rf "${repo_root}/generated"
kube::codegen::gen_client \
  --with-applyconfig \
  --with-watch \
  --output-dir "${repo_root}/generated" \
  --output-pkg "github.com/s-urbaniak/tinkerbell-client-go/generated" \
  --boilerplate "${repo_root}/hack/boilerplate.go.txt" \
  --plural-exceptions Hardware:Hardware \
  "${input_dir}"

"${repo_root}/hack/rebind-generated.sh"
gofmt -w "${repo_root}/generated"
