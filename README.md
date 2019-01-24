# Kubectl Dispatcher

The kubectl dispatcher is a wrapper which retrieves the server version from
a cluster, and executes the appropriate kubectl version. For example, if a
user is configured to talk to their Kubernetes cluster that is version
v1.10.3-gke.1, then this binary will execute "kubectl-1.10" (in the same
directory as this dispatcher) passing the arguments and environment of
this binary.

IMPORTANT: Versioned kubectl binaries that are dispatched to, MUST be in
the same directory as this dispatcher binary. Versioned kubectl binaries
MUST follow the naming convention: kubectl-<major>.<minor>. Example:
kubectl-1.12.

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
$ mv kubectl.1 kubectl-1.10
$ chmod +x kubectl-1.10
```

Run the kubectl dispatcher. The verbosity is useful for debugging.

```bash
$ ./kubectl -v=5 --alsologtostderr version
```

