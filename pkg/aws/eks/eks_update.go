package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

func (ke *EksEnvironment) UpdateEks() EksOutput {

	errChannel := make(chan error)

	go ke.updateEks(errChannel)

	for err := range errChannel {
		if err != nil {
			return EksOutput{Result: err.Error()}
		}
		break
	}
	return EksOutput{
		Result:  ClusterVersionUpradeInitated,
		Success: true,
	}
}

func (ke *EksEnvironment) updateEks(errorChan chan<- error) {
	eksClient := eks.NewFromConfig(ke.Config)

	_, err := eksClient.UpdateClusterVersion(context.TODO(), &eks.UpdateClusterVersionInput{
		Name:    &ke.Env.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name,
		Version: aws.String(ke.Env.Spec.CloudInfra.Eks.Version),
	})

	if err != nil {
		errorChan <- err
	}
}
