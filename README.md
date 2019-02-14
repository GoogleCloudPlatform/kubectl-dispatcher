# Kubectl Dispatcher

## The Version Skew Problem

A kubectl executable and an API Server it is communicating with might
not be at the same version. Kubernetes dictates that kubectl must
correctly communicate with an API Server that is plus or minus one
version away from the kubectl version. For example, kubectl version 1.11
is guaranteed to correctly communicate with a 1.10, 1.11, and 1.12 
Kubernetes cluster. Outside of this support window, there are no
correctness guarantees.

## The Kubectl Dispatcher Solution

The kubectl dispatcher addresses this version skew problem by executing the
exact kubectl version which matches the API Server version. The kubectl
dispatcher is a wrapper which retrieves the server version from a cluster, 
and delegates to the appropriate kubectl version. For example, if a user 
is configured to talk to their Kubernetes cluster that is version 1.10.3-gke, 
then the dispatcher will execute "kubectl.1.10" (in the "clibin" subdirectory 
relative to the dispatcher) passing the arguments and environment passed to
the dispatcher.

IMPORTANT: Versioned kubectl binaries that are dispatched to, MUST be in
the "clibin" subdirectory relative to the dispatcher directory. Versioned
kubectl binaries MUST follow the naming convention: kubectl.<major>.<minor>.
Example: kubectl.1.12.

NOTE: versioned kubectl filenames must NOT start with "kubectl-", since
that is reserved for plugins. Therefore, we prefix versioned kubectl
filenames with "kubectl.". Example: "kubectl.1.12"

- [Build](#build)
- [Test](#test)
- [Run](#run)

## Build

```bash
$ go build cmd/kubectl/kubectl.go
```
## Test

```bash
$ go test -v ./pkg/...
```

## Run

Assuming the dispatcher has been compiled into this current directory, download
the versioned kubectl binary here. NOTE: This download is assuming the
major/minor version of the configured Kubernetes cluster is 1.10. If not, modify
the downloaded release.

```bash
$ wget https://storage.googleapis.com/kubernetes-release/release/v1.10.11/bin/linux/amd64/kubectl
$ mv kubectl clibin/kubectl-1.10
$ chmod +x clibin/kubectl-1.10
```

Run the kubectl dispatcher. The verbosity is useful for debugging.

```bash
$ ./kubectl -v=5 --alsologtostderr version
```

