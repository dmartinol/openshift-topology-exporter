package model

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
func (s Service) ConnectedResources(kind string, resources []Resource) ([]Resource, string) {
	connected := make([]Resource, 0)
	for _, resource := range resources {
		pod := resource.(Pod)
		if s.matchSelector(pod) {
			connected = append(connected, pod)
		}
	}

	return connected, ""
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
