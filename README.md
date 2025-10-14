# kdiscover
[![CI](https://github.com/mateimicu/kdiscover/actions/workflows/golang-ci.yaml/badge.svg?branch=master)](https://github.com/mateimicu/kdiscover/actions/workflows/golang-ci.yaml)
![CodeQL](https://github.com/mateimicu/kdiscover/workflows/CodeQL/badge.svg)
[![codecov](https://codecov.io/gh/mateimicu/kdiscover/branch/master/graph/badge.svg)](https://codecov.io/gh/mateimicu/kdiscover)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/mateimicu/kdiscover?sort=semver)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mateimicu/kdiscover)
![GitHub](https://img.shields.io/github/license/mateimicu/kdiscover)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmateimicu%2Fkdiscover.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmateimicu%2Fkdiscover?ref=badge_shield)
[![Go Report Card](https://goreportcard.com/badge/github.com/mateimicu/kdiscover)](https://goreportcard.com/report/github.com/mateimicu/kdiscover)


Kdiscover is a simple utility to list and configure access to all clusters it can find.
The basic usecase revolves in having access to a lot of clusters but you still need to discover and export apposite kubeconfig.

Currently we support EKS (AWS) and GKE (Google Cloud) clusters with plans to support other k8s providers (AKS, etc ...).

- [kdiscover](#kdiscover)
  - [Example](#example)
    - [AWS EKS Clusters](#aws-eks-clusters)
    - [Google GKE Clusters](#google-gke-clusters)
  - [Prerequisites](#prerequisites)
  - [Demo](#demo)
  - [Install](#install)
    - [Krew (Recommended)](#krew-recommended)
    - [MacOs](#macos)
    - [Binary](#binary)
    - [Go](#go)
  - [Future Plans](#future-plans)
  - [FAQ](docs/FAQ.md)


<!--## Demo-->
<!--[![asciicast](https://asciinema.org/a/qfxDubtATYtLJ1W1vOK6rBzSE.svg)](https://asciinema.org/a/qfxDubtATYtLJ1W1vOK6rBzSE)-->

## Example

### AWS EKS Clusters

```bash
~ $ kubectl discover aws list
┌────────────────────────────────────────────────────────────────────────────────┐
│     cluster name                  region              status  exported locally │
├────────────────────────────────────────────────────────────────────────────────┤
│  1  production-us                 us-east-1           ACTIVE          No       │
│  2  production-eu                 eu-west-1           ACTIVE          No       │
│  3  dev-eu                        eu-central-1        ACTIVE          No       │
│  4  sandbox-eu                    eu-central-1        ACTIVE          No       │
├────────────────────────────────────────────────────────────────────────────────┤
│                                   number of clusters  4                        │
└────────────────────────────────────────────────────────────────────────────────┘
~ $ kubectl discover aws update
Update all EKS Clusters
Found 4 clusters remote
Backup kubeconfig to /Users/tuxy/.kube/config.bak
```

### Google GKE Clusters

```bash
~ $ kubectl discover gke list
┌────────────────────────────────────────────────────────────────────────────────┐
│     cluster name                  region              status  exported locally │
├────────────────────────────────────────────────────────────────────────────────┤
│  1  production-gke                us-central1-a       RUNNING         No       │
│  2  staging-gke                   us-central1-b       RUNNING         No       │
│  3  dev-gke                       europe-west1-a      RUNNING         No       │
├────────────────────────────────────────────────────────────────────────────────┤
│                                   number of clusters  3                        │
└────────────────────────────────────────────────────────────────────────────────┘
~ $ kubectl discover gke update --gcp-projects my-project --gcp-zones us-central1-a,europe-west1-a
Update all GKE Clusters
Found 2 clusters remote
Backup kubeconfig to /Users/tuxy/.kube/config.bak
```
~ $ kubectl discover aws list
┌────────────────────────────────────────────────────────────────────────────────┐
│     cluster name                  region              status  exported locally │
├────────────────────────────────────────────────────────────────────────────────┤
│  1  production-us                 us-east-1           ACTIVE         Yes       │
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

## Prerequisites

### For AWS EKS clusters
- AWS CLI configured with appropriate credentials
- IAM permissions to list and describe EKS clusters

### For Google GKE clusters
- `gcloud` CLI installed and authenticated
- `gke-gcloud-auth-plugin` installed for cluster authentication:
  ```bash
  gcloud components install gke-gcloud-auth-plugin
  ```
- GCP credentials configured (via `gcloud auth login` or service account)
- IAM permissions to list GKE clusters and access Container API
- Set environment variable: `export USE_GKE_GCLOUD_AUTH_PLUGIN=True`


## Install

## Krew (Recommended)

Using the [Krew](https://krew.sigs.k8s.io/) plugin manager:

```bash
kubectl krew install discover
```
Note that in this context the command will need to invoked using `kubectl discover`.

## macOs

You can also install a binary release on macOS using [brew](https://brew.sh/):

```bash
brew install mateimicu/tap/kdiscover
brew upgrade mateimicu/tap/kdiscover
```
Note that in this context the executable is name `kdiscover`.

## Binary

You can download a specific version from the [release page](https://github.com/mateimicu/kdiscover/releases)

## Go

```bash
GO111MODULE=on go get github.com/mateimicu/kdiscover
```

# Future Plans


Development is tracked in [this board](https://github.com/mateimicu/kdiscover/projects/1) and we also have specific [milestones](https://github.com/mateimicu/kdiscover/milestones?direction=asc&sort=due_date)

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmateimicu%2Fkdiscover.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmateimicu%2Fkdiscover?ref=badge_large)
