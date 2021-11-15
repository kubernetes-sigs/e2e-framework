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

package wait

import (
	apimachinerywait "k8s.io/apimachinery/pkg/util/wait"
	"time"
)

const (
	defaultPollTimeout  = 5 * time.Minute
	defaultPollInterval = 5 * time.Second
)

type Options struct {
	// Interval is used to specify the poll interval while waiting for a condition to be met
	Interval time.Duration
	// Timeout is used to indicate the total time to be spent in polling for the condition
	// to be met.
	Timeout time.Duration
	// StopChan is used to setup a wait mechanism using the apimachinerywait.PollUntil method
	StopChan chan struct{}
	// Immediate is used to indicate if the apimachinerywait's immediate wait method are to be
	// called instead of the regular one
	Immediate bool
}

type Option func(*Options)

func WithTimeout(timeout time.Duration) Option {
	return func(options *Options) {
		options.Timeout = timeout
	}
}

func WithInterval(interval time.Duration) Option {
	return func(options *Options) {
		options.Interval = interval
	}
}

func WithStopChannel(stopChan chan struct{}) Option {
	return func(options *Options) {
		options.StopChan = stopChan
	}
}

func WithImmediate() Option {
	return func(options *Options) {
		options.Immediate = true
	}
}

func For(conditionFunc apimachinerywait.ConditionFunc, opts ...Option) error {
	options := &Options{
		Interval: defaultPollInterval,
		Timeout: defaultPollTimeout,
		StopChan: nil,
		Immediate: false,
	}

	for _, fn := range opts {
		fn(options)
	}

	// Setting the options.StopChan will force the usage of `PollUntil`
	if options.StopChan != nil {
		if options.Immediate {
			return apimachinerywait.PollImmediateUntil(options.Interval, conditionFunc, options.StopChan)
		}
		return apimachinerywait.PollUntil(options.Interval, conditionFunc, options.StopChan)
	}

	if options.Immediate {
		return apimachinerywait.PollImmediate(options.Interval, options.Timeout, conditionFunc)
	}
	return apimachinerywait.Poll(options.Interval, options.Timeout, conditionFunc)
}
