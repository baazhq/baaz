package eks

import (
	"encoding/base64"
	"fmt"
	"log"

	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go/aws"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

// Deployer is responsible for deploying apps
func (eksEnv *EksEnvironment) ReconcileDeployer() error {
	clientset, err := eksEnv.getEksConfig()
	if err != nil {
		return err
	}

	nodes, err := clientset.CoreV1().Nodes().List(eksEnv.Context, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error getting EKS nodes: %v", err)
	}

	fmt.Println(nodes.Items)

	return nil
}

func (eksEnv *EksEnvironment) getEksConfig() (*kubernetes.Clientset, error) {
	eksClient := awseks.NewFromConfig(eksEnv.Config)

	resultDescribe, err := eksClient.DescribeCluster(eksEnv.Context, &awseks.DescribeClusterInput{
		Name: &eksEnv.Env.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name,
	})
	if err != nil {
		return nil, err
	}

	return newClientset(resultDescribe.Cluster)

}

func newClientset(cluster *types.Cluster) (*kubernetes.Clientset, error) {
	log.Printf("%+v", cluster)
	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return nil, err
	}
	opts := &token.GetTokenOptions{
		ClusterID: *cluster.Name,
	}
	tok, err := gen.GetWithOptions(opts)
	if err != nil {
		return nil, err
	}
	ca, err := base64.StdEncoding.DecodeString(aws.StringValue(cluster.CertificateAuthority.Data))
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(
		&rest.Config{
			Host:        *cluster.Endpoint,
			BearerToken: tok.Token,
			TLSClientConfig: rest.TLSClientConfig{
				CAData: ca,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
