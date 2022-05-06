## Cluster API Provider Kind

This is an experimental implementation of Kind infrastructure provider for Cluster API

It consists of 2 coponents:

- `kind-wrapper-api`: a wrapper for Kind CLI, which provides an API for remote access. It needs to run on the same machine as Kind and it needs to be accessible to the Cluster API management cluster
- `cluster-api-provider-kind`: Kind provider implementation for Cluster API. It defines a `KindCluster` custom resource definition and it needs to be deployed to the Cluster API management cluster. It is also necessary to configure a URL of the wrapper API - it can be configured as an environment variable

### Quick start

These guide provides instructions how to set up the project, push it to a local registry and run it in a cluster on a local machine

#### Prerequisites

- Kind: A working installation of Kind and all its dependencies
- Clusterctl: A working installation of `clusterctl` and all its dependencies
- Go (at least 1.17)

#### Setup

1. Clone the project
2. Navigate to `/path/to/project/kind-wrapper-api`, build and start it. No scripts are available here, so `go build` and `./kind-wrapper-api` will have to do. It will start the API server on `0.0.0.0:8888`. it is possible to change the hostname and port by setting the `API_HOST` and `API_PORT` environment variables.
2. Create a local registry, new Kind cluster and make the registry available to the cluster: `/path/to/project/scripts/create-cluster-with-registry.sh` (taken from Kind)
3. Initialize Cluster API in the cluster: `kubectl config use-context kind-kind` (if necessary) and `clusterapi init`
4. Clone the project and navigate to `/path/to/project/cluster-api-provider-kind`
5. In `config/default` add a patch for the `manager` deployment and provide a host for the Kind Wrapper API as an environment variable called `KIND_API_HOST`. There is a sample patch in `/path/to/project/cluster-api-provider-kind/samples/manager_env_patch.yaml`, so the file can be copied and pasted to `/path/to/project/cluster-api-provider-kind/config/default/`, the value of the `KIND_API_HOST` environment variable in it can be adjusted and a reference to this patch file can be uncommented in `/path/to/project/cluster-api-provider-kind/config/default/kustomization.yaml`
6. Push the project to the local registry and deploy it to the cluster: `IMG=registry/image-name:version make docker-build docker-push deploy` - e.g. `IMG=localhost:5001/capi-provider-kind:0.1.35 make docker-build docker-push deploy`

Once the infrastructure provider is deployed and the API wrapper running, it should be possible to create KindCluster resources.

#### Create a cluster

Run `kubectl apply -f /path/to/project/samples/kind-cluster.yaml

#### Limitations

This guide and the project were only tested on MacOS. It should work on Linux, but it may not work on Windows.
