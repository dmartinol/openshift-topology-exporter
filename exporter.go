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

	appsv1T "github.com/openshift/api/apps/v1"
	authv1T "github.com/openshift/api/authorization/v1"
	routev1T "github.com/openshift/api/route/v1"
	appsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	authv1 "github.com/openshift/client-go/authorization/clientset/versioned/typed/authorization/v1"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	k8appsv1T "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
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

	diagram, err := clusterDiagramOf(exporterConfig.Namespaces)
	if err != nil {
		return err
	}

	file, err := os.Create("diagram.dot")
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(diagram)
	return nil
}

func roleBindingsOf(serviceAccount v1.ServiceAccount, roleBindings *authv1T.RoleBindingList) []authv1T.RoleBinding {
	var saRoleBindings []authv1T.RoleBinding
	userName := fmt.Sprintf("system:serviceaccount:%s:%s", serviceAccount.Namespace, serviceAccount.Name)
	for _, roleBinding := range roleBindings.Items {
		for _, rbUserName := range roleBinding.UserNames {
			if strings.Compare(rbUserName, userName) == 0 {
				saRoleBindings = append(saRoleBindings, roleBinding)
			}
		}
	}
	return saRoleBindings
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

func diagramOf(ns string) (string, error) {
	diagram := strings.Builder{}
	diagram.WriteString(fmt.Sprintf("subgraph cluster_%d {\n", clusterCount))
	diagram.WriteString("style=filled;\n")
	diagram.WriteString("color=lightgrey;\n")
	diagram.WriteString("node [style=filled,color=white];\n")
	diagram.WriteString(fmt.Sprintf("label =\"%s\";\n", ns))
	clusterCount++

	fmt.Printf("Running on NS %s\n", ns)
	roleBindings, err := authClient.RoleBindings(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, roleBinding := range roleBindings.Items {
		fmt.Printf("Found RoleBinding %s/%s\n", roleBinding.RoleRef.Name, roleBinding.UserNames)
	}

	fmt.Println("=== Routes ===")
	routes, err := routeClient.Routes(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, route := range routes.Items {
		fmt.Printf("Found %s/%s\n", route.Kind, route.Name)
		diagram.WriteString(routeNodeOf(route))
		diagram.WriteString(connectRouteToService(route))
	}

	fmt.Println("=== Services ===")
	services, err := coreClient.Services(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, service := range services.Items {
		fmt.Printf("Found %s/%s\n", service.Kind, service.Name)
		diagram.WriteString(serviceNodeOf(service))
	}

	fmt.Println("=== Deployments ===")
	deployments, err := appsV1Client.Deployments(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, deployment := range deployments.Items {
		fmt.Printf("Found %s/%s\n", deployment.Kind, deployment.Name)
		diagram.WriteString(deploymentNodeOf(deployment))
		diagram.WriteString(addOwners(deployment.OwnerReferences, deploymentLabel(deployment)))
	}

	fmt.Println("=== StatefulSets ===")
	statefulSets, err := appsV1Client.StatefulSets(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, statefulSet := range statefulSets.Items {
		fmt.Printf("Found %s/%s\n", statefulSet.Kind, statefulSet.Name)
		diagram.WriteString(statefulSetNodeOf(statefulSet))
		diagram.WriteString(addOwners(statefulSet.OwnerReferences, statefulSetLabel(statefulSet)))
	}

	fmt.Println("=== DeploymentConfigs ===")
	deploymentConfigs, err := appsClient.DeploymentConfigs(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, deploymentConfig := range deploymentConfigs.Items {
		fmt.Printf("Found %s/%s\n", deploymentConfig.Kind, deploymentConfig.Name)
		diagram.WriteString(deploymentConfigNodeOf(deploymentConfig))
		diagram.WriteString(addOwners(deploymentConfig.OwnerReferences, deploymentConfigLabel(deploymentConfig)))
	}

	fmt.Println("=== Pods ===")
	pods, err := coreClient.Pods(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, pod := range pods.Items {
		fmt.Printf("Found %s/%s\n", pod.Kind, pod.Name)
		diagram.WriteString(podNodeOf(pod))
		// Pod -> ReplicationController -> DeploymentConfig
		// diagram.WriteString(addOwners(pod.OwnerReferences, podLabel(pod)))

		serviceAccount, err := coreClient.ServiceAccounts(ns).Get(context.TODO(), pod.Spec.ServiceAccountName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		fmt.Printf("Found ServiceAccount %s/%s\n", serviceAccount.Kind, serviceAccount.Name)
		diagram.WriteString(serviceAccountNodeOf(*serviceAccount))
		diagram.WriteString(connectPodToServiceAccount(pod, *serviceAccount))
		saRoleBindings := roleBindingsOf(*serviceAccount, roleBindings)
		for _, roleBinding := range saRoleBindings {
			fmt.Printf("Found RoleBinding %s/%s\n", roleBinding.RoleRef.Name, roleBinding.UserNames)
			diagram.WriteString(roleNodeOf(roleBinding))
			diagram.WriteString(connectServiceAccountToRole(*serviceAccount, roleBinding))
		}
	}

	for _, service := range services.Items {
		diagram.WriteString(connectServiceToPods(service, pods.Items))
	}
	for _, deployment := range deployments.Items {
		diagram.WriteString(connectDeploymentToPods(deployment, pods.Items))
	}
	for _, statefulSet := range statefulSets.Items {
		diagram.WriteString(connectStatefulSetToPods(statefulSet, pods.Items))
	}
	for _, deploymentConfig := range deploymentConfigs.Items {
		diagram.WriteString(connectDeploymentConfigToPods(deploymentConfig, pods.Items))
	}
	diagram.WriteString("\n}")
	return diagram.String(), nil
}

func routeNodeOf(route routev1T.Route) string {
	diagram := strings.Builder{}
	diagram.WriteString(fmt.Sprintf("\"%s\" [ image=\"images/ingress.png\", labelloc=b ];\n", routeLabel(route)))
	return diagram.String()
}

func serviceNodeOf(service v1.Service) string {
	diagram := strings.Builder{}
	diagram.WriteString(fmt.Sprintf("\"%s\" [ image=\"images/svc.png\", labelloc=b ];\n", serviceLabel(service)))
	return diagram.String()
}

func deploymentNodeOf(deployment k8appsv1T.Deployment) string {
	diagram := strings.Builder{}
	diagram.WriteString(fmt.Sprintf("\"%s\" [ image=\"images/deployment.png\", labelloc=b ];\n", deploymentLabel(deployment)))
	return diagram.String()
}
func statefulSetNodeOf(statefulSet k8appsv1T.StatefulSet) string {
	diagram := strings.Builder{}
	diagram.WriteString(fmt.Sprintf("\"%s\" [ image=\"images/sts.png\", labelloc=b ];\n", statefulSetLabel(statefulSet)))
	return diagram.String()
}
func deploymentConfigNodeOf(deploymentConfig appsv1T.DeploymentConfig) string {
	diagram := strings.Builder{}
	diagram.WriteString(fmt.Sprintf("\"%s\" [ image=\"images/deployment.png\", labelloc=b ];\n", deploymentConfigLabel(deploymentConfig)))
	return diagram.String()
}

func podNodeOf(pod v1.Pod) string {
	diagram := strings.Builder{}

	diagram.WriteString(fmt.Sprintf("\"%s\" [ image=\"images/pod.png\", labelloc=b, color=\"%s\" ];\n", podLabel(pod), podStatusColor(pod)))
	return diagram.String()
}

func serviceAccountNodeOf(serviceAccount v1.ServiceAccount) string {
	diagram := strings.Builder{}

	diagram.WriteString(fmt.Sprintf("\"%s\" [ image=\"images/sa.png\", labelloc=b ];\n", serviceAccountLabel(serviceAccount)))
	return diagram.String()
}
func roleNodeOf(roleBinding authv1T.RoleBinding) string {
	diagram := strings.Builder{}

	diagram.WriteString(fmt.Sprintf("\"%s\" [ image=\"images/role.png\", labelloc=b ];\n", roleLabel(roleBinding.RoleRef.Name)))
	return diagram.String()
}

func addOwners(owners []metav1.OwnerReference, childLabel string) string {
	diagram := strings.Builder{}
	for _, owner := range owners {
		fmt.Printf("Owner is %s/%s/%s\n", owner.Kind, owner.APIVersion, owner.Name)
		diagram.WriteString(addOwner(owner, childLabel))
	}
	return diagram.String()
}

func addOwner(owner metav1.OwnerReference, childLabel string) string {
	diagram := strings.Builder{}
	if !isManagedKind(owner.Kind) {
		diagram.WriteString(ownerNodeOf(owner))
	} else {
		fmt.Printf("Discarding owner of managed kind %s\n", owner.Kind)
	}
	diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\" [label=\"owns\"];\n", ownerLabel(owner), childLabel))
	return diagram.String()
}

func ownerNodeOf(owner metav1.OwnerReference) string {
	diagram := strings.Builder{}

	icon := "crd.png"
	if strings.Compare(owner.Kind, "ClusterServiceVersion") == 0 {
		icon = "operator.png"
	}
	diagram.WriteString(fmt.Sprintf("\"%s\" [ image=\"images/%s\", labelloc=b ];\n", ownerLabel(owner), icon))
	return diagram.String()
}

func connectRouteToService(route routev1T.Route) string {
	diagram := strings.Builder{}
	if strings.Compare(route.Spec.To.Kind, "Service") == 0 {
		serviceName := route.Spec.To.Name
		diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", routeLabel(route), serviceLabelName(serviceName)))
	}
	return diagram.String()
}

func matchSelector(service v1.Service, pod v1.Pod) bool {
	for label, value := range service.Spec.Selector {
		found := false
		for podLabel, podValue := range pod.ObjectMeta.Labels {
			if strings.Compare(podLabel, label) == 0 {
				if strings.Compare(podValue, value) != 0 {
					return false
				}
				found = true
			}
		}
		if !found {
			return false
		}
	}

	return len(service.Spec.Selector) > 0
}

func connectServiceToPods(service v1.Service, pods []v1.Pod) string {
	diagram := strings.Builder{}
	for _, pod := range pods {
		if matchSelector(service, pod) {
			diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", serviceLabel(service), podLabel(pod)))
		}
	}
	return diagram.String()
}
func connectDeploymentToPods(deployment k8appsv1T.Deployment, pods []v1.Pod) string {
	diagram := strings.Builder{}

	for _, pod := range pods {
		for _, owner := range pod.OwnerReferences {
			// fmt.Printf("Owner %s, %s, comparing with %s=%s\n", owner.Kind, owner.Name, deploymentConfig.Name+"-", strings.HasPrefix(owner.Name, deploymentConfig.Name+"-"))
			if strings.Compare(owner.Kind, "ReplicaSet") == 0 && strings.HasPrefix(owner.Name, deployment.Name+"-") {
				fmt.Printf("ADDING CONNECTION %s %s\n", deploymentLabel(deployment), podLabel(pod))
				fmt.Printf("\"%s\" -> \"%s\"\n", deploymentLabel(deployment), podLabel(pod))
				diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", deploymentLabel(deployment), podLabel(pod)))
			}
		}
	}
	return diagram.String()
}
func connectStatefulSetToPods(statefulSet k8appsv1T.StatefulSet, pods []v1.Pod) string {
	diagram := strings.Builder{}

	for _, pod := range pods {
		for _, owner := range pod.OwnerReferences {
			// fmt.Printf("Owner %s, %s, comparing with %s=%s\n", owner.Kind, owner.Name, deploymentConfig.Name+"-", strings.HasPrefix(owner.Name, deploymentConfig.Name+"-"))
			if strings.Compare(owner.Kind, "StatefulSet") == 0 && strings.Compare(owner.Name, statefulSet.Name) == 0 {
				fmt.Printf("ADDING CONNECTION %s %s\n", statefulSetLabel(statefulSet), podLabel(pod))
				fmt.Printf("\"%s\" -> \"%s\"\n", statefulSetLabel(statefulSet), podLabel(pod))
				diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", statefulSetLabel(statefulSet), podLabel(pod)))
			}
		}
	}
	return diagram.String()
}
func connectDeploymentConfigToPods(deploymentConfig appsv1T.DeploymentConfig, pods []v1.Pod) string {
	diagram := strings.Builder{}

	for _, pod := range pods {
		for _, owner := range pod.OwnerReferences {
			// fmt.Printf("Owner %s, %s, comparing with %s=%s\n", owner.Kind, owner.Name, deploymentConfig.Name+"-", strings.HasPrefix(owner.Name, deploymentConfig.Name+"-"))
			if strings.Compare(owner.Kind, "ReplicationController") == 0 && strings.HasPrefix(owner.Name, deploymentConfig.Name+"-") {
				fmt.Printf("ADDING CONNECTION %s %s\n", deploymentConfigLabel(deploymentConfig), podLabel(pod))
				fmt.Printf("\"%s\" -> \"%s\"\n", deploymentConfigLabel(deploymentConfig), podLabel(pod))
				diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", deploymentConfigLabel(deploymentConfig), podLabel(pod)))
			}
		}
	}
	return diagram.String()
}
func connectPodToServiceAccount(pod v1.Pod, serviceAccount v1.ServiceAccount) string {
	diagram := strings.Builder{}
	diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", podLabel(pod), serviceAccountLabel(serviceAccount)))
	return diagram.String()
}

func connectServiceAccountToRole(serviceAccount v1.ServiceAccount, roleBinding authv1T.RoleBinding) string {
	diagram := strings.Builder{}
	diagram.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", serviceAccountLabel(serviceAccount), roleLabel(roleBinding.RoleRef.Name)))
	return diagram.String()
}

func clusterDiagramOf(namespaces []string) (string, error) {
	diagram := strings.Builder{}
	diagram.WriteString("digraph G {\n")
	diagram.WriteString("node [shape=plaintext];\n")
	diagram.WriteString(legend())
	for _, namespace := range namespaces {
		diagram.WriteString("\n")
		nsDiagram, err := diagramOf(namespace)
		if err != nil {
			return "", err
		}
		diagram.WriteString(nsDiagram)
	}
	diagram.WriteString("\n}")
	return diagram.String(), nil
}

func routeLabel(route routev1T.Route) string {
	return fmt.Sprintf("Route %s", route.Name)
}
func serviceLabel(service v1.Service) string {
	return fmt.Sprintf("Service %s", service.Name)
}
func serviceLabelName(serviceName string) string {
	return fmt.Sprintf("Service %s", serviceName)
}
func deploymentLabel(deployment k8appsv1T.Deployment) string {
	return fmt.Sprintf("Deployment %s", deployment.Name)
}
func statefulSetLabel(statefulSet k8appsv1T.StatefulSet) string {
	return fmt.Sprintf("StatefulSet %s", statefulSet.Name)
}
func deploymentConfigLabel(deploymentConfig appsv1T.DeploymentConfig) string {
	return fmt.Sprintf("DeploymentConfig %s", deploymentConfig.Name)
}
func podLabel(pod v1.Pod) string {
	return fmt.Sprintf("Pod %s", pod.Name)
}
func serviceAccountLabel(serviceAccount v1.ServiceAccount) string {
	return fmt.Sprintf("system:serviceaccount:%s:%s", serviceAccount.Namespace, serviceAccount.Name)
}
func roleLabel(roleName string) string {
	return fmt.Sprintf("Role %s", roleName)
}

func ownerLabel(owner metav1.OwnerReference) string {
	if strings.Compare(owner.Kind, "ClusterServiceVersion") == 0 {
		return owner.Name
	}
	return fmt.Sprintf("%s %s", owner.Kind, owner.Name)
}

var managedKinds = map[string]string{"Route": "", "Service": "", "Deployment": "", "StatefulSet": "", "Pod": ""}

func isManagedKind(kind string) bool {
	_, hasKey := managedKinds[kind]
	return hasKey
}

func legend() string {
	diagram := strings.Builder{}
	diagram.WriteString("subgraph legend {\n")
	diagram.WriteString("legend [\n")
	diagram.WriteString("label=<<TABLE border=\"0\" cellspacing=\"2\" cellpadding=\"0\">\n")
	diagram.WriteString(fmt.Sprintf("<TR><TD border=\"0\" bgcolor=\"%s\">Completed</TD></TR>\n", completedColor))
	diagram.WriteString(fmt.Sprintf("<TR><TD border=\"0\" bgcolor=\"%s\">Running</TD></TR>\n", runningColor))
	diagram.WriteString(fmt.Sprintf("<TR><TD border=\"0\" bgcolor=\"%s\">Failed</TD></TR>\n", failedColor))
	diagram.WriteString("<TR><TD>Legend</TD></TR>\n")
	diagram.WriteString("</TABLE>>];\n")
	diagram.WriteString("}\n")

	return diagram.String()
}

const completedColor = "#66ff33"
const runningColor = "#00ffff"
const failedColor = "#ff3300"

func podStatusColor(pod v1.Pod) string {
	switch pod.Status.Phase {
	case "Succeeded":
		return completedColor
	case "Running":
		if isReady(pod) {
			return runningColor
		}
	}
	return failedColor
}

func isReady(pod v1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if strings.Compare(string(c.Type), "Ready") == 0 {
			return strings.Compare(string(c.Status), "True") == 0
		}
	}
	return false
}
