package resources

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeTenantNamespace(
	name, namespace string,
	ownerRef *metav1.OwnerReference,
) *v1.Namespace {

	networkPolicy := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			OwnerReferences: []metav1.OwnerReference{*ownerRef},
			Labels: map[string]string{
				"tenant": name,
			},
		},
		Spec: v1.NamespaceSpec{},
	}

	return networkPolicy
}
