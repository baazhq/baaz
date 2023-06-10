#!/bin/sh
set -e
EKS_NAME=$1
aws eks update-kubeconfig --name ${EKS_NAME} --dry-run > ${EKS_NAME}.yaml
export KUBECONFIG=${EKS_NAME}.yaml
if command -v kubectl &> /dev/null; then
if output=$(kubectl get daemonset -n kube-system 2>&1); then
if echo "$output" | grep -q "aws-node"; then
kubectl delete daemonset -n kube-system aws-node
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.0/manifests/tigera-operator.yaml
kubectl apply -f - <<EOF
kind: Installation
apiVersion: operator.tigera.io/v1
metadata:
  name: default
spec:
  kubernetesProvider: EKS
  cni:
    type: Calico
  calicoNetwork:
    bgp: Disabled
EOF
unset KUBECONFIG
rm -rf ${EKS_NAME}.yaml
else
    echo "DaemonSet does not exist in kube-system namespace."
fi
else
echo "Failed to execute 'kubectl get daemonset' command"
echo "$output"
fi
else
    echo "kubectl command not found."
fi

