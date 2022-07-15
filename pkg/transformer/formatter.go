package transformer

import (
	"github.com/dmartinol/openshift-topology-exporter/pkg/config"
	"github.com/dmartinol/openshift-topology-exporter/pkg/model"
)

type Formatter interface {
	Init()
	AddNamespace(name string, resources []model.Resource, connections []model.Connection)
	BuildOutput() (string, error)
}

func NewFormatterForConfig(config config.ExporterConfig) Formatter {
	if config.FormatterClass == "graphviz" {
		return NewGraphVizFormatter()
	} else if config.FormatterClass == "mermaid" {
		return NewMermaidFormatter()
	}
	return NewGraphVizFormatter()
}
