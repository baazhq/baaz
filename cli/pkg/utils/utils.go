package utils

import (
	"context"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetLocalKubeClientset() *kubernetes.Clientset {

	var conf *rest.Config
	var err error

	// for running locally

	var kubeconfig string
	path, ok := os.LookupEnv("KUBECONFIG")
	if ok {
		kubeconfig = path
	} else {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

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

func CreateNamespace(cs *kubernetes.Clientset, customerName string) error {
	_, err := cs.CoreV1().Namespaces().Create(context.TODO(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: customerName,
			Labels: map[string]string{
				"private":                  "true",
				"customer_" + customerName: customerName,
				"baaz":                     "controlplane",
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
