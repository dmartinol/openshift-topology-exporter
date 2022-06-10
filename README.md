# openshift-topology-exporter
Simple tool to export the topology configuration and visualize it in GraphViz.

Resources exported and connected in the diagrams are:
* [Namespace [core/v1]](https://docs.openshift.com/online/pro/rest_api/core/namespace-core-v1.html)
* [Route [route.openshift.io/v1]](https://docs.openshift.com/online/pro/rest_api/route_openshift_io/route-route-openshift-io-v1.html)
* [Service [core/v1]](https://docs.openshift.com/online/pro/rest_api/core/service-core-v1.html)
* [Deployment [apps/v1]](https://docs.openshift.com/online/pro/rest_api/apps/deployment-apps-v1.html)
* [DeploymentConfig [apps.openshift.io/v1]](https://docs.openshift.com/online/pro/rest_api/apps_openshift_io/deploymentconfig-apps-openshift-io-v1.html)
* [Pod [core/v1]](https://docs.openshift.com/online/pro/rest_api/core/pod-core-v1.html)
* [ServiceAccount [core/v1]](https://docs.openshift.com/online/pro/rest_api/core/serviceaccount-core-v1.html)
* [RoleBinding [rbac.authorization.k8s.io/v1]](https://docs.openshift.com/online/pro/rest_api/rbac_authorization_k8s_io/rolebinding-rbac-authorization-k8s-io-v1.html)

This tool is based on the [OpenShift Client in Go](https://github.com/openshift/client-go) and requires [Golang](https://go.dev/).

## Instructions
> **Note**: You must be logged in to the OpenShift console to successfully run the tool

Configure the target namespaces in [config.yaml](./config.yaml), then run as:
```shell
go run exporter.go
```
Install [Graphviz](https://graphviz.org/) and visualize it as:
```shell
dot -Tpng diagram.dot > diagram.png
```

The resulting diagram is in the generated `diagram.png` image file.

An example diagram is given, captured from a real deployment of the [Red Hat Single Sign-On](https://access.redhat.com/products/red-hat-single-sign-on):
![](./examples/rhsso.png)

In alternative you can paste the content of the generated `diagram.dot` file in an online visualizer like [https://dreampuf.github.io/GraphvizOnline](https://dreampuf.github.io/GraphvizOnline/)