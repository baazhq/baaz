package tenant_controller

// func (c *cloud) tenant_expansion() error {
// 	currentTenantObj := &v1.Tenants{}
// 	err := c.ae.client.Get(c.ae.ctx, types.NamespacedName{
// 		Namespace: c.ae.tenant.Namespace,
// 		Name:      c.ae.tenant.Name,
// 	}, currentTenantObj)
// 	if err != nil {
// 		return err
// 	}

// 	// for _, curentTenantSize := range currentTenantObj.Spec.TenantSizes {
// 	// 	for _, desiredTenantSize := range c.ae.tenant.Spec.TenantSizes {
// 	// 		for _, currentNodeSpec := range curentTenantSize.Spec {
// 	// 			for _, desiredNodeSpec := range desiredTenantSize.Spec {
// 	// 				if currentNodeSpec.Size != desiredNodeSpec.Size {
// 	// 					klog.Infof(
// 	// 						"Tenant Expansion: Tenant [%s], Node Name [%s], Current Size [%s], Desired Size [%s]",
// 	// 						currentTenantObj.Name,
// 	// 						currentNodeSpec.Name,
// 	// 						currentNodeSpec.Size,
// 	// 						desiredNodeSpec.Size,
// 	// 					)

// 	// 					var nodeName string
// 	// 					for _, tenantConfig := range c.ae.tenant.Spec.TenantConfig {
// 	// 						if tenantConfig.Size == desiredNodeSpec.Name {
// 	// 							nodeName = makeNodeName(desiredNodeSpec.Name, string(tenantConfig.AppType), tenantConfig.Size)

// 	// 						}
// 	// 					}

// 	// 					describeNodegroupOutput, found, _ := c.ae.eksIC.DescribeNodegroup(nodeName)
// 	// 					if !found {
// 	// 						nodeRole, err := c.ae.eksIC.CreateNodeIamRole(nodeName)
// 	// 						if err != nil {
// 	// 							return err
// 	// 						}
// 	// 						if nodeRole.Role == nil {
// 	// 							return errors.New("node role is nil")
// 	// 						}

// 	// 						createNodeGroupOutput, err := c.ae.eksIC.CreateNodegroup(c.ae.getNodegroupInput(nodeName, *nodeRole.Role.Arn, &nodeSpec))
// 	// 						if err != nil {
// 	// 							return err
// 	// 						}
// 	// 						if createNodeGroupOutput != nil && createNodeGroupOutput.Nodegroup != nil {
// 	// 							klog.Infof("Initated NodeGroup Launch [%s]", *createNodeGroupOutput.Nodegroup.ClusterName)
// 	// 							if err := ae.patchStatus(*createNodeGroupOutput.Nodegroup.NodegroupName, string(createNodeGroupOutput.Nodegroup.Status)); err != nil {
// 	// 								return err
// 	// 							}
// 	// 						}
// 	// 					}

// 	// 					c.ae.eksIC.CreateNodegroup()
// 	// 				}

// 	// 			}
// 	// 		}
// 	// 	}
// 	// }

// }
