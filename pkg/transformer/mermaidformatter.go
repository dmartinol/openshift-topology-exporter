package transformer

import (
	"fmt"
	"os"
	"strings"

	logger "github.com/dmartinol/openshift-topology-exporter/pkg/log"
	"github.com/dmartinol/openshift-topology-exporter/pkg/model"
)

type MermaidFormatter struct {
	diagram strings.Builder
}

func NewMermaidFormatter() *MermaidFormatter {
	formatter := MermaidFormatter{}
	formatter.diagram = strings.Builder{}
	return &formatter
}

func (formatter *MermaidFormatter) Init() {
	formatter.diagram.WriteString("graph TD\n")

	formatter.legend()
}

func (formatter *MermaidFormatter) legend() {
	formatter.diagram.WriteString("subgraph legend\n")
	formatter.diagram.WriteString("\tCompleted\n")
	formatter.diagram.WriteString("\tRunning\n")
	formatter.diagram.WriteString("\tFailed\n")
	formatter.diagram.WriteString(fmt.Sprintf("\tstyle Completed fill: %s\n", model.CompletedColor))
	formatter.diagram.WriteString(fmt.Sprintf("\tstyle Running fill: %s\n", model.RunningColor))
	formatter.diagram.WriteString(fmt.Sprintf("\tstyle Failed fill: %s\n", model.FailedColor))
	formatter.diagram.WriteString("end\n")
}

func (formatter *MermaidFormatter) initNamespace(name string) {
	formatter.diagram.WriteString(fmt.Sprintf("\nsubgraph %s\n", name))
}

func (formatter *MermaidFormatter) AddNamespace(name string, resources []model.Resource, connections []model.Connection) {
	formatter.initNamespace(name)
	for _, resource := range resources {
		formatter.diagram.WriteString(fmt.Sprintf("\t%s(<b>%s</b><br/>%s)\n",
			normalizeId(resource.Id()), resource.Kind(), resource.Label()))

		color, hasStatusColor := resource.StatusColor()
		if hasStatusColor {
			formatter.diagram.WriteString(fmt.Sprintf("\tstyle %s fill:%s\n", normalizeId(resource.Id()), color))
		}
	}

	logger.Debugf("Adding %d connections", len(connections))
	for _, connection := range connections {
		if len(connection.Name) != 0 {
			formatter.diagram.WriteString(fmt.Sprintf("\t%s -- %s --> %s\n", normalizeId(connection.From.Id()),
				connection.Name, normalizeId(connection.To.Id())))
		} else {
			formatter.diagram.WriteString(fmt.Sprintf("\t%s ----> %s\n", normalizeId(connection.From.Id()),
				normalizeId(connection.To.Id())))
		}
	}
	formatter.diagram.WriteString("end")
}

func (formatter *MermaidFormatter) BuildOutput() (string, error) {
	formatter.diagram.WriteString("\n")
	output := formatter.diagram.String()

	file, err := os.Create("diagram.mermaid")
	if err != nil {
		return "", err
	}
	defer file.Close()
	file.WriteString(output)
	return output, nil
}

func normalizeId(id string) string {
	return strings.ReplaceAll(id, " ", "_")
}
