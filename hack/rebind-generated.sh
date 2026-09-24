#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
module=github.com/s-urbaniak/tinkerbell-client-go

# The generator input is a temporary group/version view of the released API.
# Point generated clients back to the public version/group packages.
while IFS= read -r -d '' file; do
  sed -E \
    -e "s#${module}/internal/codegen/input/bmc/v1alpha1#github.com/tinkerbell/tinkerbell/api/v1alpha1/bmc#g" \
    -e "s#${module}/internal/codegen/input/tinkerbell/v1alpha1#github.com/tinkerbell/tinkerbell/api/v1alpha1/tinkerbell#g" \
    -e 's/SchemeGroupVersion/GroupVersion/g' \
    -e 's/Resource: "hardwares"/Resource: "hardware"/g' \
    "${file}" > "${file}.tmp"
  mv "${file}.tmp" "${file}"

  if [[ "${file}" == */listers/* ]]; then
    sed -E 's/([[:alnum:]_]+)\.Resource\("([^"]+)"\)/\1.GroupVersion.WithResource("\2").GroupResource()/g' \
      "${file}" > "${file}.tmp"
    mv "${file}.tmp" "${file}"
  fi
done < <(find "${repo_root}/generated" -name '*.go' -print0)
