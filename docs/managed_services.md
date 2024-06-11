# Managed Services Infra SaaS

Welcome to the Managed Services Infra SaaS documentation!

A Managed Services infra SaaS is where the SaaS provider manages the applications, but they run on the customer's cloud infrastructure. The SaaS provider manages the application and its SLA.

## Table of Contents

1. [The need for Managed Services Infra SaaS](#the-need-for-managed-services-infra-saas)
2. [Architecture](#architecture)
3. [BaaZ Features for Managed Services Infra SaaS](#baaz-features-for-managed-services-infra-saas)
4. [Getting started building MS](#user-flow)

## The need for Managed Services Infra SaaS

#### Go-to-Market 

- **Monetization Opportunity:** Incorporating managed services into your infrastructure product provides a revenue stream by offering services alongside your product.
- **Lower Entry Barrier:** Managed services require less initial investment and development effort, making it easier for developers to get started.
- **Customizable Solutions:** Managed services can be tailored to meet the specific use cases and network configurations of individual customers, enhancing flexibility and addressing unique requirements.
- **Startup Advantage:** For startups, managed services offer a straightforward path to market entry, allowing them to quickly offer value-added services to customers.


#### Security

- **Data Sovereignty:** Maintain control over data location, ensuring compliance with local data privacy laws and regulations.
- **Data Security:** Protect sensitive data by keeping it within the customer's network.
- **Compliance:** Meet regulatory requirements by maintaining control over data location.
- **Confidentiality:** Ensure proprietary information stays within trusted network boundaries.
- **Control:** Customize and manage the SaaS environment according to specific needs.

## Architecture

![managed-services](https://github.com/baazhq/baaz/assets/34169002/ca216e2f-14e4-4431-945b-4af723514650)

## BaaZ Features for Managed Services Infra SaaS

- **Security-Focused Model:** Implemented as a pull-based system, prioritizing security by ensuring minimal external access to customer networks.
- **Network Localization:** BaaZ's control plane operates within the customer's network, enhancing data security by keeping critical operations within trusted boundaries.
- **Simplified Access:** Our controllers seamlessly synchronize with BaaZ's main control plane, offering ease of use without compromising security.
- **Controlled Interaction:** BaaZ's control plane maintains strict limitations on network access, ensuring that it doesn't interfere with or compromise customer network internals.
- **User-Friendly Operations:** All management and reconciliation operations are designed to occur within the customer's network environment, simplifying usage while upholding security standards.

## User Flow

We expect the customers to run the BaaZ control plane in their own network and provision infrastructure. At the same time, as a backend engineer maintaining various SaaS infra dataplanes, I would still be managing infra from my network, without switching between dataplanes or accessing customer networks.

### Service Account

- For every customer we create on BaaZ control plane, we create a service account for it.
- For each service account, we create a kubeconfig file scoped to that namespace only.
- Create appropriate RBAC to scope SA to reconcile only dataplanes, tenants, tenantsinfra, and applications.

### Run locally in the same cluster

#### Run BaaZ Control Plane in Standard Mode
- Set local env
```sh
source local.sh
```
