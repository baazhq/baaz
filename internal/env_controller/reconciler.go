package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"os"

	"datainfra.io/ballastdata/pkg/aws/eks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	v1 "datainfra.io/ballastdata/api/v1"
	"github.com/datainfrahq/operator-builder/builder"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func reconcileEnvironment(client client.Client, env *v1.Environment, recorder record.EventRecorder) error {

	os.Setenv("AWS_ACCESS_KEY_ID", env.Spec.CloudInfra.Auth.AwsAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", env.Spec.CloudInfra.Auth.AwsSecretAccessKey)

	err := CreateEnvironment(context.TODO(), env, client, recorder)
	if err != nil {
		return err
	}

	return nil
}

func CreateEnvironment(ctx context.Context, env *v1.Environment, client client.Client, record record.EventRecorder) error {

	eks := eks.EksEnvironment{
		Env:    env,
		Config: *eks.NewConfig(env.Spec.CloudInfra.AwsRegion),
		Client: client,
	}
	getOwnerRef := makeOwnerRef(
		env.APIVersion,
		env.Kind,
		env.Name,
		env.UID,
	)

	cm := makeEnvConfigMap(env, client, getOwnerRef, env.Spec)

	build := builder.NewBuilder(
		builder.ToNewBuilderConfigMap([]builder.BuilderConfigMap{*cm}),
		builder.ToNewBuilderRecorder(builder.BuilderRecorder{Recorder: record, ControllerName: "envoperator"}),
		builder.ToNewBuilderContext(builder.BuilderContext{Context: ctx}),
		builder.ToNewBuilderStore(
			*builder.NewStore(client, map[string]string{"app": env.Name}, env.Namespace, env),
		),
	)

	resp, err := build.ReconcileConfigMap()
	if err != nil {
		return err
	}

	fmt.Println(resp)

	if resp == controllerutil.OperationResultCreated {
		fmt.Println("creating eks")
		output := eks.CreateEks()
		fmt.Println(output)
	} else if resp == controllerutil.OperationResultUpdated {
		fmt.Println("updating eks")
		output := eks.UpdateEks()
		fmt.Println(output)
	}

	return nil
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
