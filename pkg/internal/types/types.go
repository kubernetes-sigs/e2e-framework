package types

import (
	"context"
	"testing"
)

// Config configures and exposes an test environment
type Config interface {
	Env() (Environment, error)
}

// EnvFunc represents a user-defined operation that
// can be used to customized the behavior of the
// environment.
type EnvFunc func(context.Context, Config)

// Environment represents an environment where
// features can be tested.
type Environment interface {
	Config() Config

	// Setup registers environment operations that are executed once
	// prior to the environment being ready and prior to any test.
	Setup(context.Context, ...EnvFunc) error

	// BeforeTest registers funcs that are executed before each Env.Test(...)
	BeforeTest(context.Context, *testing.T, ...EnvFunc)

	//Test executes a test feature
	Test(context.Context, *testing.T, Feature)

	// AfterTest registers funcs that are executed after each Env.Test(...)
	AfterTest(context.Context, *testing.T, ...EnvFunc)

	// Finish registers funcs that are executed at the end.
	Finish(context.Context, ...EnvFunc) error

	// Launches the test suite from within a TestMain
	Run(*testing.M) int
}

type Feature interface {

}