package model

type TopologyModel struct {
	namespacesByName map[string]*NamespaceModel
}

func NewTopologyModel() *TopologyModel {
	var topology TopologyModel
	topology.namespacesByName = make(map[string]*NamespaceModel)
	return &topology
}

func (topology TopologyModel) AddNamespace(name string) *NamespaceModel {
	namespace := NamespaceModel{name: name, resourcesByKind: make(map[string][]Resource)}
	topology.namespacesByName[name] = &namespace
	return &namespace
}
func (topology TopologyModel) NamespaceByName(name string) *NamespaceModel {
	return topology.namespacesByName[name]
}
func (topology TopologyModel) AllNamespaces() []NamespaceModel {
	namespaces := make([]NamespaceModel, 0, len(topology.namespacesByName))
	for _, namespace := range topology.namespacesByName {
		namespaces = append(namespaces, *namespace)
	}
	return namespaces
}
