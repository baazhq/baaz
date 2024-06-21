## Private-SaaS Flow

### Cloud: AWS

1. git clone the baaz server locally

```bash
$ git clone git@github.com:baazhq/baaz.git
$ cd baaz
```

2. Create a EKS cluster from AWS UI to use it as the Provider Cluster

3. Install baaz chart into the Provider Cluster with the service type as LoadBalancer.

```bash
$ cd chart/baaz
$ helm install baaz . --namespace=baaz --create-namespace \
            --set service.type=LoadBalancer \
            --set image.repository=pkbhowmick/baaz \
            --set image.tag=latest
```

Note: Edit baaz deployment and add the `KUBERNETES_CONFIG_SERVER_URL` as the EKS host

4. Create a Kind Cluster for the provider cluster

```bash
$ kind create cluster
```

5. Export the Baaz URL and create the customer

```bash
$ export BAAZ_URL=http://<LB_URL>:8000
```

Sample customer yaml

```yaml
customer:
  name: foo
  saas_type: private
  cloud_type: aws
  labels: 
    tier: business
    region: us-east-1
```

```bash
$ bz create customer --private_mode=true -f <yaml_file_path>
```

6. Run baaz init with the cloud auth

```bash
$ bz init --private_mode=true --customer=foo --aws_access_key=<aws_access_key> \
                       --aws_secret_key=<aws_secret_key>
```


7. Create dataplane in private saas

Sample dataplane yaml

```yaml
dataplane:
  cloudType: aws
  cloudRegion: us-east-1
  saasType: private
  customerName: foo
  provisionNetwork: true
  kubernetesConfig:
    eks:
      version: '1.27'
  applicationConfig:
  - name: "nginx"
    namespace: nginx-ingress
    chartName: "ingress-nginx"
    repoName: "ingress-nginx"
    repoUrl: "https://kubernetes.github.io/ingress-nginx"
    version: "1.9.4"
    values:
    - controller.nodeSelector.nodeType=system
```

```bash
$ bz create dataplane -f <yaml_file_path>
```

