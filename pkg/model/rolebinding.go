package model

import (
	"fmt"

	authv1T "github.com/openshift/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RoleBinding struct {
	Delegate authv1T.RoleBinding
}

func (r RoleBinding) Kind() string {
	return "RoleBinding"
}
func (r RoleBinding) Id() string {
	return fmt.Sprintf("rb %s", r.Delegate.Name)
}
func (r RoleBinding) Name() string {
	return r.Delegate.Name
}
func (r RoleBinding) Label() string {
	return r.Name()
}
func (r RoleBinding) Icon() string {
	return "images/role.png"
}
func (r RoleBinding) StatusColor() (string, bool) {
	return "", false
}
func (r RoleBinding) OwnerReferences() []metav1.OwnerReference {
	return r.Delegate.OwnerReferences
}
func (r RoleBinding) IsOwnerOf(owner metav1.OwnerReference) bool {
	return false
}
func (r RoleBinding) ConnectedKinds() []string {
	return []string{}
}
func (r RoleBinding) ConnectedResources(kind string, resources []Resource) ([]Resource, string) {
	return []Resource{}, ""
}
