# openshift-topology-exporter
Simple tool to export the topology configuration and visualize it in GraphViz

Configure the target namespaces in [](./config.yaml), then run as:

```shell
go run exporter.go
```

Install [Graphviz](https://graphviz.org/) and visualize it as:
```shell
dot -Tpng diagram.dot > diagram.png
```

The resulting diagram is in the generated `diagram.png` image file.
