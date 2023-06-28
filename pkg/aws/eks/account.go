package eks

import "github.com/aws/aws-sdk-go-v2/service/sts"

func (ec *eks) getAccountID() (string, error) {

	result, err := ec.awsStsClient.GetCallerIdentity(ec.ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	return *result.Account, nil

}
