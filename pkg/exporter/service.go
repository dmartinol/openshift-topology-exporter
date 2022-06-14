package exporter

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Service struct {
	Delegate v1.Service
}

func (s Service) Kind() string {
	return "Service"
}
func (s Service) Id() string {
	return fmt.Sprintf("svc %s", s.Delegate.Name)
}
func (s Service) Name() string {
	return s.Delegate.Name
}
func (s Service) Label() string {
	return s.Delegate.Name
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
	return false
}
func (s Service) ConnectedKinds() []string {
	return []string{"Pod"}
}
func (s Service) ConnectTo(kind string, resources []Resource) string {
	diagram := strings.Builder{}

	for _, resource := range resources {
		pod := resource.(Pod)
		if s.matchSelector(pod) {
			diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", s.Id(), pod.Id()))
		}
	}

	return diagram.String()
}
func (s Service) matchSelector(pod Pod) bool {
	for label, value := range s.Delegate.Spec.Selector {
		found := false
		for podLabel, podValue := range pod.Delegate.ObjectMeta.Labels {
			if strings.Compare(podLabel, label) == 0 {
				if strings.Compare(podValue, value) != 0 {
					return false
				}
				found = true
			}
		}
		if !found {
			return false
		}
	}

	return len(s.Delegate.Spec.Selector) > 0
}
