package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/aws/eks"
	"datainfra.io/ballastdata/pkg/store"
	"datainfra.io/ballastdata/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var envFinalizer = "environment.datainfra.io/finalizer"

func (r *EnvironmentReconciler) createOrUpdateAwsEksEnvironment(ctx context.Context, env *v1.Environment) error {

	eksEnv := eks.NewEksEnvironment(ctx, r.Client, env, *eks.NewConfig(env.Spec.CloudInfra.AwsRegion))

	store := store.NewInternalStore()

	result, err := eksEnv.DescribeEks()

	if err != nil {

		var ngNotFound *types.ResourceNotFoundException
		if errors.As(err, &ngNotFound) {

			klog.Infof("Creating EKS Control plane: %s for Environment: %s/%s", env.Spec.CloudInfra.Eks.Name, env.Namespace, env.Name)
			klog.Info("Updating Environment status to creating")

			createEksResult := eksEnv.CreateEks()
			if createEksResult.Success {
				if _, _, err := utils.PatchStatus(ctx, r.Client, env, func(obj client.Object) client.Object {
					in := obj.(*v1.Environment)
					in.Status.Phase = v1.Creating
					in.Status.Conditions = in.AddCondition(v1.EnvironmentCondition{
						Type:               v1.ControlPlaneCreateInitiated,
						Status:             corev1.ConditionTrue,
						LastUpdateTime:     metav1.Time{Time: time.Now()},
						LastTransitionTime: metav1.Time{Time: time.Now()},
						Reason:             string(eks.EksControlPlaneInitatedReason),
						Message:            string(eks.EksControlPlaneInitatedMsg),
					})
					return in
				}); err != nil {
					return err
				}

				klog.Info("Successfully initiated kubernetes control plane")
				return nil

			} else {
				klog.Errorf(createEksResult.Result)
			}

		}
	}

	if result != nil {
		if err := eksEnv.UpdateAwsEksEnvironment(result); err != nil {
			return err
		}
	}
	if result != nil && result.Cluster != nil && result.Cluster.Status == eks.EKSStatusACTIVE {
		if err := eksEnv.ReconcileNodeGroup(store); err != nil {
			return err
		}
		if err := eksEnv.ReconcileOIDCProvider(result); err != nil {
			return err
		}
		if err := eksEnv.ReconcileDefaultAddons(); err != nil {
			return err
		}
		if err := eksEnv.ReconcileDeployer(); err != nil {
			return err
		}
	}

	if env.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(env, envFinalizer) {
			controllerutil.AddFinalizer(env, envFinalizer)
			if err := r.Update(ctx, env.DeepCopyObject().(*v1.Environment)); err != nil {
				return nil
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(env, envFinalizer) {
			ngList := store.GetNodeGroups(env.Spec.CloudInfra.Eks.Name)

			for _, ng := range ngList {
				if env.Status.NodegroupStatus[ng] != "DELETING" {
					_, err := eksEnv.DeleteNodeGroup(ng)
					if err != nil {
						return err
					}
					// update status with current nodegroup status
					_, _, err = utils.PatchStatus(ctx, r.Client, env, func(obj client.Object) client.Object {
						in := obj.(*v1.Environment)
						if in.Status.NodegroupStatus == nil {
							in.Status.NodegroupStatus = make(map[string]string)
						}
						in.Status.NodegroupStatus[ng] = "DELETING"
						return in
					})
				}

			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(env, envFinalizer)
			if err := r.Update(ctx, env.DeepCopyObject().(*v1.Environment)); err != nil {
				return nil
			}
		}
	}

	fmt.Println(store.GetNodeGroups(env.Spec.CloudInfra.Eks.Name))
	return nil
}
