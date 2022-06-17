package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/dmartinol/openshift-topology-exporter/pkg/model"
	t "github.com/dmartinol/openshift-topology-exporter/pkg/trasnformer"

	authv1T "github.com/openshift/api/authorization/v1"
	appsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	authv1 "github.com/openshift/client-go/authorization/clientset/versioned/typed/authorization/v1"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ExporterConfig struct {
	Namespaces []string `yaml:",flow"`
}

func main() {
	err := start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}

var exporterConfig ExporterConfig
var routeClient *routev1.RouteV1Client
var appsClient *appsv1.AppsV1Client
var appsV1Client *k8appsv1client.AppsV1Client
var coreClient *corev1client.CoreV1Client
var authClient *authv1.AuthorizationV1Client

var topologyModel = *model.NewTopologyModel()
var namespaceModel *model.NamespaceModel

var formatter = t.NewGraphVizFormatter()
var transformer = t.NewTransformer(formatter)

func start() error {
	exporterConfig = readConfig()
	fmt.Printf("Config -> %v\n", exporterConfig)

	config, err := connectCluster()
	if err != nil {
		return err
	}

	routeClient, err = routev1.NewForConfig(config)
	if err != nil {
		return err
	}
	appsClient, err = appsv1.NewForConfig(config)
	if err != nil {
		return err
	}
	appsV1Client, err = k8appsv1client.NewForConfig(config)
	if err != nil {
		return err
	}
	coreClient, err = corev1client.NewForConfig(config)
	if err != nil {
		return err
	}
	authClient, err = authv1.NewForConfig(config)
	if err != nil {
		return err
	}

	err = clusterDiagramOf(exporterConfig.Namespaces)
	if err != nil {
		return err
	}

	transformer.Transform(topologyModel)
	return nil
}

func readConfig() ExporterConfig {
	yfile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	exporterConfig := ExporterConfig{}
	err2 := yaml.Unmarshal(yfile, &exporterConfig)

	if err2 != nil {
		log.Fatal(err2)
	}
	return exporterConfig
}

