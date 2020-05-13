# kdiscover
![Lint & build Golang project](https://github.com/mateimicu/kdiscover/workflows/Lint%20&%20build%20Golang%20project/badge.svg?branch=master)
[![codecov](https://codecov.io/gh/mateimicu/kdiscover/branch/master/graph/badge.svg)](https://codecov.io/gh/mateimicu/kdiscover)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/mateimicu/kdiscover?sort=semver)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mateimicu/kdiscover)
![GitHub](https://img.shields.io/github/license/mateimicu/kdiscover)


Kdiscover is a simple utility to list and configure access to all clusters it can find.
The basic usecase revolves in havin access to a lot of clusters but you still need to discover and export apposite kubeconfig.

Currently we suport only EKS clusters but there are plans to support othe k8s providers (GKE, AKE, etc ...)

- [kdiscover](#kdiscover)
  - [Demo](#demo)
  - [Install](#install)
    - [Binary](#binary)
    - [MacOs](#macos)
    - [Go](#go)
  - [Future Plans](#future-plans)
  - [FAQ](#faq)


## Demo

```bash
~ $ kdiscover aws list
┌────────────────────────────────────────────────────────────────────────────────┐
│     cluster name                  region              status  exported locally │
├────────────────────────────────────────────────────────────────────────────────┤
│  1  production-usa                us-east-1           ACTIVE          No       │
│  2  production-eu                 eu-west-1           ACTIVE          No       │
│  3  dev-eu                        eu-central-1        ACTIVE          No       │
│  4  sandbox-eu                    eu-central-1        ACTIVE          No       │
├────────────────────────────────────────────────────────────────────────────────┤
│                                   number of clusters  4                        │
└────────────────────────────────────────────────────────────────────────────────┘
~ $ kdiscover aws update
Update all EKS Clusters
Found 4 clusters remote
Backup kubeconfig to /Users/tuxy/.kube/config.bak
~ $ kdiscover aws list
┌────────────────────────────────────────────────────────────────────────────────┐
│     cluster name                  region              status  exported locally │
├────────────────────────────────────────────────────────────────────────────────┤
│  1  production-usa                us-east-1           ACTIVE         Yes       │
│  2  production-eu                 eu-west-1           ACTIVE         Yes       │
│  3  dev-eu                        eu-central-1        ACTIVE         Yes       │
│  4  sandbox-eu                    eu-central-1        ACTIVE         Yes       │
├────────────────────────────────────────────────────────────────────────────────┤
│                                   number of clusters  4                        │
└────────────────────────────────────────────────────────────────────────────────┘
```


Columns in the list :

- `cluster name` is the name of the cluster based on the configuration
- `region` region where the cluster is deployed (it is cloud specific)
- `status` this is reported by the cloud, if the cluster is up or in another state (modifying, down, creating ... etc)
- `exported locally` uses an heuristic too see if the local config already has information about this cluster


## Install

### Binary

You can download a specific version from the [release page](https://github.com/golangci/golangci-lint/releases)

### macOs

You can also install a binary release on macOS using [brew](https://brew.sh/):

```bash
brew install mateimicu/tap/kdiscover
brew upgrade mateimicu/tap/kdiscover
```

### Go

```bash
GO111MODULE=on go get github.com/mateimicu/kdiscover
```

## Future Plans


Development is tracked in [this board](https://github.com/mateimicu/kdiscover/projects/1) and we also have specific [milestones](https://github.com/mateimicu/kdiscover/milestones?direction=asc&sort=due_date)


## FAQ

### How does it find clusters ?

It searches for Cluster credentials and trying to list all the clusters on all regions.
For example in AWS case will look for the normal credential chain then try to list and describe
all clusters on all regions (from a given partition, see `kdiscover aws --help` expecially `--aws-partitions`)

### What is the heuristic for `exported locally`

The logic is implemented [here](./internal/kubeconfig/kubeconfig.go) in `IsExported` function.
The basic idea is:

 - we have a cluster in kubeconfig with the same endpoint
 - that cluster is referenced in a `context` block (see [Organizing Cluster Access Using kubeconfig Files](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/#context))
