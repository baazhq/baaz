# Errors
- Config map update doesn't mean eks update. Possible failure scenario: Create config map -> failed to create eks -> update crd -> config map update -> panic on failed to find cluster

# Improvements
- Spec should not contain any auth. We can handle auth via Kubernetes Secret and use the ref in the spec.
- We should not use os.SetEnv() for aws auth. It can cause issue for multiple cr and one controllers.
- We don't really need config map to store crd values. We can filter out reconciler condition to make sure to reconcile for only spec/specific field updates.



