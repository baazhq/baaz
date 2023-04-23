package eks

import (
	"context"
	"log"

	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "datainfra.io/ballastdata/api/v1"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type EksEnvironment struct {
	Client client.Client
	Env    *v1.Environment
	Config aws.Config
}

const (
	ClusterLaunchInitated        = "cluster launch initated"
	ClusterVersionUpradeInitated = "cluster version upgrade initated"
)

type EksOutput struct {
	Result     string
	Properties map[string]string
	Success    bool
}

func NewConfig(awsRegion string) *aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &cfg
}
