package eks

import (
	"datainfra.io/ballastdata/pkg/store"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"k8s.io/client-go/rest"
)

type EksEnv interface {
	CreateEks() *EksOutput
	UpdateEks() *EksOutput
	DescribeEks() (*awseks.DescribeClusterOutput, error)
	DeleteEKS() (*awseks.DeleteClusterOutput, error)
	DeleteOIDCProvider(providerArn string) (*iam.DeleteOpenIDConnectProviderOutput, error)
	UpdateAwsEksEnvironment(clusterResult *awseks.DescribeClusterOutput) error
	ReconcileNodeGroup(store store.Store) error
	ReconcileOIDCProvider(clusterOutput *awseks.DescribeClusterOutput) error
	ReconcileDefaultAddons() error
	DeleteNodeGroup(nodeGroupName string) (*awseks.DeleteNodegroupOutput, error)
	DescribeNodeGroup(nodeGroupName string) (output *awseks.DescribeNodegroupOutput, found bool, err error)
	GetEksConfig() (*rest.Config, error)
}
