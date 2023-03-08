# Go E2E Test Framework for Kubernetes

- [Go E2E Test Framework for Kubernetes](#go-e2e-test-framework-for-kubernetes)
- [Design Document](#design-document)
- [Examples](#examples)
  - [Leveraging CLI Flags and Adding Custom Flags](#leveraging-cli-flags-and-adding-custom-flags)
  - [Test Filtering](#test-filtering)
  - [Using Custom Decoders](#using-custom-decoders)
  - [Custom Resource Definition](#custom-resource-definition)
  - [Runtime Features](#runtime-features)
  - [Multi Cluster Tests](#multi-cluster-tests)
  - [Third Party Integration](#third-party-integration)

# Design Document

[This document](./design) captures high-level design ideas for the next generation of a Go framework for testing components running on Kubernetes. The framework, referred to as `e2e-framework` provides ways that makes it easy to define tests functions that can programmatically test components running on a cluster.  The two overall goals of the new framework are to allow developers to quickly and easily assemble end-to-end tests and to provide a collection of support packages to help with interacting with the API-server.

# Examples

This section of the Document contains a series of examples that details how to leverage `e2e-framework` for specific use-cases or behaviors that range from interacting with real clusters to filtering the tests by labels. 

## Leveraging CLI Flags and Adding Custom Flags

1. [Leveraging Flags](../examples/flags/)
2. [Adding Custom Flags](../examples/custom_flags/)

## Test Filtering

1. [Filtering Tests by Skipping or Selecting flags](../examples/skip_flags/)

## Using Custom Decoders

1. [Custom Decoder](../examples/decoder/)

## Custom Resource Definition

1. [Integrate CRDs](../examples/crds/) into the test workflow

## Runtime Features

1. [Custom Namespace for Each Test](../examples/every_test_custom_ns/)
2. [Fail Fast Mode](../examples/fail_fast/)
3. [Dry Run Mode](../examples/dry_run/)
4. [Custom ENV Functions](../examples/custom_env_funcs/)
5. [Using Existing Kind Cluster](../examples/kind/kind_with_config/)
6. [Pod Command Execution](../examples/pod_exec/)
7. [Parallel Test Run](../examples/parallel_features/)
8. [Test Tables](../examples/table/)
9. [Resource Watch](../examples/watch_resources/)

## Multi Cluster Tests

1. [Multi Cluster Test Run](../examples/multi_cluster/)
2. [Real Cluster Integration](../examples/real_cluster/)

## Third Party Integration

1. [Helm](../examples/third_party_integration/helm/)
2. [Sonobuoy](../examples/sonobuoy/)

