package eks

import (
	"encoding/base64"

	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

func (ec *eks) GetEksClientSet() (*kubernetes.Clientset, error) {

	restConfig, err := ec.GetRestConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := makeKubeClientSet(*restConfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func (ec *eks) GetRestConfig() (*rest.Config, error) {

	resultDescribe, err := ec.awsClient.DescribeCluster(ec.ctx, &awseks.DescribeClusterInput{
		Name: &ec.environment.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name,
	})
	if err != nil {
		return nil, err
	}

	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return nil, err
	}
	opts := &token.GetTokenOptions{
		ClusterID: *resultDescribe.Cluster.Name,
	}
	tok, err := gen.GetWithOptions(opts)
	if err != nil {
		return nil, err
	}
	ca, err := base64.StdEncoding.DecodeString(aws.ToString(resultDescribe.Cluster.CertificateAuthority.Data))
	if err != nil {
		return nil, err
	}

	restConfig := &rest.Config{
		Host:        *resultDescribe.Cluster.Endpoint,
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
