package model

import (
	"reflect"
	"strings"

	logger "github.com/dmartinol/openshift-topology-exporter/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NamespaceModel struct {
	name            string
	resourcesByKind map[string][]Resource
	connections     []Connection
}

func (namespace NamespaceModel) Debug(header string) string {
	logger.Debugf("[%s] Sizes for %s are %d, %d", header, namespace.name, len(namespace.resourcesByKind), len(namespace.connections))
	return namespace.name
}

func (namespace NamespaceModel) Name() string {
	return namespace.name
}

func (namespace NamespaceModel) LookupByKindAndId(kind string, id string) Resource {
	for _, resource := range namespace.resourcesByKind[kind] {
		if strings.Compare(id, resource.Id()) == 0 {
			return resource
		}
	}

	return nil
}

func (namespace NamespaceModel) AddResource(resource Resource) bool {
	if namespace.LookupByKindAndId(resource.Kind(), resource.Id()) == nil {
		logger.Debugf("Adding resource %s of kind %s", resource.Name(), resource.Kind())
		namespace.resourcesByKind[resource.Kind()] = append(namespace.resourcesByKind[resource.Kind()], resource)
		return true
	}
	logger.Debugf("Skipped existing resource %s of kind %s", resource.Name(), resource.Kind())
	return false
}
func (namespace NamespaceModel) LookupOwner(owner metav1.OwnerReference) Resource {
	for _, resources := range namespace.resourcesByKind {
		for _, resource := range resources {
			if resource.IsOwnerOf(owner) {
				return resource
			}
		}
	}

	if strings.Compare(owner.Kind, "ClusterServiceVersion") == 0 {
		csvResource := ClusterServiceVersion{Delegate: owner}
		namespace.AddResource(csvResource)
		return csvResource
	} else {
		// TODO Just guessing ....
		customResource := CustomResource{Delegate: owner}
		namespace.AddResource(customResource)
		return customResource
	}
}
func (namespace NamespaceModel) AllKinds() []string {
	keys := make([]string, 0, len(namespace.resourcesByKind))
	for k := range namespace.resourcesByKind {
		keys = append(keys, k)
	}
	return keys
}

func (namespace NamespaceModel) ResourcesByKind(kind string) []Resource {
	return namespace.resourcesByKind[kind]
}
func (namespace NamespaceModel) AllResources() []Resource {
	resources := make([]Resource, 0)
	for kind := range namespace.resourcesByKind {
		resources = append(resources, namespace.resourcesByKind[kind]...)
	}
	return resources
}

func (namespace *NamespaceModel) AddConnection(from Resource, to Resource) *Connection {
	for _, c := range namespace.connections {
		if reflect.DeepEqual(c.From, from) && reflect.DeepEqual(c.To, to) {
			logger.Debugf("Skipped existing connection from %s of kind %s and %s of kind %s", from.Name(), from.Kind(), to.Name(), to.Kind())
			return &c
		}
	}

	connection := Connection{From: from, To: to}
	namespace.connections = append(namespace.connections, connection)
	return &namespace.connections[len(namespace.connections)-1]
}
func (namespace *NamespaceModel) AddNamedConnection(from Resource, to Resource, name string) *Connection {
	connection := namespace.AddConnection(from, to)
	connection.Name = name
	return connection
}

func (namespace NamespaceModel) AllConnections() []Connection {
	return namespace.connections
}
