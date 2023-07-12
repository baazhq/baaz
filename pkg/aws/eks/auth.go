package eks

import (
	"encoding/base64"

	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

func (ec *eks) GetEksClientSet() (*kubernetes.Clientset, error) {

	resultDescribe, err := ec.awsClient.DescribeCluster(ec.ctx, &awseks.DescribeClusterInput{
		Name: &ec.environment.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name,
	})
	if err != nil {
		return nil, err
	}

	restConfig, err := newRestConfig(resultDescribe.Cluster)
	if err != nil {
		return nil, err
	}

	clientset, err := makeKubeClientSet(*restConfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func newRestConfig(cluster *types.Cluster) (*rest.Config, error) {

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
	ca, err := base64.StdEncoding.DecodeString(aws.ToString(cluster.CertificateAuthority.Data))
	if err != nil {
		return nil, err
	}

	restConfig := &rest.Config{
		Host:        *cluster.Endpoint,
		BearerToken: tok.Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: ca,
		},
	}

	return restConfig, nil
}

func makeKubeClientSet(restConfig rest.Config) (*kubernetes.Clientset, error) {
	clientset, err := kubernetes.NewForConfig(&restConfig)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
