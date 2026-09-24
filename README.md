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
versioned typed clients, a fake clientset, listers, a shared informer factory,
and apply-configuration builders for server-side apply. Typed and fake clients
expose `Apply` and `ApplyStatus` for these resources. Clients use the published
API objects directly. No generated client code is placed in
`tinkerbell/tinkerbell`.

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

For server-side apply, build a configuration and set a field manager:

```go
import (
    applytink "github.com/s-urbaniak/tinkerbell-client-go/generated/applyconfiguration/tinkerbell/v1alpha1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

hardware := applytink.Hardware("node-1", "default").
    WithSpec(applytink.HardwareSpec().WithAgentID("aa:bb:cc:dd:ee:ff"))
_, err = clients.TinkerbellV1alpha1().Hardware("default").Apply(
    ctx, hardware, metav1.ApplyOptions{FieldManager: "my-controller"},
)
```

## Regenerate

Requirement: Go 1.26. Run:

```sh
./hack/update-codegen.sh
go test ./...
./hack/verify-codegen.sh
```

`hack/update-codegen.sh` sources the pinned Kubernetes `kube_codegen.sh` and
calls `kube::codegen::gen_client --with-applyconfig --with-watch`. It generates
the clientset, fake clients, listers, informers, and apply builders together.
This is the same script entry point used by Kubernetes projects such as Agones
and KCP. Generator tools are installed in `.bin/` at v0.36.3, matching
`client-go`.

The pattern follows [Kubernetes code-generator's documented entry point](https://github.com/kubernetes/code-generator/blob/master/kube_codegen.sh),
the [Agones generation script](https://github.com/Agones-dev/agones/blob/main/build/build-image/gen-crd-code.sh),
and the [KCP code-generator guidance](https://github.com/kcp-dev/code-generator).

The published Tinkerbell API uses `version/group` package paths, while
`kube_codegen.sh` requires `group/version` input packages with `+genclient`
markers. `hack/prepare-codegen.sh` temporarily copies the full pinned API
sources into that layout and adds the markers. `hack/rebind-generated.sh` then
points generated imports to the published API module, uses its `GroupVersion`
symbol, and preserves the CRD's singular `hardware` resource path. The
temporary input is deleted after generation. No generated code enters the
Tinkerbell repository or imports the temporary packages.

This proof of concept does not supply an OpenAPI v2 schema to the apply
generator, so it does not emit `Extract*` helpers for reconstructing
field ownership from live objects. `Apply` and `ApplyStatus` use Kubernetes'
server-side apply patch protocol.

`hack/verify-codegen.sh` checks that regeneration leaves no diff in
`generated/`; CI runs it on every push and pull request.

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
