package application

// import (
// 	"context"
// 	"encoding/base64"

// 	v1 "datainfra.io/ballastdata/api/v1"
// 	"datainfra.io/ballastdata/pkg/aws/eks"
// 	"datainfra.io/ballastdata/pkg/deployer"
// 	"github.com/aws/aws-sdk-go-v2/aws"
// 	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
// 	"github.com/aws/aws-sdk-go-v2/service/eks/types"
// 	k8stypes "k8s.io/apimachinery/pkg/types"
// 	"k8s.io/client-go/rest"
// 	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// )

// type Application struct {
// 	Context context.Context
// 	Client  client.Client
// 	App     *v1.Application
// }

// func NewApplication(
// 	ctx context.Context,
// 	client client.Client,
// 	app *v1.Application,
// ) App {
// 	return &Application{
// 		Context: ctx,
// 		Client:  client,
// 		App:     app,
// 	}
// }

// // Deployer is responsible for deploying apps
// func (app *Application) ReconcileApplicationDeployer() error {

// 	envObj := &v1.Environment{}
// 	err := app.Client.Get(app.Context, k8stypes.NamespacedName{Name: app.App.Spec.EnvRef, Namespace: app.App.Namespace}, envObj)
// 	if err != nil {
// 		return err
// 	}

// 	restConfig, err := app.GetEksConfig(envObj.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name, envObj.Spec.CloudInfra.AwsRegion)
// 	if err != nil {
// 		return err
// 	}

// 	deploy := deployer.NewDeployer(restConfig, envObj, app.App)

// 	if err := deploy.ReconcileDeployer(); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (app *Application) GetEksConfig(name, region string) (*rest.Config, error) {
// 	//eksClient := awseks.NewFromConfig(*eks.NewConfig(region))

// 	resultDescribe, err := eksClient.DescribeCluster(app.Context, &awseks.DescribeClusterInput{
// 		Name: &name,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return newRestConfig(resultDescribe.Cluster)

// }

// func newRestConfig(cluster *types.Cluster) (*rest.Config, error) {

// 	gen, err := token.NewGenerator(true, false)
// 	if err != nil {
// 		return nil, err
// 	}
// 	opts := &token.GetTokenOptions{
// 		ClusterID: *cluster.Name,
// 	}
// 	tok, err := gen.GetWithOptions(opts)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ca, err := base64.StdEncoding.DecodeString(aws.ToString(cluster.CertificateAuthority.Data))
// 	if err != nil {
// 		return nil, err
// 	}

// 	restConfig := &rest.Config{
// 		Host:        *cluster.Endpoint,
// 		BearerToken: tok.Token,
// 		TLSClientConfig: rest.TLSClientConfig{
// 			CAData: ca,
// 		},
// 	}

// 	return restConfig, nil
// }
