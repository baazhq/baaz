# Managed Services Infra SaaS

A Managed Services infra saas is where the SaaS provider manages the applications, but they run on the customer's cloud infrastructure. Applications deployed our hosted at the customer cloud infrastructure. The SaaS provider manages the application and its SLA.

# The need for Managed Services Infra SaaS

- ***Data Sovereignty:*** Maintain control over data location, ensuring compliance with local data privacy laws and regulations.

- ***Data Security:*** Protect sensitive data by keeping it within the customer's network.
- ***Compliance:*** Meet regulatory requirements by maintaining control over data location.
- ***Confidentiality:*** Ensure proprietary information stays within trusted network boundaries.
- ***Control:*** Customize and manage the SaaS environment according to specific needs.

## BaaZ Features for Managed Services Infra SaaS

- **Security-Focused Model:** Implemented as a **pull-based** system, prioritizing security by ensuring minimal external access to customer networks.

- **Network Localization:** BaaZ's control plane operates within the **customer's network**, enhancing data security by keeping critical operations within trusted boundaries.

- **Simplified Access:** Our controllers seamlessly synchronize with BaaZ's main control plane, offering ease of use without compromising security.

- **Controlled Interaction:** BaaZ's control plane maintains strict limitations on network access, ensuring that it doesn't interfere with or compromise customer network internals.

- **User-Friendly Operations:** All management and reconciliation operations are designed to occur within the **customer's network** environment, simplifying usage while upholding security standards.


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