package main

import (
	"flag"
	"fmt"
	golog "log"
	"os"
	"path/filepath"

	"github.com/dmartinol/openshift-topology-exporter/pkg/builder"
	config "github.com/dmartinol/openshift-topology-exporter/pkg/config"
	log "github.com/dmartinol/openshift-topology-exporter/pkg/log"
	t "github.com/dmartinol/openshift-topology-exporter/pkg/transformer"

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

func start() error {
	exporterConfig = *config.ReadConfig()
	golog.Printf("Starting with configuration: %+v\n", exporterConfig)
	log.InitLogger(exporterConfig)

	formatter := t.NewFormatterForConfig(exporterConfig)
	transformer := t.NewTransformer(formatter)

	config, err := connectCluster()
	if err != nil {
		return err
	}
	log.Info("Cluster connected")

	topology, err := builder.NewModelBuilder(exporterConfig).BuildForConfig(config)
	if err != nil {
		return err
	}
	output, err := transformer.Transform(*topology)
	if err != nil {
		return err
	}
	log.Debugf("Outout is %s", output)
	return err
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
