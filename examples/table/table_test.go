/*
Copyright 2021 The Kubernetes Authors.

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

package table

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var test = env.New()

func TestMain(m *testing.M) {
	// Setup the rand number source and a limit
	test.Setup(func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		return context.WithValue(context.WithValue(ctx, "limit", rand.Int31n(255)), "randsrc", rnd), nil
	})

	// Don't forget to launch the package test
	os.Exit(test.Run(m))
}

func TestTableDriven(t *testing.T) {
	// feature 1
	table0 := features.Table{
		features.TableRow{
			Name:        "testWithDescription",
			Description: "This is an example of how to create a test with description",
			Assessment: func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				if time.Now().Unix() < 0 {
					t.Errorf("Looks like the current time could not be determined")
				}
				return ctx
			},
		},
		features.TableRow{
			Name: "less than equal 64",
			Assessment: func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
				rnd := ctx.Value("randsrc").(*rand.Rand) // in real test, check asserted type
				lim := ctx.Value("limit").(int32)        // check type assertion
				if rnd.Int31n(lim) > 64 {
					t.Log("limit should be less than 64")
				}
				return ctx
			},
		},
		features.TableRow{
			Name: "more than than equal 128",
			Assessment: func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
				rnd := ctx.Value("randsrc").(*rand.Rand) // in real test, check asserted type
				lim := ctx.Value("limit").(int32)        // check type assertion
				if rnd.Int31n(lim) > 128 {
					t.Log("limit should be less than 128")
				}
				return ctx
			},
		},
		features.TableRow{
			Assessment: func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
				rnd := ctx.Value("randsrc").(*rand.Rand) // in real test, check asserted type
				lim := ctx.Value("limit").(int32)        // check type assertion
				if rnd.Int31n(lim) > 256 {
					t.Log("limit should be less than 256")
				}
				return ctx
			},
		},
	}.Build("Random numbers", "Test Features can have description too. Second Argument to the Build call is a description.").Feature()

	// feature 2
	table1 := features.Table{
		features.TableRow{
			Name: "A simple feature",
			Assessment: func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
				rnd := ctx.Value("randsrc").(*rand.Rand)
				if rnd.Int() > 100 {
					t.Log("this is a great number")
				}
				return ctx
			},
		},
	}

	test.Test(t, table0, table1.Build().Feature())
}
