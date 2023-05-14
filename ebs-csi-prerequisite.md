## ebs-csi-driver role prerequisite

```bash
$ EKS_CLUSTER_NAME="dev-eks"
# for below command make sure aws cli is already configured
$ ACCOUNT_ID=$(aws sts get-caller-identity \
  --query "Account" --output text)

# oidc provider is already added from ballasdata operator
$ OIDC_PROVIDER=$(aws eks describe-cluster --name $EKS_CLUSTER_NAME \
  --query "cluster.identity.oidc.issuer" --output text | sed -e 's|^https://||')

# in trust.json we have to make sure that aws-ebs-csi-driver controller will be running in the
# kube-system namespace with service account named "aws-ebs-csi-driver"
cat <<-EOF > trust.json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Federated": "arn:aws:iam::$ACCOUNT_ID:oidc-provider/$OIDC_PROVIDER"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "$OIDC_PROVIDER:sub": "system:serviceaccount:kube-system:aws-ebs-csi-driver", 
                    "$OIDC_PROVIDER:aud": "sts.amazonaws.com"
                }
            }
        }
    ]
}
EOF

IRSA_ROLE="ebs-csi-irsa-role"
aws iam create-role --role-name $IRSA_ROLE --assume-role-policy-document file://trust.json

# attach policy with the role 
# you can manually attach `AmazonEBSCSIDriverPolicy` policy with the Role
$ POLICY_ARN="arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
$ aws iam attach-role-policy --role-name $IRSA_ROLE --policy-arn $POLICY_ARN
```

Note: For coredns we can use same role that is used for eks. For that, it seems we don't need any irsa.