package clientgo_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	applyconfig "github.com/s-urbaniak/tinkerbell-client-go/generated/applyconfiguration"
	applybmc "github.com/s-urbaniak/tinkerbell-client-go/generated/applyconfiguration/bmc/v1alpha1"
	applytink "github.com/s-urbaniak/tinkerbell-client-go/generated/applyconfiguration/tinkerbell/v1alpha1"
	clientset "github.com/s-urbaniak/tinkerbell-client-go/generated/clientset/versioned"
	fakeclient "github.com/s-urbaniak/tinkerbell-client-go/generated/clientset/versioned/fake"
	clientscheme "github.com/s-urbaniak/tinkerbell-client-go/generated/clientset/versioned/scheme"
	informers "github.com/s-urbaniak/tinkerbell-client-go/generated/informers/externalversions"
	bmc "github.com/tinkerbell/tinkerbell/api/v1alpha1/bmc"
	tink "github.com/tinkerbell/tinkerbell/api/v1alpha1/tinkerbell"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func TestSchemeRegistersPublishedTypes(t *testing.T) {
	for _, gvk := range []struct{ group, kind string }{
		{"tinkerbell.org", "Hardware"}, {"tinkerbell.org", "Template"},
		{"tinkerbell.org", "Workflow"}, {"tinkerbell.org", "WorkflowRuleSet"},
		{"bmc.tinkerbell.org", "Job"}, {"bmc.tinkerbell.org", "Machine"},
		{"bmc.tinkerbell.org", "Task"},
	} {
		version := "v1alpha1"
		obj, err := clientscheme.Scheme.New(tink.GroupVersion.WithKind(gvk.kind))
		if gvk.group == "bmc.tinkerbell.org" {
			obj, err = clientscheme.Scheme.New(bmc.GroupVersion.WithKind(gvk.kind))
		}
		if err != nil {
			t.Fatalf("%s/%s %s: %v", gvk.group, version, gvk.kind, err)
		}
		if obj == nil {
			t.Fatalf("missing %s", gvk.kind)
		}
	}
}

func TestRESTPathsAndStatus(t *testing.T) {
	var requests []string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if !strings.Contains(r.Header.Get("Accept"), "json") {
			t.Errorf("Accept header = %q", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/hardware/"):
			fmt.Fprint(w, `{"apiVersion":"tinkerbell.org/v1alpha1","kind":"Hardware","metadata":{"name":"node","namespace":"demo"}}`)
		case strings.Contains(r.URL.Path, "/jobs/"):
			fmt.Fprint(w, `{"apiVersion":"bmc.tinkerbell.org/v1alpha1","kind":"Job","metadata":{"name":"reboot","namespace":"demo"}}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			http.Error(w, "unexpected path", http.StatusNotFound)
		}
	})
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, r)
		return recorder.Result(), nil
	})}

	client, err := clientset.NewForConfigAndClient(&rest.Config{Host: "https://example.invalid"}, httpClient)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.TinkerbellV1alpha1().Hardware("demo").Get(context.Background(), "node", metav1.GetOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TinkerbellV1alpha1().Hardware("demo").UpdateStatus(context.Background(), &tink.Hardware{ObjectMeta: metav1.ObjectMeta{Name: "node", Namespace: "demo"}}, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.BmcV1alpha1().Jobs("demo").Get(context.Background(), "reboot", metav1.GetOptions{}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"GET /apis/tinkerbell.org/v1alpha1/namespaces/demo/hardware/node",
		"PUT /apis/tinkerbell.org/v1alpha1/namespaces/demo/hardware/node/status",
		"GET /apis/bmc.tinkerbell.org/v1alpha1/namespaces/demo/jobs/reboot",
	}
	if fmt.Sprint(requests) != fmt.Sprint(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
}

func TestFakeAndInformerForHardware(t *testing.T) {
	client := fakeclient.NewSimpleClientset()
	hardware := &tink.Hardware{ObjectMeta: metav1.ObjectMeta{Name: "node", Namespace: "demo", Labels: map[string]string{"role": "test"}}}
	if _, err := client.TinkerbellV1alpha1().Hardware("demo").Create(context.Background(), hardware, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	got, err := client.TinkerbellV1alpha1().Hardware("demo").List(context.Background(), metav1.ListOptions{LabelSelector: "role=test"})
	if err != nil || len(got.Items) != 1 {
		t.Fatalf("list = %v, %v", got, err)
	}
	factory := informers.NewSharedInformerFactory(client, 0)
	informer, err := factory.ForResource(tink.GroupVersion.WithResource("hardware"))
	if err != nil {
		t.Fatal(err)
	}
	if informer.Informer() != factory.Tinkerbell().V1alpha1().Hardware().Informer() {
		t.Fatal("ForResource did not share the typed informer")
	}
}

func TestServerSideApply(t *testing.T) {
	var requests []string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/apply-patch+yaml") {
			t.Errorf("Content-Type = %q", got)
		}
		if got := r.URL.Query().Get("fieldManager"); got != "client-go-test" {
			t.Errorf("fieldManager = %q", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var object map[string]any
		if err := json.Unmarshal(body, &object); err != nil {
			t.Fatal(err)
		}
		if object["apiVersion"] != "tinkerbell.org/v1alpha1" || object["kind"] != "Hardware" {
			t.Errorf("apply TypeMeta = %v", object)
		}
		if r.URL.Path == "/apis/tinkerbell.org/v1alpha1/namespaces/demo/hardware/node" {
			spec, ok := object["spec"].(map[string]any)
			if !ok || spec["agentID"] != "aa:bb:cc:dd:ee:ff" {
				t.Errorf("apply spec = %v", object["spec"])
			}
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"apiVersion":"tinkerbell.org/v1alpha1","kind":"Hardware","metadata":{"name":"node","namespace":"demo"}}`)
	})
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, r)
		return recorder.Result(), nil
	})}
	client, err := clientset.NewForConfigAndClient(&rest.Config{Host: "https://example.invalid"}, httpClient)
	if err != nil {
		t.Fatal(err)
	}
	hardware := applytink.Hardware("node", "demo").WithSpec(applytink.HardwareSpec().WithAgentID("aa:bb:cc:dd:ee:ff"))
	opts := metav1.ApplyOptions{FieldManager: "client-go-test"}
	if _, err := client.TinkerbellV1alpha1().Hardware("demo").Apply(context.Background(), hardware, opts); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TinkerbellV1alpha1().Hardware("demo").ApplyStatus(context.Background(), applytink.Hardware("node", "demo").WithStatus(applytink.HardwareStatus()), opts); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"PATCH /apis/tinkerbell.org/v1alpha1/namespaces/demo/hardware/node",
		"PATCH /apis/tinkerbell.org/v1alpha1/namespaces/demo/hardware/node/status",
	}
	if fmt.Sprint(requests) != fmt.Sprint(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	if _, ok := applyconfig.ForKind(bmc.GroupVersion.WithKind("Job")).(*applybmc.JobApplyConfiguration); !ok {
		t.Fatal("ForKind did not resolve BMC Job")
	}
	job := applybmc.Job("reboot", "demo")
	if job.Kind == nil || *job.Kind != "Job" || job.APIVersion == nil || *job.APIVersion != "bmc.tinkerbell.org/v1alpha1" {
		t.Fatalf("BMC Job TypeMeta = %#v", job.TypeMetaApplyConfiguration)
	}
}

// roundTripFunc keeps transport assertions off the network.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