func connectCluster() (*rest.Config, error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "")
	}

	//Load config for Openshift's go-client from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	return config, err
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func diagramOf(ns string) error {
	namespaceModel = topologyModel.AddNamespace(ns)

	fmt.Printf("Running on NS %s\n", ns)
	roleBindings, err := authClient.RoleBindings(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, roleBinding := range roleBindings.Items {
		fmt.Printf("Found RoleBinding %s/%s\n", roleBinding.RoleRef.Name, roleBinding.UserNames)
	}

	fmt.Println("=== Routes ===")
	routes, err := routeClient.Routes(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, route := range routes.Items {
		fmt.Printf("Found %s/%s\n", route.Kind, route.Name)
		resource := model.Route{Delegate: route}
		namespaceModel.AddResource(resource)
	}

	fmt.Println("=== Services ===")
	services, err := coreClient.Services(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, service := range services.Items {
		fmt.Printf("Found %s/%s\n", service.Kind, service.Name)
		resource := model.Service{Delegate: service}
		namespaceModel.AddResource(resource)
	}

	fmt.Println("=== Deployments ===")
	deployments, err := appsV1Client.Deployments(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, deployment := range deployments.Items {
		fmt.Printf("Found %s/%s\n", deployment.Kind, deployment.Name)
		resource := model.Deployment{Delegate: deployment}
		namespaceModel.AddResource(resource)
	}

	fmt.Println("=== StatefulSets ===")
	statefulSets, err := appsV1Client.StatefulSets(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, statefulSet := range statefulSets.Items {
		fmt.Printf("Found %s/%s\n", statefulSet.Kind, statefulSet.Name)
		resource := model.StatefulSet{Delegate: statefulSet}
		namespaceModel.AddResource(resource)
	}

	fmt.Println("=== DeploymentConfigs ===")
	deploymentConfigs, err := appsClient.DeploymentConfigs(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, deploymentConfig := range deploymentConfigs.Items {
		fmt.Printf("Found %s/%s\n", deploymentConfig.Kind, deploymentConfig.Name)
		resource := model.DeploymentConfig{Delegate: deploymentConfig}
		namespaceModel.AddResource(resource)
	}

	fmt.Println("=== Pods ===")
	pods, err := coreClient.Pods(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		fmt.Printf("Found %s/%s with SA %s\n", pod.Kind, pod.Name, pod.Spec.ServiceAccountName)
		resource := model.Pod{Delegate: pod}
		namespaceModel.AddResource(resource)

		serviceAccount, err := coreClient.ServiceAccounts(ns).Get(context.TODO(), pod.Spec.ServiceAccountName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		saResource := model.ServiceAccount{Delegate: *serviceAccount}
		added := namespaceModel.AddResource(saResource)
		if added {
			saRoleBindings := saResource.TheRoleBindings(roleBindings)
			for _, roleBinding := range saRoleBindings {
				fmt.Printf("For SA %s found RoleBinding %s/%s\n", serviceAccount.Name, roleBinding.RoleRef.Name, roleBinding.UserNames)
				rbResource := model.RoleBinding{Delegate: roleBinding}
				namespaceModel.AddResource(rbResource)
				namespaceModel.AddConnection(saResource, rbResource)
			}
			saClusterRoleBindings := saResource.TheClusterRoleBindings(clusterRoleBindings)
			for _, clusterRoleBinding := range saClusterRoleBindings {
				fmt.Printf("For SA %s found ClusterRoleBinding %s/%s\n", serviceAccount.Name, clusterRoleBinding.RoleRef.Name, clusterRoleBinding.UserNames)
				rbResource := model.ClusterRoleBinding{Delegate: clusterRoleBinding}
				namespaceModel.AddResource(rbResource)
				namespaceModel.AddConnection(saResource, rbResource)
			}
		}
	}

	addOwners()
	connectResources()

	return nil
}

// TODO move to builder
func connectResources() {
	for _, kind := range namespaceModel.AllKinds() {
		for _, fromResource := range namespaceModel.ResourcesByKind(kind) {
			for _, kind := range fromResource.ConnectedKinds() {
				potentialTos := namespaceModel.ResourcesByKind(kind)
				connectedResources, connectionName := fromResource.ConnectedResources(kind, potentialTos)
				for _, connectedResource := range connectedResources {
					fmt.Printf("Connecting %s of kind %s to %s of kind %s with name %s\n",
						fromResource.Label(), fromResource.Kind(), connectedResource.Label(), connectedResource.Kind(), connectionName)
					if connectionName != "" {
						namespaceModel.AddNamedConnection(fromResource, connectedResource, connectionName)
					} else {
						namespaceModel.AddConnection(fromResource, connectedResource)
					}
					namespaceModel.AllConnections()
				}
			}
		}
	}
}

func addOwners() {
	for _, kind := range namespaceModel.AllKinds() {
		resourcesByKind := namespaceModel.ResourcesByKind(kind)
		for _, resource := range resourcesByKind {
			for _, owner := range resource.OwnerReferences() {
				fmt.Printf("Adding ownership of %s of kind %s to %s of kind %s\n",
					resource.Label(), resource.Kind(), owner.Name, owner.Kind)
				ownerResource := namespaceModel.LookupOwner(owner)
				namespaceModel.AddNamedConnection(ownerResource, resource, "owns")
				namespaceModel.AllConnections()
			}
		}
	}
}

var clusterRoleBindings *authv1T.ClusterRoleBindingList

func clusterDiagramOf(namespaces []string) error {
	var err error
	clusterRoleBindings, err = authClient.ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, clusterRoleBinding := range clusterRoleBindings.Items {
		fmt.Printf("Found ClusterRoleBindings %s/%s\n", clusterRoleBinding.RoleRef.Name, clusterRoleBinding.UserNames)
	}

	for _, namespace := range namespaces {
		err := diagramOf(namespace)
		if err != nil {
			return err
		}
	}
	return nil
}
