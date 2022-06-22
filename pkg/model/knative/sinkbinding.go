package knative

import (
	"fmt"
	"strings"

	model "github.com/dmartinol/openshift-topology-exporter/pkg/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"
)

type SinkBinding struct {
	Delegate sourcesv1.SinkBinding
}

func (s SinkBinding) Kind() string {
	return "knative.SinkBinding"
}
func (s SinkBinding) Id() string {
	return fmt.Sprintf("sinkbinding %s", s.Delegate.Name)
}
func (s SinkBinding) Name() string {
	return s.Delegate.Name
}
func (s SinkBinding) Label() string {
	return fmt.Sprintf("sinkbinding %s", s.Delegate.Name)
}
func (s SinkBinding) Icon() string {
	return "images/generic.png"
}
func (s SinkBinding) StatusColor() (string, bool) {
	return "", false
}
func (s SinkBinding) OwnerReferences() []metav1.OwnerReference {
	return s.Delegate.OwnerReferences
}
func (s SinkBinding) IsOwnerOf(owner metav1.OwnerReference) bool {
	return false
}
func (s SinkBinding) ConnectedKinds() []string {
	return []string{"knative.Service", "knative.Broker"}
}
func (s SinkBinding) ConnectedResources(kind string, resources []model.Resource) ([]model.Resource, string) {
	connected := make([]model.Resource, 0)
	connectionName := ""
	for _, resource := range resources {

		if strings.Compare(resource.Kind(), "knative.Service") == 0 {
			if strings.Compare(s.Delegate.Spec.BindingSpec.Subject.Kind, "Service") == 0 {
				service := resource.(Service)
				serviceName := s.Delegate.Spec.BindingSpec.Subject.Name
				if strings.Compare(serviceName, service.Name()) == 0 {
					connected = append(connected, service)
				}
			}
			connectionName = "subject"
		} else if strings.Compare(resource.Kind(), "knative.Broker") == 0 {
			if strings.Compare(s.Delegate.Spec.Sink.Ref.Kind, "Broker") == 0 {
				broker := resource.(Broker)
				brokerName := s.Delegate.Spec.Sink.Ref.Name
				if strings.Compare(brokerName, broker.Name()) == 0 {
					connected = append(connected, broker)
				}
			}
			connectionName = "sink"
		}
	}

	return connected, connectionName
}
