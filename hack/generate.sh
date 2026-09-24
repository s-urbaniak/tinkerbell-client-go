#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

codegen_version=v0.36.3
module=github.com/s-urbaniak/tinkerbell-client-go
input_base="$repo_root/internal/codegen/apis"
output_root="$repo_root/generated"
apply_input="$repo_root/internal/codegen/applyinput"
export GOBIN="$repo_root/.bin"
export GOCACHE="${GOCACHE:-$repo_root/.cache}"
mkdir -p "$GOBIN"
trap 'rm -rf "$apply_input"' EXIT

go install "k8s.io/code-generator/cmd/applyconfiguration-gen@$codegen_version" \
  "k8s.io/code-generator/cmd/client-gen@$codegen_version" \
  "k8s.io/code-generator/cmd/lister-gen@$codegen_version" \
  "k8s.io/code-generator/cmd/informer-gen@$codegen_version"

rm -rf "$output_root"
python3 "$repo_root/hack/prepare_applygen.py"
"$GOBIN/applyconfiguration-gen" \
  --go-header-file "$repo_root/hack/boilerplate.go.txt" \
  --output-dir "$output_root/applyconfiguration" \
  --output-pkg "$module/generated/applyconfiguration" \
  "$module/internal/codegen/applyinput/bmc/v1alpha1" \
  "$module/internal/codegen/applyinput/tinkerbell/v1alpha1"

# Client generator input mirrors only root type names and status presence.
# Generated imports are rebound to the published API module below.
"$GOBIN/client-gen" \
  --go-header-file "$repo_root/hack/boilerplate.go.txt" \
  --output-dir "$output_root/clientset" \
  --output-pkg "$module/generated/clientset" \
  --clientset-name versioned \
  --apply-configuration-package "$module/generated/applyconfiguration" \
  --input-base "$input_base" \
  --plural-exceptions Hardware:Hardware \
  --input bmc/v1alpha1 --input tinkerbell/v1alpha1

"$GOBIN/lister-gen" \
  --go-header-file "$repo_root/hack/boilerplate.go.txt" \
  --output-dir "$output_root/listers" \
  --output-pkg "$module/generated/listers" \
  --plural-exceptions Hardware:Hardware \
  "$module/internal/codegen/apis/bmc/v1alpha1" \
  "$module/internal/codegen/apis/tinkerbell/v1alpha1"

"$GOBIN/informer-gen" \
  --go-header-file "$repo_root/hack/boilerplate.go.txt" \
  --output-dir "$output_root/informers" \
  --output-pkg "$module/generated/informers" \
  --versioned-clientset-package "$module/generated/clientset/versioned" \
  --listers-package "$module/generated/listers" \
  --plural-exceptions Hardware:Hardware \
  "$module/internal/codegen/apis/bmc/v1alpha1" \
  "$module/internal/codegen/apis/tinkerbell/v1alpha1"

python3 "$repo_root/hack/rebind.py"
gofmt -w "$output_root"
