package transformer

import (
	"github.com/dmartinol/openshift-topology-exporter/pkg/model"
)

type Transformer struct {
	formatter Formatter
}

func NewTransformer(formatter Formatter) *Transformer {
	return &Transformer{formatter: formatter}
}

func (transformer Transformer) Transform(topologyModel model.TopologyModel) (string, error) {
	transformer.formatter.Init()
	for _, namespace := range topologyModel.AllNamespaces() {
		transformer.formatter.AddNamespace(namespace.Name(), namespace.AllResources(), namespace.AllConnections())
	}
	return transformer.formatter.BuildOutput()
}
