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

package conf

import (
	"context"
)

type cfgKey struct{}

// FromContext extracts a *Config value from ctx.
// If *Config is not found, it returns nil.
func FromContext(ctx context.Context) *Config{
	if ctx == nil {
		return nil
	}
	if cfg, ok := ctx.Value(cfgKey{}).(*Config); ok {
		return cfg
	}
	return nil
}

// WithContext returns ctx with a config stored.
func WithContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, cfgKey{}, cfg)
}