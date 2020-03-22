# kdiscover

A cli application used for discovering all EKS clusters that it can and exporting there kube-config's


## ToDo

- [x] allow for templated alias for the cluster name (give access to region, partition, cluster name, cluster arn)
- [x] investigate maybe it is worth parsing the kubeconfig instead of executing another command
- [ ] CleanUp comments. print statements
- [ ] prepare a ci pipeline and maybe a cd one
- [ ] prepare packages for brew to distribute this project (and maybe others)
- [ ] add documentation to the readme
- [ ] refactor modules (move from cmd to internals and add a better structure to internals)
- [ ] add documentation to modules
- [ ] add tests for important parts and expecially some integration tests
