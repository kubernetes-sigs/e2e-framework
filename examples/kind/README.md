# Example with KinD Clusters

The examples in this directory shows how to use the framework to set up kind clusters that can be used to test cluster resources.

## Predefined environment functions
> See [source code](./custom_funcs)

This directory provides an example that highlights the kind predefined environment functions, that ship with the framework.
These functions can be used to in `Setup` and `Finish` steps to automatically create and teardown a kind cluster respectively.

## Custom environment functions
> See [source code](./custom_funcs)

This example shows how you can write your own custom environment functions to specify the `Setup` and `Finish` 
steps that are used to create and teardown a kind cluster respectively.