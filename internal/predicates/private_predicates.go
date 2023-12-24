package predicates

import (
	"fmt"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// PrivateSaaSPredicates to be passed to manager
type PrivateSaaSPredicates struct {
	predicate.Funcs
	CustomerName string
}

// create() to filter create events
func (p PrivateSaaSPredicates) Create(e event.CreateEvent) bool {
	return IgnoreGenericSaaSCustomer(e.Object, p.CustomerName)
}

// update() to filter update events
func (p PrivateSaaSPredicates) Update(e event.UpdateEvent) bool {
	return IgnoreGenericSaaSCustomer(e.ObjectNew, p.CustomerName)
}

// delete() to filter delete events
func (p PrivateSaaSPredicates) Delete(e event.DeleteEvent) bool {
	return IgnoreGenericSaaSCustomer(e.Object, p.CustomerName)
}

func IgnoreGenericSaaSCustomer(obj client.Object, customerName string) bool {

	if obj.GetNamespace() != customerName {
		msg := fmt.Sprintf("baaz controllers will only renconcile %s", customerName)
		klog.Info(msg)
		return false
	}

	return true
}
