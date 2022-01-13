/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package decoder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

const (
	testLabel            = "labelvalue"
	serviceAccountPrefix = "example-sa*"
)

func TestDecode(t *testing.T) {
	testYAML := filepath.Join("testdata", "example-configmap-1.yaml")
	f, err := os.Open(testYAML)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cfg := v1.ConfigMap{}
	if err := Decode(f, &cfg); err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Data["foo.cfg"]; !ok {
		t.Fatal("key foo.cfg not found in decoded ConfigMap")
	}
}

func TestDecodeUnstructuredCRD(t *testing.T) {
	testYAML := filepath.Join("testdata", "fake-crd.yaml")
	f, err := os.Open(testYAML)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	obj, err := DecodeAny(f)
	if err != nil {
		t.Fatal(err)
	}
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		t.Fatalf("expected unstructured.Unstructured, got %T", u)
	}

	if _, ok := u.Object["spec"]; !ok {
		t.Fatalf("spec field of CRD not found")
	}

	spec, ok := u.Object["spec"].(map[string]interface{})
	if !ok {
		t.Fatalf("spec not expected map[string]interface{}, got: %T", u.Object["spec"])
	}

	example, ok := spec["example"].(string)
	if !ok {
		t.Fatalf("spec.example not expectedstring, got: %T", spec["example"])
	}
	if example != "value" {
		t.Fatalf("spec.example not expected 'value', got %q", spec["example"])
	}
}

func TestDecodeAny(t *testing.T) {
	testYAML := filepath.Join("testdata", "example-configmap-3.json")
	f, err := os.Open(testYAML)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if obj, err := DecodeAny(f); err != nil {
		t.Fatal(err)
	} else if cfg, ok := obj.(*v1.ConfigMap); !ok && cfg.Data["foo.cfg"] != "" {
		t.Fatal("key foo.cfg not found in decoded ConfigMap")
	} else if _, ok := cfg.Data["foo.cfg"]; !ok {
		t.Fatal("key foo.cfg not found in decoded ConfigMap")
	}
}

func TestDecodeFile(t *testing.T) {
	testYAML := "example-configmap-1.yaml"
	testdata := os.DirFS("testdata")

	cfg := v1.ConfigMap{}
	if err := DecodeFile(testdata, testYAML, &cfg, MutateOption(func(o k8s.Object) error {
		obj, ok := o.(*v1.ConfigMap)
		if !ok {
			t.Fatalf("unexpected type %T not ConfigMap", o)
		}
		if obj.ObjectMeta.Labels == nil {
			obj.Labels = make(map[string]string)
		}
		obj.ObjectMeta.Labels["inject-value"] = "test123"
		return nil
	})); err != nil {
		t.Fatal(err)
	}
	if cfg.ObjectMeta.Labels["inject-value"] != "test123" {
		t.Fatal("injected label value not found", cfg.ObjectMeta.Labels)
	}
	cfg = v1.ConfigMap{}
	if err := DecodeFile(testdata, testYAML, &cfg, MutateLabels(map[string]string{"injected": testLabel})); err != nil {
		t.Fatal(err)
	}
	if cfg.ObjectMeta.Labels["injected"] != testLabel {
		t.Fatal("injected label value not found", cfg.ObjectMeta.Labels)
	}
}

