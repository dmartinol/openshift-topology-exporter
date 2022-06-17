package main

import (
	"flag"
	"fmt"
	golog "log"
	"os"
	"path/filepath"

	config "github.com/dmartinol/openshift-topology-exporter/pkg/config"
	log "github.com/dmartinol/openshift-topology-exporter/pkg/log"
	"github.com/dmartinol/openshift-topology-exporter/pkg/model"
	t "github.com/dmartinol/openshift-topology-exporter/pkg/trasnformer"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	err := start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}

var exporterConfig config.ExporterConfig
var formatter = t.NewGraphVizFormatter()
var transformer = t.NewTransformer(formatter)

func start() error {
	exporterConfig = *config.ReadConfig()
	golog.Printf("Starting with configuration: %+v\n", exporterConfig)
	log.InitLogger(exporterConfig)

	config, err := connectCluster()
	if err != nil {
		return err
	}
	log.Info("Cluster connected")

	topology, err := model.NewModelBuilder(exporterConfig).BuildForConfig(config)
	if err != nil {
		return err
	}
	transformer.Transform(*topology)
	return nil
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
