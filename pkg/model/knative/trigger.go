package knative

import (
	"fmt"
	"strings"

	model "github.com/dmartinol/openshift-topology-exporter/pkg/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

type Trigger struct {
	Delegate servingv1.Trigger
}

func (t Trigger) Kind() string {
	return "knative.Trigger"
}
func (t Trigger) Id() string {
	return fmt.Sprintf("trigger %s", t.Delegate.Name)
}
func (t Trigger) Name() string {
	return t.Delegate.Name
}
func (t Trigger) Label() string {
	return fmt.Sprintf("trigger %s", t.Delegate.Name)
}
func (t Trigger) Icon() string {
	return "images/generic.png"
}
func (t Trigger) StatusColor() (string, bool) {
	return "", false
}
func (t Trigger) OwnerReferences() []metav1.OwnerReference {
	return t.Delegate.OwnerReferences
}
func (t Trigger) IsOwnerOf(owner metav1.OwnerReference) bool {
	return false
}
func (t Trigger) ConnectedKinds() []string {
	return []string{"knative.Service", "knative.Broker"}
}
func (t Trigger) ConnectedResources(kind string, resources []model.Resource) ([]model.Resource, string) {
	connected := make([]model.Resource, 0)
	connectionName := ""
	for _, resource := range resources {

		if strings.Compare(resource.Kind(), "knative.Service") == 0 {
			if strings.Compare(t.Delegate.Spec.Subscriber.Ref.Kind, "Service") == 0 {
				service := resource.(Service)
				serviceName := t.Delegate.Spec.Subscriber.Ref.Name
				if strings.Compare(serviceName, service.Name()) == 0 {
					connected = append(connected, service)
				}
			}
			connectionName = "subscriber"
		} else if strings.Compare(resource.Kind(), "knative.Broker") == 0 {
			broker := resource.(Broker)
			brokerName := t.Delegate.Spec.Broker
			if strings.Compare(brokerName, broker.Name()) == 0 {
				connected = append(connected, broker)
			}
			connectionName = "broker"
		}
	}

	return connected, connectionName
}
