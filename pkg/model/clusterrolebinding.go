package model

import (
	"fmt"

	authv1T "github.com/openshift/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterRoleBinding struct {
	Delegate authv1T.ClusterRoleBinding
}

func (c ClusterRoleBinding) Kind() string {
	return "ClusterRoleBinding"
}
func (c ClusterRoleBinding) Id() string {
	return fmt.Sprintf("crb %s", c.Delegate.Name)
}
func (c ClusterRoleBinding) Name() string {
	return c.Delegate.Name
}
func (c ClusterRoleBinding) Label() string {
	return fmt.Sprintf("crb %s", c.Delegate.Name)
}
func (c ClusterRoleBinding) Icon() string {
	return "images/role.png"
}
func (c ClusterRoleBinding) StatusColor() (string, bool) {
	return "", false
}
func (c ClusterRoleBinding) OwnerReferences() []metav1.OwnerReference {
	return c.Delegate.OwnerReferences
}
func (c ClusterRoleBinding) IsOwnerOf(owner metav1.OwnerReference) bool {
	return false
}
func (c ClusterRoleBinding) ConnectedKinds() []string {
	return []string{}
}
func (c ClusterRoleBinding) ConnectedResources(kind string, resources []Resource) ([]Resource, string) {
	return []Resource{}, ""
}
