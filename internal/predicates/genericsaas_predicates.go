package predicates

import (
	"fmt"

	v1 "datainfra.io/baaz/api/v1/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// GenericSaaSPredicates to be passed to manager
type GenericSaaSPredicates struct {
	predicate.Funcs
}

// create() to filter create events
func (GenericSaaSPredicates) Create(e event.CreateEvent) bool {
	return IgnorePrivateSaaSCustomer(e.Object)
}

// update() to filter update events
func (GenericSaaSPredicates) Update(e event.UpdateEvent) bool {
	return IgnorePrivateSaaSCustomer(e.ObjectNew)
}

// delete() to filter delete events
func (GenericSaaSPredicates) Delete(e event.DeleteEvent) bool {
	return IgnorePrivateSaaSCustomer(e.Object)
}

func IgnorePrivateSaaSCustomer(obj client.Object) bool {

	if obj.GetLabels()["saas_type"] != string(v1.PrivateSaaS) {
		msg := fmt.Sprintf("baaz controllers will not renconcile private saas %s", obj.GetNamespace())
		klog.Info(msg)
		return false
	}

	return true
}
