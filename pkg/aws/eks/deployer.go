package eks

import (
	"encoding/base64"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/helm"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go/aws"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

// druid
const (
	druidControlPlaneReleaseName string = "druid-control-plane"
	druidControlPlaneNamespace   string = "druid-control-plane"
	druidControlPlaneChartName   string = "druid-operator"
	repoName                     string = "datainfra"
	repoUrl                      string = "https://charts.datainfra.io"
)

// Deployer is responsible for deploying apps
func (eksEnv *EksEnvironment) ReconcileDeployer() error {
	restConfig, err := eksEnv.getEksConfig()
	if err != nil {
		return err
	}

	for _, app := range eksEnv.Env.Spec.Tenant {

		switch app.AppType {
		case v1.Druid:
			// deploy druid operator
			druidOperatorHelm := helm.NewHelm(
				druidControlPlaneReleaseName,
				druidControlPlaneNamespace,
				druidControlPlaneChartName,
				repoName,
				repoUrl,
				nil)
			err := druidOperatorHelm.HelmList(restConfig)
			if err != nil {
				return nil
			}

			err = druidOperatorHelm.HelmInstall(restConfig)
			if err != nil {
				return err
			}

		}
	}
	return nil
}

func (eksEnv *EksEnvironment) getEksConfig() (*rest.Config, error) {
	eksClient := awseks.NewFromConfig(eksEnv.Config)

	resultDescribe, err := eksClient.DescribeCluster(eksEnv.Context, &awseks.DescribeClusterInput{
		Name: &eksEnv.Env.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name,
	})
	if err != nil {
		return nil, err
	}

	return newRestConfig(resultDescribe.Cluster)

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
	ca, err := base64.StdEncoding.DecodeString(aws.StringValue(cluster.CertificateAuthority.Data))
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
