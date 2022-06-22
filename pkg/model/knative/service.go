package knative

import (
	"fmt"
	"strings"

	model "github.com/dmartinol/openshift-topology-exporter/pkg/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

type Service struct {
	Delegate servingv1.Service
}

func (s Service) Kind() string {
	return "knative.Service"
}
func (s Service) Id() string {
	return fmt.Sprintf("ksvc %s", s.Delegate.Name)
}
func (s Service) Name() string {
	return s.Delegate.Name
}
func (s Service) Label() string {
	return fmt.Sprintf("ksvc %s", s.Delegate.Name)
}
func (s Service) Icon() string {
	return "images/svc.png"
}
func (s Service) StatusColor() (string, bool) {
	return "", false
}
func (s Service) OwnerReferences() []metav1.OwnerReference {
	return s.Delegate.OwnerReferences
}
func (s Service) IsOwnerOf(owner metav1.OwnerReference) bool {
	switch owner.Kind {
	case "Revision":
		return strings.HasPrefix(owner.Name, s.Name())
	}
	return false
}
func (s Service) ConnectedKinds() []string {
	return []string{}
}
func (s Service) ConnectedResources(kind string, resources []model.Resource) ([]model.Resource, string) {
	return []model.Resource{}, ""
}