func TestDecodeEachFile(t *testing.T) {
	testdata := os.DirFS(filepath.Join("testdata", "examples"))

	count := 0
	if err := DecodeEachFile(context.TODO(), testdata, serviceAccountPrefix, func(ctx context.Context, obj k8s.Object) error {
		count++
		return nil
	}); err != nil {
		t.Fatal(err)
	} else if expected := 3; count != expected {
		t.Fatalf("expected %d objects, got: %d", expected, count)
	}
	// load `testdata/examples/*`
	count = 0
	serviceAccounts := 0
	configs := 0
	if err := DecodeEachFile(context.TODO(), testdata, "*", func(ctx context.Context, obj k8s.Object) error {
		count++
		switch obj.(type) {
		case *v1.ConfigMap:
			configs++
		case *v1.ServiceAccount:
			serviceAccounts++
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	} else if expected := 4; count != expected {
		t.Fatalf("expected %d objects, got: %d", expected, count)
	} else if expected := 3; expected != serviceAccounts {
		t.Fatalf("expected %d serviceAccounts got %d", expected, serviceAccounts)
	} else if expected := 1; expected != configs {
		t.Fatalf("expected %d configs got %d", expected, configs)
	}
}

func TestDecodeAllFiles(t *testing.T) {
	// load `testdata/examples/example-sa*`
	testdata := os.DirFS(filepath.Join("testdata", "examples"))
	if objects, err := DecodeAllFiles(context.TODO(), testdata, serviceAccountPrefix); err != nil {
		t.Fatal(err)
	} else if expected, got := 3, len(objects); got != expected {
		t.Fatalf("expected %d objects, got: %d", expected, got)
	}
	// load `testdata/examples/*`
	if objects, err := DecodeAllFiles(context.TODO(), testdata, "*"); err != nil {
		t.Fatal(err)
	} else if expected, got := 4, len(objects); got != expected {
		t.Fatalf("expected %d objects, got: %d", expected, got)
	}
}

func TestDecodeEach(t *testing.T) {
	testYAML := filepath.Join("testdata", "example-multidoc-1.yaml")
	f, err := os.Open(testYAML)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	count := 0
	err = DecodeEach(context.TODO(), f, func(ctx context.Context, obj k8s.Object) error {
		count++
		switch cfg := obj.(type) {
		case *v1.ConfigMap:
			if _, ok := cfg.Data["foo"]; !ok {
				t.Fatalf("expected key 'foo' in ConfigMap.Data, got: %v", cfg.Data)
			}
		default:
			t.Fatalf("unexpected type returned not ConfigMap: %T", cfg)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	} else if count != 2 {
		t.Fatalf("expected 2 documents, got: %d", count)
	}
}

func TestDecodeAll(t *testing.T) {
	testYAML := filepath.Join("testdata", "example-multidoc-1.yaml")
	f, err := os.Open(testYAML)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if objects, err := DecodeAll(context.TODO(), f); err != nil {
		t.Fatal(err)
	} else if expected, got := 2, len(objects); got != expected {
		t.Fatalf("expected 2 documents, got: %d", got)
	}
}

func TestDecodersWithMutateFunc(t *testing.T) {
	t.Run("DecodeAny", func(t *testing.T) {
		testYAML := filepath.Join("testdata", "example-configmap-3.json")
		f, err := os.Open(testYAML)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if obj, err := DecodeAny(f, MutateLabels(map[string]string{"injected": testLabel})); err != nil {
			t.Fatal(err)
		} else if cfg, ok := obj.(*v1.ConfigMap); !ok && cfg.Data["foo.cfg"] != "" {
			t.Fatal("key foo.cfg not found in decoded ConfigMap")
		} else if cfg.ObjectMeta.Labels["injected"] != testLabel {
			t.Fatal("injected label value not found", cfg.ObjectMeta.Labels)
		}
	})
	t.Run("DecodeEach", func(t *testing.T) {
		testdata := os.DirFS(filepath.Join("testdata", "examples"))
		if err := DecodeEachFile(context.TODO(), testdata, serviceAccountPrefix, func(ctx context.Context, obj k8s.Object) error {
			if labels := obj.GetLabels(); labels["injected"] != testLabel {
				t.Fatalf("unexpected value in labels: %q", labels["injected"])
			}
			return nil
		}, MutateLabels(map[string]string{"injected": testLabel})); err != nil {
			t.Fatal(err)
		}
	})
}

func TestHandlerFuncs(t *testing.T) {
	handlerNS := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "handler-test"}}
	res, err := resources.New(cfg)
	if err != nil {
		t.Fatalf("Error creating new resources object: %v", err)
	}
	err = res.Create(context.TODO(), handlerNS)
	if err != nil {
		t.Fatalf("error while creating namespace %q: %s", handlerNS.Name, err)
	}
	testdata := os.DirFS(filepath.Join("testdata", "examples"))
	patches := []DecodeOption{MutateNamespace(handlerNS.Name), MutateLabels(map[string]string{"injected": testLabel})}
	// lookup all objects to use for verification / deletion steps later on
	objects, err := DecodeAllFiles(context.TODO(), testdata, "*", patches...)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("DecodeEach_Create", func(t *testing.T) {
		if err := DecodeEachFile(context.TODO(), testdata, "*", CreateHandler(res), patches...); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("DecodeEach_ReadHandler", func(t *testing.T) {
		count := 0
		serviceAccounts := 0
		configs := 0
		if err := DecodeEachFile(context.TODO(), testdata, "*", ReadHandler(res, func(ctx context.Context, obj k8s.Object) error {
			if labels := obj.GetLabels(); labels["injected"] != testLabel {
				t.Fatalf("unexpected value in labels: %q", labels["injected"])
			} else {
				count++
				switch cfg := obj.(type) {
				case *v1.ConfigMap:
					if _, ok := cfg.Data["foo.cfg"]; !ok {
						t.Fatalf("expected key 'foo.cfg' in ConfigMap.Data, got: %v", cfg.Data)
					}
					configs++
				case *v1.ServiceAccount:
					serviceAccounts++
				default:
					t.Fatalf("unexpected type returned not ConfigMap: %T", cfg)
				}
			}
			return nil
		}), MutateNamespace(handlerNS.Name)); err != nil {
			t.Fatal(err)
		}

		if expected := 4; count != expected {
			t.Fatalf("expected %d objects, got: %d", expected, count)
		} else if expected := 3; expected != serviceAccounts {
			t.Fatalf("expected %d serviceAccounts got %d", expected, serviceAccounts)
		} else if expected := 1; expected != configs {
			t.Fatalf("expected %d configs got %d", expected, configs)
		}
	})

	t.Run("DecodeEach_Delete", func(t *testing.T) {
		if err := DecodeEachFile(context.TODO(), testdata, "*", DeleteHandler(res), patches...); err != nil {
			t.Fatal(err)
		}

		t.Run("Verify", func(t *testing.T) {
			count := 0
			for i := range objects {
				if err := IgnoreErrorHandler(ReadHandler(res, func(ctx context.Context, obj k8s.Object) error {
					t.Logf("Object { apiVersion: %q; Kind:%q; Namespace:%q; Name:%q } found", obj.GetObjectKind().GroupVersionKind().Version, obj.GetObjectKind().GroupVersionKind().Kind, obj.GetNamespace(), obj.GetName())
					count++
					return nil
				}), apierrors.IsNotFound)(ctx, objects[i]); err != nil {
					t.Fatal(err)
				}
			}
			if count > 0 {
				t.Fatalf("%d test objects were not deleted", count)
			}
		})
	})
}
