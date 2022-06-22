package knative

import (
	"fmt"

	model "github.com/dmartinol/openshift-topology-exporter/pkg/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

type Broker struct {
	Delegate servingv1.Broker
}

func (b Broker) Kind() string {
	return "knative.Broker"
}
func (b Broker) Id() string {
	return fmt.Sprintf("broker %s", b.Delegate.Name)
}
func (b Broker) Name() string {
	return b.Delegate.Name
}
func (b Broker) Label() string {
	return fmt.Sprintf("broker %s", b.Delegate.Name)
}
func (b Broker) Icon() string {
	return "images/generic.png"
}
func (b Broker) StatusColor() (string, bool) {
	return "", false
}
func (b Broker) OwnerReferences() []metav1.OwnerReference {
	return b.Delegate.OwnerReferences
}
func (b Broker) IsOwnerOf(owner metav1.OwnerReference) bool {
	return false
}
func (b Broker) ConnectedKinds() []string {
	return []string{}
}
func (b Broker) ConnectedResources(kind string, resources []model.Resource) ([]model.Resource, string) {
	return nil, ""
}
