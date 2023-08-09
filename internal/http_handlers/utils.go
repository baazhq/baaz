package http_handlers

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog/v2"
)

type client struct {
	*kubernetes.Clientset
}

func getKubeClientset() *client {

	var conf *rest.Config
	var err error

	if os.Getenv("RUN_LOCAL") == "true" {
		// for running locally
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		conf, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	} else {
		// creates the in-cluster config
		conf, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}
	cs, err := kubernetes.NewForConfig(conf)
	if err != nil {
		klog.Infof("error in getting clientset from Kubeconfig: %v", err)
	}

	return &client{cs}
}

func mergeMaps(m1 map[string]string, m2 map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}
