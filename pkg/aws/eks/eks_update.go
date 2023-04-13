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
		return EksOutput{Result: err.Error()}

	}
	return EksOutput{Result: ClusterVersionUpradeInitated}

}

func (ke *EksEnvironment) updateEks(errorChan chan<- error) eks.UpdateClusterVersionOutput {
	eksClient := eks.NewFromConfig(ke.Config)

	output, err := eksClient.UpdateClusterVersion(context.TODO(), &eks.UpdateClusterVersionInput{
		Name:    &ke.Env.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name,
		Version: aws.String(ke.Env.Spec.CloudInfra.Eks.Version),
	})

	if err != nil {
		errorChan <- err
	}

	return *output

}
