package predicates

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func GetPredicates(enablePrivate bool, customerName string, c client.Client) predicate.Predicate {

	if enablePrivate {
		return PrivateSaaSPredicates{CustomerName: customerName}
	}
	return GenericSaaSPredicates{C: c}
}
