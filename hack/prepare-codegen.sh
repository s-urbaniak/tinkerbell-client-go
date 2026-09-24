#!/usr/bin/env bash
set -euo pipefail

input_dir="$1"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"
api_dir="$(go list -m -f '{{.Dir}}' github.com/tinkerbell/tinkerbell/api)"
module=github.com/s-urbaniak/tinkerbell-client-go
rm -rf "${input_dir}"

for group in bmc tinkerbell; do
  destination="${input_dir}/${group}/v1alpha1"
  mkdir -p "${destination}"
  if [[ "${group}" == bmc ]]; then
    kinds='Job Machine Task'
    group_name=bmc.tinkerbell.org
  else
    kinds='Hardware Template Workflow WorkflowRuleSet'
    group_name=tinkerbell.org
  fi

  for source in "${api_dir}/v1alpha1/${group}/"*.go; do
    [[ "${source}" == *_test.go ]] && continue
    awk -v group="${group}" -v kinds="${kinds}" -v module="${module}" '
      BEGIN {
        split(kinds, names, " ")
        for (i in names) roots[names[i]] = 1
      }
      /^package (bmc|tinkerbell)$/ { $0 = "package v1alpha1" }
      {
        gsub(/"github[.]com\/tinkerbell\/tinkerbell\/api\/v1alpha1\/bmc"/,
          "bmc \"" module "/internal/codegen/input/bmc/v1alpha1\"")
        if (group == "bmc" && /metav1[.]TypeMeta/)
          gsub(/json:""/, "json:\",inline\"")
        if ($1 == "type" && roots[$2] && $3 == "struct" && $4 == "{")
          print "// +genclient"
        print
      }
    ' "${source}" > "${destination}/$(basename "${source}")"
  done

  printf '// +groupName=%s\n// Package v1alpha1 contains temporary code-generation input.\npackage v1alpha1\n' \
    "${group_name}" > "${destination}/doc.go"
done
