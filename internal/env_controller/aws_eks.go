package controller

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"datainfra.io/ballastdata/pkg/aws/iam"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/aws/eks"
	"datainfra.io/ballastdata/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createOrUpdateAwsEksEnvironment(ctx context.Context, env *v1.Environment, c client.Client, record record.EventRecorder) error {

	// init eks environment
	eksEnv := eks.NewEksEnvironment(ctx, c, env, *eks.NewConfig(env.Spec.CloudInfra.AwsRegion))

	result, err := eksEnv.DescribeEks()

	if err != nil {

		var ngNotFound *types.ResourceNotFoundException
		if errors.As(err, &ngNotFound) {

			klog.Infof("Creating EKS Control plane: %s for Environment: %s/%s", env.Spec.CloudInfra.Eks.Name, env.Namespace, env.Name)
			klog.Info("Updating Environment status to creating")

			createEksResult := eksEnv.CreateEks()
			if createEksResult.Success {
				if _, _, err := utils.PatchStatus(ctx, c, env, func(obj client.Object) client.Object {
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
	if result != nil && result.Result != nil && result.Result.Cluster != nil && result.Result.Cluster.Status == eks.EKSStatusACTIVE {
		if err := reconcileNodeGroup(ctx, eksEnv); err != nil {
			return err
		}
		if err := reconcileOIDCProvider(ctx, eksEnv, result); err != nil {
			return err
		}
		return reconcileDefaultAddons(ctx, eksEnv)
	}
	return nil
}

func reconcileOIDCProvider(ctx context.Context, eksEnv *eks.EksEnvironment, clusterOutput *eks.DescribeClusterOutput) error {
	if clusterOutput == nil || clusterOutput.Result == nil || clusterOutput.Result.Cluster == nil ||
		clusterOutput.Result.Cluster.Identity == nil || clusterOutput.Result.Cluster.Identity.Oidc == nil {
		return errors.New("oidc provider url not found in cluster output")
	}
	oidcProviderUrl := *clusterOutput.Result.Cluster.Identity.Oidc.Issuer

	// Compute the SHA-1 thumbprint of the OIDC provider certificate
	thumbprintBytes := sha1.Sum([]byte(oidcProviderUrl))
	thumbprint := hex.EncodeToString(thumbprintBytes[:])

	input := &iam.CreateOIDCProviderInput{
		URL:            oidcProviderUrl,
		ThumbPrintList: []string{thumbprint},
	}

	oidcProviderArn := eksEnv.Env.Status.CloudInfraStatus.EksStatus.OIDCProviderArn

	if oidcProviderArn != "" {
		// oidc provider is previously created
		// looking for it
		providers, err := iam.ListOIDCProvider(ctx, eksEnv)
		if err != nil {
			return err
		}

		for _, oidc := range providers.Result.OpenIDConnectProviderList {
			if *oidc.Arn == eksEnv.Env.Status.CloudInfraStatus.EksStatus.OIDCProviderArn {
				// oidc provider is already created and existed
				return nil
			}
		}
	}

	result, err := iam.CreateOIDCProvider(ctx, eksEnv, input)
	if err != nil {
		return err
	}
	fmt.Println("OIDC provider create initiated")
	_, _, err = utils.PatchStatus(ctx, eksEnv.Client, eksEnv.Env, func(obj client.Object) client.Object {
		in := obj.(*v1.Environment)
		in.Status.CloudInfraStatus.EksStatus.OIDCProviderArn = *result.Result.OpenIDConnectProviderArn

		return in
	})
	return err
}

func reconcileDefaultAddons(ctx context.Context, eksEnv *eks.EksEnvironment) error {
	clusterName := eksEnv.Env.Spec.CloudInfra.Eks.Name
	ebsAddon, err := eksEnv.DescribeAddon(ctx, "aws-ebs-csi-driver", eksEnv.Env.Spec.CloudInfra.Eks.Name)
	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			klog.Info("Creating aws-ebs-csi-driver addon")
			_, cErr := eksEnv.CreateAddon(ctx, &eks.CreateAddonInput{
				Name:        "aws-ebs-csi-driver",
				ClusterName: clusterName,
				RoleArn:     "arn:aws:iam::437639712640:role/ebs-sa-role",
			})
			if cErr != nil {
				return cErr
			}
			klog.Info("aws-ebs-csi-driver addon creation is initiated")
		} else {
			return err
		}
		return nil
	}
	if ebsAddon.Result != nil && ebsAddon.Result.Addon != nil {
		klog.Info("aws-ebs-csi-driver addon status: ", ebsAddon.Result.Addon.Status)
	}

	coreDnsAddon, err := eksEnv.DescribeAddon(ctx, "coredns", eksEnv.Env.Spec.CloudInfra.Eks.Name)
	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			klog.Info("Creating coredns addon")
			_, cErr := eksEnv.CreateAddon(ctx, &eks.CreateAddonInput{
				Name:        "coredns",
				ClusterName: clusterName,
				RoleArn:     eksEnv.Env.Spec.CloudInfra.Eks.RoleArn,
			})
			if cErr != nil {
				return cErr
			}
			klog.Info("coredns addon creation is initiated")
		} else {
			return err
		}
		return nil
	}
	if coreDnsAddon != nil && coreDnsAddon.Result != nil && coreDnsAddon.Result.Addon != nil {
		klog.Info("coredns addon status: ", coreDnsAddon.Result.Addon.Status)
	}

	return nil
}

func reconcileNodeGroup(ctx context.Context, env *eks.EksEnvironment) error {

	for _, app := range env.Env.Spec.Application {

		nodeSpec, err := getNodegroupSpecForAppSize(env.Env, app)
		if err != nil {
			return err
		}

		ngs := eks.NewNodeGroup(ctx, env, &app, nodeSpec)

		_, err = ngs.CreateNodeGroupForApp()
		if err != nil {
			return err
		}

	}

	return nil
}

func updateStatusWithNodegroup(ctx context.Context, eksEnv *eks.EksEnvironment, nodegroup, status string) error {
	// update status with current nodegroup status
	_, _, err := utils.PatchStatus(ctx, eksEnv.Client, eksEnv.Env, func(obj client.Object) client.Object {
		in := obj.(*v1.Environment)
		if in.Status.NodegroupStatus == nil {
			in.Status.NodegroupStatus = make(map[string]string)
		}
		in.Status.NodegroupStatus[nodegroup] = status
		return in
	})
	return err
}

func syncNodegroup(ctx context.Context, eksEnv *eks.EksEnvironment) error {
	// update node group if spec node group is updated
	return nil
}

func getNodegroupSpecForAppSize(env *v1.Environment, app v1.ApplicationConfig) (*v1.NodeGroupSpec, error) {
	for _, size := range env.Spec.Size {
		if size.Name == app.Size && size.Spec.AppType == app.AppType {
			return size.Spec.Nodes, nil
		}
	}
	return nil, fmt.Errorf("no NodegroupSpec for app %s & size %s", app.Name, app.Size)
}
