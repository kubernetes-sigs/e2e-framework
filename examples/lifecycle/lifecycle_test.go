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

package lifecycle

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var test = env.New()

// TestMain defined package-wide (or test suite) configuration.
func TestMain(m *testing.M){
	// define a setup function
	test.Setup(func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		fileName := filepath.Join(os.TempDir(), "randfile.txt")
		return context.WithValue(ctx, "randfile", fileName), nil
	})

	// BeforeEachFeature specifies behavior that occurs before each feature is tested.
	// Write a random number in the file.
	test.BeforeEachFeature(func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		fileName := ctx.Value("randfile").(string) // in real world use, check type assertion
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		// save random number in file
		err := ioutil.WriteFile(fileName, []byte(fmt.Sprintf("%d", rnd.Intn(255))), 0644)
		return ctx, err
	})

	// AfterEachFeature specifies behavior that occurs after each feature is tested.
	// Clear the file prior to next text (this step is illustrative)
	test.AfterEachFeature(func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		fileName := ctx.Value("randfile").(string) // in real world use, check type assertion
		err := ioutil.WriteFile(fileName, []byte(""), 0644)
		return ctx, err
	})

	// Finish tears down resources used in the test suite
	test.Finish(func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		fileName := ctx.Value("randfile").(string) // in real world use, check type assertion
		return ctx, os.RemoveAll(fileName)
	})

	// Don't forget to run the package test
	os.Exit(test.Run(m))
}

// TestGuessNumberFeature shows the use of before and after
func TestGuessNumberFeature(t *testing.T) {
	f := features.New("guesses")
	// Setup defines setup behavior executed prior to running assessment
	// This can be used to prepare data for the test.
	f.Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		fileName := ctx.Value("randfile").(string) // in real world use, check type assertion
		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			t.Fatal(err)
		}
		val, err := strconv.Atoi(string(data))
		if err != nil {
			t.Fatalf("unable to read file: %s", err)
		}
		return context.WithValue(ctx, "number", val)
	})

	// Assess define a test behavior for the feature
	f.Assess("gess-high", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		number := ctx.Value("number").(int) // in real world use, check type assertion
		if number >= 200 {
			t.Error("Guessed too high")
		}
		return ctx
	})

	// Assess define at test behavior
	f.Assess("gess-low", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		number := ctx.Value("number").(int) // in real world use, check type assertion
		if number <= 50 {
			t.Error("Guessed too high")
		}
		return ctx
	})

	// Teardown defines behavior to clean up after the feature is tested.
	f.Teardown(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		return context.WithValue(ctx, "number", -1)
	})

	// Test the feature
	test.Test(t, f.Feature())
}