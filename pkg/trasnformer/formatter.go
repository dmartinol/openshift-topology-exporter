package transformer

import "github.com/dmartinol/openshift-topology-exporter/pkg/model"

type Formatter interface {
	Init()
	AddNamespace(name string, resources []model.Resource, connections []model.Connection)
	BuildOutput() error
}
