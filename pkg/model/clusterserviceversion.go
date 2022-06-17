package model

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterServiceVersion struct {
	Delegate metav1.OwnerReference
}

func (csv ClusterServiceVersion) Kind() string {
	return csv.Delegate.Kind
}
func (csv ClusterServiceVersion) Id() string {
	return csv.Name()
}
func (csv ClusterServiceVersion) Name() string {
	return csv.Delegate.Name
}
func (csv ClusterServiceVersion) Label() string {
	return csv.Delegate.Name
}
func (csv ClusterServiceVersion) Icon() string {
	return "images/operator.png"
}
func (csv ClusterServiceVersion) OwnerReferences() []metav1.OwnerReference {
	return []metav1.OwnerReference{}
}
func (csv ClusterServiceVersion) IsOwnerOf(owner metav1.OwnerReference) bool {
	return strings.Compare(owner.Kind, csv.Kind()) == 0 && strings.Compare(owner.Name, csv.Name()) == 0
}
func (csv ClusterServiceVersion) ConnectedKinds() []string {
	return []string{""}
}
func (csv ClusterServiceVersion) ConnectedResources(kind string, resources []Resource) ([]Resource, string) {
	return []Resource{}, ""
}

func (csv ClusterServiceVersion) StatusColor() (string, bool) {
	return "", false
}
