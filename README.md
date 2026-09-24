# Tinkerbell client-go

Community proof of concept for out-of-tree Tinkerbell clients, generated from
`github.com/tinkerbell/tinkerbell/api` v0.25.0. No client code is generated in
the Tinkerbell repository.

The `v1alpha1` clients cover:

| API group | Resources |
| --- | --- |
| `tinkerbell.org` | Hardware, Template, Workflow, WorkflowRuleSet |
| `bmc.tinkerbell.org` | Job, Machine, Task |

`generated/` contains a clientset, fake clients, listers, informers, and
server-side apply builders. Clients use the published Tinkerbell API types.

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

For server-side apply:

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

Requires Go 1.26:

```sh
./hack/update-codegen.sh
go test ./...
```

`hack/update-codegen.sh` runs the pinned Kubernetes `kube_codegen.sh` with
`--with-applyconfig --with-watch`. The generator version matches `client-go`
v0.36.3.

Tinkerbell's `version/group` API layout requires a temporary `group/version`
copy for generation. The scripts remove that copy and point generated imports
back to the published API. They also preserve the singular `hardware` resource
name.

`hack/verify-codegen.sh` checks that regeneration leaves `generated/` unchanged.
Apply builders currently omit `Extract*` helpers because no OpenAPI v2 schema
is supplied.
