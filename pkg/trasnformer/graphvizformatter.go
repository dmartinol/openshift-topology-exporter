package transformer

import (
	"fmt"
	"os"
	"strings"

	logger "github.com/dmartinol/openshift-topology-exporter/pkg/log"
	"github.com/dmartinol/openshift-topology-exporter/pkg/model"
)

type GraphVizFormatter struct {
	diagram      strings.Builder
	clusterCount int
}

func NewGraphVizFormatter() *GraphVizFormatter {
	formatter := GraphVizFormatter{clusterCount: 0}
	formatter.diagram = strings.Builder{}
	return &formatter
}

func (formatter *GraphVizFormatter) Init() {
	formatter.diagram.WriteString("digraph G {\n")
	formatter.diagram.WriteString("node [shape=plaintext];\n")

	formatter.legend()
}

func (formatter *GraphVizFormatter) legend() {
	formatter.diagram.WriteString("subgraph legend {\n")
	formatter.diagram.WriteString("legend [\n")
	formatter.diagram.WriteString("label=<<TABLE border=\"0\" cellspacing=\"2\" cellpadding=\"0\">\n")
	formatter.diagram.WriteString(fmt.Sprintf("<TR><TD border=\"0\" bgcolor=\"%s\">Completed</TD></TR>\n", model.CompletedColor))
	formatter.diagram.WriteString(fmt.Sprintf("<TR><TD border=\"0\" bgcolor=\"%s\">Running</TD></TR>\n", model.RunningColor))
	formatter.diagram.WriteString(fmt.Sprintf("<TR><TD border=\"0\" bgcolor=\"%s\">Failed</TD></TR>\n", model.FailedColor))
	formatter.diagram.WriteString("<TR><TD>Legend</TD></TR>\n")
	formatter.diagram.WriteString("</TABLE>>];\n")
	formatter.diagram.WriteString("}\n")
}

func (formatter *GraphVizFormatter) initNamespace(name string) {
	formatter.diagram.WriteString(fmt.Sprintf("\nsubgraph cluster_%d {\n", formatter.clusterCount))
	formatter.diagram.WriteString("style=filled;\n")
	formatter.diagram.WriteString("color=lightgrey;\n")
	formatter.diagram.WriteString("node [style=filled,color=white];\n")
	formatter.diagram.WriteString(fmt.Sprintf("label =\"%s\";\n", name))
	formatter.clusterCount++
}

func (formatter *GraphVizFormatter) AddNamespace(name string, resources []model.Resource, connections []model.Connection) {
	formatter.initNamespace(name)
	for _, resource := range resources {
		color, hasStatusColor := resource.StatusColor()
		if hasStatusColor {
			formatter.diagram.WriteString(fmt.Sprintf("\"%s\" [ label=\"%s\", image=\"%s\", labelloc=b, color=\"%s\" ];\n",
				resource.Id(), resource.Label(), resource.Icon(), color))
		} else {
			formatter.diagram.WriteString(fmt.Sprintf("\"%s\" [ label=\"%s\", image=\"%s\", labelloc=b ];\n",
				resource.Id(), resource.Label(), resource.Icon()))
		}
	}

	logger.Debugf("Adding %d connections", len(connections))
	for _, connection := range connections {
		options := ""
		if len(connection.Name) != 0 {
			options = fmt.Sprintf(" [label=\"%s\"]", connection.Name)
		}
		formatter.diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"%s\n", connection.From.Id(), connection.To.Id(), options))
	}
	formatter.diagram.WriteString("\n}")
}

func (formatter *GraphVizFormatter) BuildOutput() error {
	formatter.diagram.WriteString("\n}")

	file, err := os.Create("diagram.dot")
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(formatter.diagram.String())
	return nil
}
