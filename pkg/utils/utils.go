package utils

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VerbType string

type (
	TransformStatusFunc func(obj client.Object) client.Object
)

const (
	VerbPatched   VerbType = "Patched"
	VerbUnchanged VerbType = "Unchanged"
)

func PatchStatus(ctx context.Context, c client.Client, obj client.Object, transform TransformStatusFunc, opts ...client.SubResourcePatchOption) (client.Object, VerbType, error) {
	key := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
	err := c.Get(ctx, key, obj)
	if err != nil {
		return nil, VerbUnchanged, err
	}

	// The body of the request was in an unknown format -
	// accepted media types include:
	//   - application/json-patch+json,
	//   - application/merge-patch+json,
	//   - application/apply-patch+yaml
	patch := client.MergeFrom(obj)
	obj = transform(obj.DeepCopyObject().(client.Object))
	err = c.Status().Patch(ctx, obj, patch, opts...)
	if err != nil {
		return nil, VerbUnchanged, err
	}
	return obj, VerbPatched, nil
}

func MakeKubeClientSet(restConfig rest.Config) (*kubernetes.Clientset, error) {
	clientset, err := kubernetes.NewForConfig(&restConfig)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

type EnvVars struct {
}

func SetEnvVars() *EnvVars {
	return &EnvVars{}
}
