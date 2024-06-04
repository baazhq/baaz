package predicates

import (
	"context"

	core "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// GenericSaaSPredicates to be passed to manager
type GenericSaaSPredicates struct {
	predicate.Funcs
	C client.Client
}

// create() to filter create events
func (g GenericSaaSPredicates) Create(e event.CreateEvent) bool {
	return IgnorePrivateSaaSCustomer(e.Object, g.C)
}

// update() to filter update events
func (g GenericSaaSPredicates) Update(e event.UpdateEvent) bool {
	return IgnorePrivateSaaSCustomer(e.ObjectNew, g.C)
}

// delete() to filter delete events
func (g GenericSaaSPredicates) Delete(e event.DeleteEvent) bool {
	return IgnorePrivateSaaSCustomer(e.Object, g.C)
}

func IgnorePrivateSaaSCustomer(obj client.Object, c client.Client) bool {
	ns := &core.Namespace{}

	if err := c.Get(context.TODO(), client.ObjectKey{Name: obj.GetNamespace()}, ns); err != nil {
		klog.Error(err)
		return false
	}

	return ns.Labels["private_mode"] != "true"
}
