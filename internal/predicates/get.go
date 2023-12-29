package predicates

import (
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func GetPredicates(enablePrivate bool, customerName string) predicate.Predicate {
	if enablePrivate {
		return PrivateSaaSPredicates{CustomerName: customerName}
	}
	return GenericSaaSPredicates{}
}
