# FAQ

### How does it find clusters ?

It searches for Cluster credentials and trying to list all the clusters on all regions.
For example in AWS case will look for the normal credential chain then try to list and describe
all clusters on all regions (from a given partition, see `kdiscover aws --help` expecially `--aws-partitions`)

### What is the heuristic for `exported locally`

The logic is implemented [here](./internal/kubeconfig/kubeconfig.go) in `IsExported` function.
The basic idea is:

 - we have a cluster in kubeconfig with the same endpoint
 - that cluster is referenced in a `context` block (see [Organizing Cluster Access Using kubeconfig Files][kubeconfig-context])


### Configur context name with `--context-name-alias`

The [kubeconfig context][kubeconfig-context] is used to identify a cluster and user pair, with the `--context-name-alias` you can provide a go template that will be
used to generate the name of the context. In the template you have access to the [Cluster struct](https://github.com/mateimicu/kdiscover/blob/master/internal/cluster/cluster.go#L23). The default template is `{{.Name}}`.


[kubeconfig-context]: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/#context
