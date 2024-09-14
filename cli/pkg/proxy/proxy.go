package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func Run() {
	namespace := "baaz"
	serviceName := "baaz"
	localPort := "8000"
	svcPort := "8000"

	// Initialize Kubernetes client
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		fmt.Printf("Error building Kubernetes client config: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// Check if the service exists
	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			fmt.Printf("Service %s not found in namespace %s\n", serviceName, namespace)
		} else {
			fmt.Printf("Error retrieving service: %v\n", err)
		}
		os.Exit(1)
	}

	// Check if the service has the correct ports
	if len(service.Spec.Ports) == 0 {
		fmt.Printf("Service %s has no ports defined\n", serviceName)
		os.Exit(1)
	}

	// Port-forward service
	podName, err := getFirstRunningPod(clientset, namespace)
	if err != nil {
		fmt.Printf("Error finding running pod for service: %v\n", err)
		os.Exit(1)
	}

	err = forwardServicePort(config, namespace, podName, localPort, svcPort)
	if err != nil {
		fmt.Printf("Error port-forwarding: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Service %s is now accessible on localhost:%s\n", serviceName, localPort)
}

// Get the first running pod in the namespace (since port-forward needs a pod)
func getFirstRunningPod(clientset *kubernetes.Clientset, namespace string) (string, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: "status.phase=Running",
	})
	if err != nil {
		return "", fmt.Errorf("error listing pods: %w", err)
	}
	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no running pods found in namespace %s", namespace)
	}
	return pods.Items[0].Name, nil
}

// Set up port forwarding using client-go portforward package
func forwardServicePort(config *restclient.Config, namespace, podName, localPort, svcPort string) error {
	// Create spdy roundtripper for port forwarding
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, podName)
	hostIP := strings.TrimLeft(config.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return fmt.Errorf("error creating round tripper: %w", err)
	}

	// Create the request for port forwarding
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", config.Host, path), nil)
	if err != nil {
		return fmt.Errorf("error creating port forward request: %w", err)
	}

	req.Header.Add("Host", hostIP)

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL)
	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{}, 1)
	out := io.Discard // Change this to os.Stdout if you want to see output
	errOut := os.Stderr

	ports := []string{fmt.Sprintf("%s:%s", localPort, svcPort)}

	pf, err := portforward.New(dialer, ports, stopChan, readyChan, out, errOut)
	if err != nil {
		return fmt.Errorf("error creating port forwarder: %w", err)
	}

	go func() {
		<-readyChan // Block until the port forward is ready
		fmt.Println("Port forwarding is ready")
	}()

	// Start the port forwarding process
	if err := pf.ForwardPorts(); err != nil {
		return fmt.Errorf("error forwarding ports: %w", err)
	}

	return nil
}
