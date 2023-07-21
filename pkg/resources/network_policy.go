package resources

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeNetworkPolicy(
	name, namespace string,
	allowNamespaces []string,
	ownerRef *metav1.OwnerReference,
) *networkingv1.NetworkPolicy {

	var peerPolicy []networkingv1.NetworkPolicyPeer

	for _, allowNamespace := range allowNamespaces {
		peerPolicy = append(peerPolicy, networkingv1.NetworkPolicyPeer{
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": allowNamespace,
				},
			},
		})
	}

	networkPolicy := &networkingv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "NetworkPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			OwnerReferences: []metav1.OwnerReference{*ownerRef},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: peerPolicy,
				},
			},
		},
	}

	return networkPolicy
}
