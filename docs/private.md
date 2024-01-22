# Private Infra SaaS

A private infra saas is where we run our SaaS in the customer network. Private SaaS can also be called BYOC ie bring your own cloud.

## System Design
Model: Pull based model
BaaZ control plane runs its 4 controllers in the customer network. The controllers further connect with baaz control plane and reconcile the same cluster. The only catch here is they reconcile from customer network. BaaZ control plane is no where aware of the customer network internals nor has any access to it. 

## User Flow

We expect the customers to run the BaaZ control plane in there own network, and provision infrastructure. At the same time as an backend engineer maintaining various SaaS infra dataplanes, i would still be managing infra from my network, without switching between dataplanes or accessing customer networks.

### Service Account

- For every customer we create on BaaZ control plane, we create a service account for it.
- For each service account we create a kubeconfig file scoped to that namespace only.
- Create appropriate RBAC to scope SA to reconcile only dataplanes, tenants, tenantsinfra and applications.

### Run locally in same cluster

# Run BaaZ Control Plane in Standard Mode
- Set local env
```
source local.sh
```
- Run BaaZ HTTP server with controllers
```
make run
```

# Run BaaZ Control Plane in Private Mode

- create customer with ```saas_type: private```
```
cat << EOF > customer.yaml
customer:
  name: foo
  saas_type: "private"
  cloud_type: "aws"
  labels: 
    tier: "enterprise"
EOF
```

```
bz create customer -f customer.yaml
```

- create kubeconfig for customer
```
curl --location 'http://localhost:8000/api/v1/customer/foo/config'
```

- construct kubeconfig and save it in a file

```
go run cmd/main.go -kubeconfig=hack/kind -private_mode=true -customer_name=foo 
```