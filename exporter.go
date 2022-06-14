package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmartinol/openshift-topology-exporter/pkg/exporter"

	authv1T "github.com/openshift/api/authorization/v1"
	routev1T "github.com/openshift/api/route/v1"
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

var resourcesByKind map[string][]exporter.Resource = make(map[string][]exporter.Resource)

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

var diagram strings.Builder = strings.Builder{}

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

	file, err := os.Create("diagram.dot")
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(diagram.String())
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

var clusterCount int = 0

func diagramOf(ns string) error {
	diagram.WriteString(fmt.Sprintf("subgraph cluster_%d {\n", clusterCount))
	diagram.WriteString("style=filled;\n")
	diagram.WriteString("color=lightgrey;\n")
	diagram.WriteString("node [style=filled,color=white];\n")
	diagram.WriteString(fmt.Sprintf("label =\"%s\";\n", ns))
	clusterCount++

	//TBD: we should move to per NS model builder
	resourcesByKind = make(map[string][]exporter.Resource)

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
		resource := exporter.Route{Delegate: route}
		addResource(resource)
	}

	fmt.Println("=== Services ===")
	services, err := coreClient.Services(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, service := range services.Items {
		fmt.Printf("Found %s/%s\n", service.Kind, service.Name)
		resource := exporter.Service{Delegate: service}
		addResource(resource)
	}

	fmt.Println("=== Deployments ===")
	deployments, err := appsV1Client.Deployments(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, deployment := range deployments.Items {
		fmt.Printf("Found %s/%s\n", deployment.Kind, deployment.Name)
		resource := exporter.Deployment{Delegate: deployment}
		addResource(resource)
	}

	fmt.Println("=== StatefulSets ===")
	statefulSets, err := appsV1Client.StatefulSets(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, statefulSet := range statefulSets.Items {
		fmt.Printf("Found %s/%s\n", statefulSet.Kind, statefulSet.Name)
		resource := exporter.StatefulSet{Delegate: statefulSet}
		addResource(resource)
	}

	fmt.Println("=== DeploymentConfigs ===")
	deploymentConfigs, err := appsClient.DeploymentConfigs(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, deploymentConfig := range deploymentConfigs.Items {
		fmt.Printf("Found %s/%s\n", deploymentConfig.Kind, deploymentConfig.Name)
		resource := exporter.DeploymentConfig{Delegate: deploymentConfig}
		addResource(resource)
	}

	fmt.Println("=== Pods ===")
	pods, err := coreClient.Pods(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		fmt.Printf("Found %s/%s with SA %s\n", pod.Kind, pod.Name, pod.Spec.ServiceAccountName)
		resource := exporter.Pod{Delegate: pod}
		addResource(resource)

		serviceAccount, err := coreClient.ServiceAccounts(ns).Get(context.TODO(), pod.Spec.ServiceAccountName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		saResource := exporter.ServiceAccount{Delegate: *serviceAccount}
		addResource(saResource)

		saRoleBindings := saResource.TheRoleBindings(roleBindings)
		for _, roleBinding := range saRoleBindings {
			fmt.Printf("For SA %s found RoleBinding %s/%s\n", serviceAccount.Name, roleBinding.RoleRef.Name, roleBinding.UserNames)
			rbResource := exporter.RoleBinding{Delegate: roleBinding}
			addResource(rbResource)
			diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", saResource.Id(), rbResource.Id()))
		}
		saClusterRoleBindings := saResource.TheClusterRoleBindings(clusterRoleBindings)
		for _, clusterRoleBinding := range saClusterRoleBindings {
			fmt.Printf("For SA %s found ClusterRoleBinding %s/%s\n", serviceAccount.Name, clusterRoleBinding.RoleRef.Name, clusterRoleBinding.UserNames)
			rbResource := exporter.ClusterRoleBinding{Delegate: clusterRoleBinding}
			addResource(rbResource)
			diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", saResource.Id(), rbResource.Id()))
		}
	}

	addOwners()
	connectResources()

	diagram.WriteString("\n}")
	return nil
}

func addResource(resource exporter.Resource) {
	if lookupByKindAndName(resource.Kind(), resource.Id()) == nil {
		resourcesByKind[resource.Kind()] = append(resourcesByKind[resource.Kind()], resource)

		color, hasStatusColor := resource.StatusColor()
		if hasStatusColor {
			diagram.WriteString(fmt.Sprintf("\"%s\" [ label=\"%s\", image=\"%s\", labelloc=b, color=\"%s\" ];\n",
				resource.Id(), resource.Label(), resource.Icon(), color))
		} else {
			diagram.WriteString(fmt.Sprintf("\"%s\" [ label=\"%s\", image=\"%s\", labelloc=b ];\n",
				resource.Id(), resource.Label(), resource.Icon()))
		}
		fmt.Printf("Added instance %s of kind %s\n", resource.Label(), resource.Kind())
	} else {
		fmt.Printf("Skipped already added instance %s of kind %s\n", resource.Label(), resource.Kind())
	}
}

func connectResources() {
	for _, owners := range resourcesByKind {
		for _, owner := range owners {
			for _, kind := range owner.ConnectedKinds() {
				fmt.Printf("Connecting %s of kind %s to %v of kind %s\n",
					owner.Label(), owner.Kind(), len(resourcesByKind[kind]), kind)
				diagram.WriteString(owner.ConnectTo(kind, resourcesByKind[kind]))
			}
		}
	}
}

func lookupByKindAndName(kind string, name string) exporter.Resource {
	for _, resource := range resourcesByKind[kind] {
		if strings.Compare(name, resource.Name()) == 0 {
			return resource
		}
	}

	return nil
}

func lookupOwner(owner metav1.OwnerReference) exporter.Resource {
	for _, resources := range resourcesByKind {
		for _, resource := range resources {
			if resource.IsOwnerOf(owner) {
				return resource
			}
		}
	}

	if strings.Compare(owner.Kind, "ClusterServiceVersion") == 0 {
		csvResource := exporter.ClusterServiceVersion{Delegate: owner}
		addResource(csvResource)
		return csvResource
	} else {
		// TODO Just guessing ....
		customResource := exporter.CustomResource{Delegate: owner}
		addResource(customResource)
		return customResource
	}
}

func addOwners() {
	for _, resources := range resourcesByKind {
		for _, resource := range resources {
			for _, owner := range resource.OwnerReferences() {
				fmt.Printf("Adding ownership of %s of kind %s to %s of kind %s\n",
					resource.Label(), resource.Kind(), owner.Name, owner.Kind)
				ownerResource := lookupOwner(owner)
				diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\" [label=\"owns\"];\n", ownerResource.Id(), resource.Id()))
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

	diagram.Reset()
	diagram.WriteString("digraph G {\n")
	diagram.WriteString("node [shape=plaintext];\n")
	legend()
	for _, namespace := range namespaces {
		diagram.WriteString("\n")
		err := diagramOf(namespace)
		if err != nil {
			return err
		}
	}
	diagram.WriteString("\n}")
	return nil
}

func routeLabel(route routev1T.Route) string {
	return fmt.Sprintf("Route %s", route.Name)
}
func roleLabel(roleName string) string {
	return fmt.Sprintf("Role %s", roleName)
}

func legend() {
	diagram.WriteString("subgraph legend {\n")
	diagram.WriteString("legend [\n")
	diagram.WriteString("label=<<TABLE border=\"0\" cellspacing=\"2\" cellpadding=\"0\">\n")
	diagram.WriteString(fmt.Sprintf("<TR><TD border=\"0\" bgcolor=\"%s\">Completed</TD></TR>\n", exporter.CompletedColor))
	diagram.WriteString(fmt.Sprintf("<TR><TD border=\"0\" bgcolor=\"%s\">Running</TD></TR>\n", exporter.RunningColor))
	diagram.WriteString(fmt.Sprintf("<TR><TD border=\"0\" bgcolor=\"%s\">Failed</TD></TR>\n", exporter.FailedColor))
	diagram.WriteString("<TR><TD>Legend</TD></TR>\n")
	diagram.WriteString("</TABLE>>];\n")
	diagram.WriteString("}\n")
}
