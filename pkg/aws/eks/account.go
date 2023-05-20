package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func (eksEnv *EksEnvironment) getAccountID(ctx context.Context) (string, error) {
	stsClient := sts.NewFromConfig(eksEnv.Config)

	result, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	return *result.Account, nil
}
