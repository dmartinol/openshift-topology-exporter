package exporter

import (
	"fmt"
	"strings"

	routev1T "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Route struct {
	Delegate routev1T.Route
}

func (r Route) Kind() string {
	return "Route"
}
func (r Route) Id() string {
	return fmt.Sprintf("route %s", r.Delegate.Name)
}
func (r Route) Name() string {
	return r.Delegate.Name
}
func (r Route) Label() string {
	return r.Delegate.Name
}
func (r Route) Icon() string {
	return "images/ingress.png"
}
func (r Route) StatusColor() (string, bool) {
	return "", false
}
func (r Route) OwnerReferences() []metav1.OwnerReference {
	return r.Delegate.OwnerReferences
}
func (r Route) IsOwnerOf(owner metav1.OwnerReference) bool {
	return false
}
func (r Route) ConnectedKinds() []string {
	return []string{"Service"}
}
func (r Route) ConnectTo(kind string, resources []Resource) string {
	diagram := strings.Builder{}

	if strings.Compare(r.Delegate.Spec.To.Kind, "Service") == 0 {
		for _, resource := range resources {
			service := resource.(Service)
			serviceName := r.Delegate.Spec.To.Name
			if strings.Compare(serviceName, service.Name()) == 0 {
				diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\" [label=\"exposes\"];\n", r.Id(), service.Id()))
			}
		}
	}

	return diagram.String()
}
