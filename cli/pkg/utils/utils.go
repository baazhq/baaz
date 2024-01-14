package utils

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetKubeClientset() *kubernetes.Clientset {

	var conf *rest.Config
	var err error

	// for running locally
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	conf, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	cs, err := kubernetes.NewForConfig(conf)
	if err != nil {
		panic(err.Error())
	}

	return cs
}
