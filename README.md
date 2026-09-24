# Tinkerbell client-go generation proof of concept

This repository generates Kubernetes-style Go clients for Tinkerbell outside the
[tinkerbell/tinkerbell](https://github.com/tinkerbell/tinkerbell) source tree. It
is a community proof of concept, not an official Tinkerbell release.

The initial surface covers every `v1alpha1` root resource in the published
`github.com/tinkerbell/tinkerbell/api` v0.25.0 module:

| API group | Resources |
| --- | --- |
| `tinkerbell.org` | Hardware, Template, Workflow, WorkflowRuleSet |
| `bmc.tinkerbell.org` | Job, Machine, Task |

The generated packages under `generated/` provide an aggregate clientset,
versioned typed clients, a fake clientset, listers, and a shared informer
factory. Clients use the published API objects directly. No generated client
code is placed in `tinkerbell/tinkerbell`.

## Use

```go
import (
    clientset "github.com/s-urbaniak/tinkerbell-client-go/generated/clientset/versioned"
    informers "github.com/s-urbaniak/tinkerbell-client-go/generated/informers/externalversions"
)

clients, err := clientset.NewForConfig(restConfig)
if err != nil {
    return err
}
factory := informers.NewSharedInformerFactory(clients, 0)
_ = factory.Tinkerbell().V1alpha1().Hardware()
```

The `restConfig` value is a normal `*rest.Config`. This module is a proof of
concept; depend on a commit until it has a release tag.

## Regenerate

Requirements: Go 1.26 and Python 3. Run:

```sh
./hack/generate.sh
go test ./...
```

`hack/generate.sh` installs Kubernetes `client-gen`, `lister-gen`, and
`informer-gen` v0.36.3 into `.bin/`, then runs them entirely in this repository.
The generator input is under `internal/codegen/apis`. Those minimal concrete
stubs exist because Kubernetes' generators expect `group/version` packages and
do not discover aliases to Tinkerbell's public `version/group` API types. The
post-generation step binds all output to the published API module, changes the
scheme symbol to Tinkerbell's `GroupVersion`, and uses the CRD's singular
`hardware` resource path. The stubs are never imported by generated clients.

To check reproducibility, run generation and confirm `git diff --exit-code --
generated` succeeds.

## Scope and next step

The [original client-go PR](https://github.com/tinkerbell/tinkerbell/pull/932)
provided a handwritten Hardware proof of concept. The
[September 8 community meeting notes](https://docs.google.com/document/d/1Hmqrhj2rPjZ5W0DvRynFNY2cJq6jFCbNOc4p26U5Dgg/edit)
requested this out-of-tree generation spike.

`v1alpha2` is a separate next step: its Tinkerbell package in the current
published API module lacks `AddToScheme`, which generated aggregate and fake
clientsets require. Its BMC package also has a distinct nested Go import path.
Those API compatibility questions can be addressed here without changing the
Tinkerbell source tree.
