package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/aws/eks"
	"github.com/datainfrahq/operator-builder/builder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func reconcileEnvironment(ctx context.Context, client client.Client, env *v1.Environment, recorder record.EventRecorder) error {
	os.Setenv("AWS_ACCESS_KEY_ID", env.Spec.CloudInfra.Auth.AwsAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", env.Spec.CloudInfra.Auth.AwsSecretAccessKey)

	return createOrUpdateEnvironment(ctx, env, client, recorder)
}

func createOrUpdateEnvironment(ctx context.Context, env *v1.Environment, c client.Client, record record.EventRecorder) error {
	eksEnv := eks.EksEnvironment{
		Env:    env,
		Config: *eks.NewConfig(env.Spec.CloudInfra.AwsRegion),
		Client: c,
	}

	if err := eks.DescribeCluster(ctx, eksEnv); err != nil {
		klog.Info("Updating Environment status to creating")
		if _, _, err := PatchStatus(ctx, c, env, func(obj client.Object) client.Object {
			in := obj.(*v1.Environment)
			in.Status.Phase = v1.Creating
			return in
		}); err != nil {
			return err
		}
		if err := createEnvironment(eksEnv); err != nil {
			return err
		}
		klog.Info("Successfully created kubernetes control plane")
	}
	return updateEnvironment(eksEnv)
}

func updateEnvironment(eksEnv eks.EksEnvironment) error {
	// TODO: Update environment
	fmt.Println("Todo: Updating environment")
	return nil
}

func createEnvironment(eksEnv eks.EksEnvironment) error {
	output := eksEnv.CreateEks()
	if output.Success {
		return nil
	}
	return errors.New(output.Result)
}

func makeEnvConfigMap(
	env *v1.Environment,
	client client.Client,
	ownerRef *metav1.OwnerReference,
	data interface{},
) *builder.BuilderConfigMap {

	dataByte, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	configMap := &builder.BuilderConfigMap{
		CommonBuilder: builder.CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      env.GetName(),
				Namespace: env.GetNamespace(),
			},
			Client:   client,
			CrObject: env,
			OwnerRef: *ownerRef,
		},
		Data: map[string]string{
			"data": string(dataByte),
		},
	}

	return configMap
}

// create owner ref ie parseable tenant controller
func makeOwnerRef(apiVersion, kind, name string, uid types.UID) *metav1.OwnerReference {
	controller := true

	return &metav1.OwnerReference{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       name,
		UID:        uid,
		Controller: &controller,
	}
}
