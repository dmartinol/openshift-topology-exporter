package model

import (
	"context"
	"fmt"

	"github.com/dmartinol/openshift-topology-exporter/pkg/config"
	authv1T "github.com/openshift/api/authorization/v1"
	appsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	authv1 "github.com/openshift/client-go/authorization/clientset/versioned/typed/authorization/v1"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/rest"
)

type ModelBuilder struct {
	exporterConfig config.ExporterConfig
	routeClient    *routev1.RouteV1Client
	appsClient     *appsv1.AppsV1Client
	appsV1Client   *k8appsv1client.AppsV1Client
	coreClient     *corev1client.CoreV1Client
	authClient     *authv1.AuthorizationV1Client

	topologyModel       *TopologyModel
	namespaceModel      *NamespaceModel
	clusterRoleBindings *authv1T.ClusterRoleBindingList
}

func NewModelBuilder(exporterConfig config.ExporterConfig) *ModelBuilder {
	builder := ModelBuilder{exporterConfig: exporterConfig}
	builder.topologyModel = NewTopologyModel()
	return &builder
}

func (builder *ModelBuilder) BuildForConfig(config *rest.Config) (*TopologyModel, error) {
	var err error
	builder.routeClient, err = routev1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	builder.appsClient, err = appsv1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	builder.appsV1Client, err = k8appsv1client.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	builder.coreClient, err = corev1client.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	builder.authClient, err = authv1.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	err = builder.buildCluster()
	if err != nil {
		return nil, err
	}

	return builder.topologyModel, nil
}

func (builder *ModelBuilder) buildCluster() error {
	var err error
	builder.clusterRoleBindings, err = builder.authClient.ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, clusterRoleBinding := range builder.clusterRoleBindings.Items {
		fmt.Printf("Found ClusterRoleBindings %s/%s\n", clusterRoleBinding.RoleRef.Name, clusterRoleBinding.UserNames)
	}

	for _, namespace := range builder.exporterConfig.Namespaces {
		err := builder.buildNamespace(namespace)
		if err != nil {
			return err
		}
	}
	return nil
}

func (builder *ModelBuilder) buildNamespace(namespace string) error {
	builder.namespaceModel = builder.topologyModel.AddNamespace(namespace)

	fmt.Printf("Running on NS %s\n", namespace)
	roleBindings, err := builder.authClient.RoleBindings(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, roleBinding := range roleBindings.Items {
		fmt.Printf("Found RoleBinding %s/%s\n", roleBinding.RoleRef.Name, roleBinding.UserNames)
	}

	fmt.Println("=== Routes ===")
	routes, err := builder.routeClient.Routes(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, route := range routes.Items {
		fmt.Printf("Found %s/%s\n", route.Kind, route.Name)
		resource := Route{Delegate: route}
		builder.namespaceModel.AddResource(resource)
	}

	fmt.Println("=== Services ===")
	services, err := builder.coreClient.Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, service := range services.Items {
		fmt.Printf("Found %s/%s\n", service.Kind, service.Name)
		resource := Service{Delegate: service}
		builder.namespaceModel.AddResource(resource)
	}

	fmt.Println("=== Deployments ===")
	deployments, err := builder.appsV1Client.Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, deployment := range deployments.Items {
		fmt.Printf("Found %s/%s\n", deployment.Kind, deployment.Name)
		resource := Deployment{Delegate: deployment}
		builder.namespaceModel.AddResource(resource)
	}

	fmt.Println("=== StatefulSets ===")
	statefulSets, err := builder.appsV1Client.StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, statefulSet := range statefulSets.Items {
		fmt.Printf("Found %s/%s\n", statefulSet.Kind, statefulSet.Name)
		resource := StatefulSet{Delegate: statefulSet}
		builder.namespaceModel.AddResource(resource)
	}

	fmt.Println("=== DeploymentConfigs ===")
	deploymentConfigs, err := builder.appsClient.DeploymentConfigs(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, deploymentConfig := range deploymentConfigs.Items {
		fmt.Printf("Found %s/%s\n", deploymentConfig.Kind, deploymentConfig.Name)
		resource := DeploymentConfig{Delegate: deploymentConfig}
		builder.namespaceModel.AddResource(resource)
	}

	fmt.Println("=== Pods ===")
	pods, err := builder.coreClient.Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		fmt.Printf("Found %s/%s with SA %s\n", pod.Kind, pod.Name, pod.Spec.ServiceAccountName)
		resource := Pod{Delegate: pod}
		builder.namespaceModel.AddResource(resource)

		serviceAccount, err := builder.coreClient.ServiceAccounts(namespace).Get(context.TODO(), pod.Spec.ServiceAccountName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		saResource := ServiceAccount{Delegate: *serviceAccount}
		added := builder.namespaceModel.AddResource(saResource)
		if added {
			saRoleBindings := saResource.TheRoleBindings(roleBindings)
			for _, roleBinding := range saRoleBindings {
				fmt.Printf("For SA %s found RoleBinding %s/%s\n", serviceAccount.Name, roleBinding.RoleRef.Name, roleBinding.UserNames)
				rbResource := RoleBinding{Delegate: roleBinding}
				builder.namespaceModel.AddResource(rbResource)
				builder.namespaceModel.AddConnection(saResource, rbResource)
			}
			saClusterRoleBindings := saResource.TheClusterRoleBindings(builder.clusterRoleBindings)
			for _, clusterRoleBinding := range saClusterRoleBindings {
				fmt.Printf("For SA %s found ClusterRoleBinding %s/%s\n", serviceAccount.Name, clusterRoleBinding.RoleRef.Name, clusterRoleBinding.UserNames)
				rbResource := ClusterRoleBinding{Delegate: clusterRoleBinding}
				builder.namespaceModel.AddResource(rbResource)
				builder.namespaceModel.AddConnection(saResource, rbResource)
			}
		}
	}

	builder.addOwners()
	builder.connectResources()

	return nil
}

func (builder *ModelBuilder) connectResources() {
	for _, kind := range builder.namespaceModel.AllKinds() {
		for _, fromResource := range builder.namespaceModel.ResourcesByKind(kind) {
			for _, kind := range fromResource.ConnectedKinds() {
				potentialTos := builder.namespaceModel.ResourcesByKind(kind)
				connectedResources, connectionName := fromResource.ConnectedResources(kind, potentialTos)
				for _, connectedResource := range connectedResources {
					fmt.Printf("Connecting %s of kind %s to %s of kind %s with name %s\n",
						fromResource.Label(), fromResource.Kind(), connectedResource.Label(), connectedResource.Kind(), connectionName)
					if connectionName != "" {
						builder.namespaceModel.AddNamedConnection(fromResource, connectedResource, connectionName)
					} else {
						builder.namespaceModel.AddConnection(fromResource, connectedResource)
					}
					builder.namespaceModel.AllConnections()
				}
			}
		}
	}
}

func (builder *ModelBuilder) addOwners() {
	for _, kind := range builder.namespaceModel.AllKinds() {
		resourcesByKind := builder.namespaceModel.ResourcesByKind(kind)
		for _, resource := range resourcesByKind {
			for _, owner := range resource.OwnerReferences() {
				fmt.Printf("Adding ownership of %s of kind %s to %s of kind %s\n",
					resource.Label(), resource.Kind(), owner.Name, owner.Kind)
				ownerResource := builder.namespaceModel.LookupOwner(owner)
				builder.namespaceModel.AddNamedConnection(ownerResource, resource, "owns")
				builder.namespaceModel.AllConnections()
			}
		}
	}
}
