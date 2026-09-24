# Temporary generator input

Kubernetes `kube_codegen.sh` discovers concrete API types in `group/version`
packages with `+genclient` comments. Tinkerbell publishes its API in
`version/group` packages and does not include those comments. The generation
script creates a temporary copy of the pinned API source under `input/`, runs
the official generator script against that copy, and removes it afterward.

The copy includes the complete field definitions so apply configuration
builders remain typed. Generated imports are rebound to the published API
module. The BMC `TypeMeta` JSON tags are normalized only in generator input;
the released API types are never edited. The singular `hardware` resource name
and Tinkerbell's `GroupVersion` symbol are handled after generation.
