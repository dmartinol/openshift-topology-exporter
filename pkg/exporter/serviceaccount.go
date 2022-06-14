package exporter

import (
	"fmt"
	"strings"

	authv1T "github.com/openshift/api/authorization/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceAccount struct {
	Delegate v1.ServiceAccount
}

func (s ServiceAccount) Kind() string {
	return "ServiceAccount"
}
func (s ServiceAccount) Id() string {
	return fmt.Sprintf("sa %s", s.Delegate.Name)
}
func (s ServiceAccount) Name() string {
	return s.Delegate.Name
}
func (s ServiceAccount) Label() string {
	return fmt.Sprintf("system:serviceaccount:%s:%s", s.Delegate.Namespace, s.Delegate.Name)
}
func (s ServiceAccount) Icon() string {
	return "images/sa.png"
}
func (s ServiceAccount) StatusColor() (string, bool) {
	return "", false
}
func (s ServiceAccount) OwnerReferences() []metav1.OwnerReference {
	return s.Delegate.OwnerReferences
}
func (s ServiceAccount) IsOwnerOf(owner metav1.OwnerReference) bool {
	return false
}
func (s ServiceAccount) ConnectedKinds() []string {
	return []string{"RoleBinging", "ClusterRoleBinding"}
}
func (s ServiceAccount) ConnectTo(kind string, resources []Resource) string {
	diagram := strings.Builder{}

	// for _, resource := range resources {
	// 	pod := resource.(Pod)
	// 	if s.matchSelector(pod) {
	// 		diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", s.Id(), pod.Id()))
	// 	}
	// }

	return diagram.String()
}

func (s ServiceAccount) TheRoleBindings(roleBindings *authv1T.RoleBindingList) []authv1T.RoleBinding {
	var saRoleBindings []authv1T.RoleBinding
	userName := s.Label()
	for _, roleBinding := range roleBindings.Items {
		for _, rbUserName := range roleBinding.UserNames {
			if strings.Compare(rbUserName, userName) == 0 {
				saRoleBindings = append(saRoleBindings, roleBinding)
			}
		}
	}
	return saRoleBindings
}
func (s ServiceAccount) TheClusterRoleBindings(clusterRoleBindings *authv1T.ClusterRoleBindingList) []authv1T.ClusterRoleBinding {
	var saClusterRoleBindings []authv1T.ClusterRoleBinding
	userName := s.Label()
	for _, clusterRoleBinding := range clusterRoleBindings.Items {
		for _, rbUserName := range clusterRoleBinding.UserNames {
			if strings.Compare(rbUserName, userName) == 0 && strings.Compare(s.Delegate.Namespace, clusterRoleBinding.Namespace) == 0 {
				saClusterRoleBindings = append(saClusterRoleBindings, clusterRoleBinding)
			}
		}
	}
	return saClusterRoleBindings
}
