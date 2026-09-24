# Generator input

The Kubernetes generators expect API packages in `group/version` order and
concrete root type declarations. Tinkerbell's published Go API keeps the stable
`version/group` import layout. These small packages supply the names, metadata,
and status markers that the generators inspect. They are not client API types.

`hack/generate.sh` rewrites generated imports to the published Tinkerbell API
module. Keep each stub in sync with the resource inventory and `/status`
subresources in Tinkerbell's CRDs. `Hardware` uses the singular resource name.

Apply builders need the complete API field definitions. During regeneration,
`hack/prepare_applygen.py` copies those definitions from the pinned published
API module into `applyinput/`, adds generator markers and normalizes BMC's
TypeMeta tags. `hack/generate.sh` removes the temporary copy when it finishes.
Generated imports are then rebound to the published API types.
