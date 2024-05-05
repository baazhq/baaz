package khota_handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "github.com/baazhq/baaz/api/v1/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubeClientset() (*kubernetes.Clientset, dynamic.Interface) {

	var conf *rest.Config
	var err error

	if os.Getenv("RUN_LOCAL") == "true" {
		// for running locally
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		//kubeconfig := filepath.Join("hack/aws")
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
		panic(err.Error())
	}

	dynClient, err := dynamic.NewForConfig(conf)
	if err != nil {
		panic(err.Error())
	}

	return cs, dynClient
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

func getNamespace(customerName string) string {
	if customerName == "" {
		return shared_namespace
	}

	return customerName
}

func makeDataPlaneName(cloudType v1.CloudType, customerName, region string) string {
	// here we assume, its a shared saas
	if customerName == "" {
		return string(cloudType) + "-" + region + "-" + String(4)
	}
	return customerName + "-" + string(cloudType) + "-" + region + "-" + String(4)
}

func makeTenantName(appName, appSize string) string {
	s := appName + "-" + appSize
	return strings.ToLower(s) + "-" + String(4)
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

const charset = "abcdefghijklmnopqrstuvwxyz"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return stringWithCharset(length, charset)
}

func labels2Slice(labels map[string]string) []string {
	var sliceString []string
	for k, v := range labels {
		if strings.Contains(k, "customer_") {
			sliceString = append(sliceString, v)
		}
	}
	return sliceString
}

func checkKeyInMap(matchString string, labels map[string]string) bool {

	for k := range labels {
		if strings.Contains(k, matchString) {
			return true
		}
	}
	return false
}

func checkValueInMap(matchString string, labels map[string]string) bool {

	for _, v := range labels {
		if strings.Contains(v, matchString) {
			return true
		}
	}
	return false
}

func matchStringInMap(matchString string, labels map[string]string) string {

	for k, v := range labels {
		if strings.Contains(k, matchString) {
			return v
		}
	}
	return ""
}

// patchValue specifies a patch operation.
type patchValue struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// constructor for patchValue{}
func NewPatchValue(op, path string, value interface{}) []byte {
	patchPayload := make([]patchValue, 1)

	patchPayload[0].Op = op
	patchPayload[0].Path = path
	patchPayload[0].Value = value

	bytes, _ := json.Marshal(patchPayload)
	return bytes
}

func Ptr[T any](v T) *T {
	return &v
}

func JsonWrap(data string) ([]byte, error) {
	payload := strings.NewReader(fmt.Sprintf(`[%s]`, data))
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
