package exporter

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CustomResource struct {
	Delegate metav1.OwnerReference
}

func (cr CustomResource) Kind() string {
	return cr.Delegate.Kind
}
func (cr CustomResource) Id() string {
	return fmt.Sprintf("%s %s", strings.ToLower(cr.Kind()), cr.Delegate.Name)
}
func (cr CustomResource) Name() string {
	return cr.Delegate.Name
}
func (cr CustomResource) Label() string {
	return cr.Delegate.Name
}
func (cr CustomResource) Icon() string {
	return "images/crd.png"
}
func (cr CustomResource) OwnerReferences() []metav1.OwnerReference {
	return []metav1.OwnerReference{}
}
func (cr CustomResource) IsOwnerOf(owner metav1.OwnerReference) bool {
	return strings.Compare(owner.Kind, cr.Kind()) == 0 && strings.Compare(owner.Name, cr.Name()) == 0
}
func (cr CustomResource) ConnectedKinds() []string {
	return []string{""}
}
func (cr CustomResource) ConnectTo(kind string, resources []Resource) string {
	diagram := strings.Builder{}
	return diagram.String()
}

func (cr CustomResource) StatusColor() (string, bool) {
	return "", false
}
