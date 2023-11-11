# Kubernetes resource cache

This project is uses an inmemory cache with kubernetes dynamic watchers to cache custom resources instead of going to api for every request. Resources are cache based on namespace and a cache enabled label.

# Build

Build the kubectl-resourcecache plugin and install it in your GOPATH:

```
make cli
```

Test out with custom resource

```
make test
```

# Usage

**Note:** This example uses [cert-manager](https://cert-manager.io/docs/installation/), make sure you have the CRDs installed in your cluster.

1. Install the kubectl plugin.
2. Apply test data
  ```sh
  kubectl apply -f testdata/test-certissuer.yaml
  ```
3. Use `get` command in the plugin and pass in the group, version, kind and namespace of the resource
   ```sh
   kubectl resourcecache get 
   ```

   Output: 
   ```
   Group: cert-manager.io
   Version: v1
   Kind: issuers
   Namespace: default
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  101584μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  256μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  211μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  297μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  188μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  188μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  216μs
   12 resources found of type: cert-manager.io/v1, Resource=issuers
   Time taken:  380μs
   ...
   ...
   ```