Kibble
=================

Kibble pulls metrics from [Prometheus Query API](https://prometheus.io/docs/prometheus/latest/querying/api/), transform and submit to Datadog

## Limitations

Kibble only supports Temporal Cloud metrics at the moment.

# Running Kibble locally

## Prerequisites

* Go 1.20+
* A Datadog API key exported as `DD_API_KEY` in your shell
* You have [configured your Temporal account with CA certificate](https://docs.temporal.io/cloud/how-to-monitor-temporal-cloud-metrics)

## Build Kibble

```
make
```

## Running Kibble

```
./kibble \
  --prom-endpoint https://<temporal-account-id>.tmprl.cloud/prometheus \
  --client-cert <replace with the path to CA cert> \
  --client-key <replace with the path to CA key>
```

# Install Kibble on a Kubernetes cluster

## Prerequisites

In addition to the Datadog API key and Temporal Cloud CA certs, you have:
* the system configured to access a Kubernetes cluster (e. g. [AWS EKS](https://aws.amazon.com/eks/), [kind](https://kind.sigs.k8s.io/), or [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)), and
* your machine below clis installed:
  - [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/), and
  - [Helm v3](https://helm.sh)

## Install Kibble with Helm Chart

Run below command from the `helm-charts` folder.

```
helm install kibble . \
  --set prom_endpoint=https://<temporal-account-id>.tmprl.cloud/prometheus \
  --set dd_api_key=${DD_API_KEY} \
  --set-file 'ca_cert=<replace with the path to CA cert>' \
  --set-file 'ca_key=<replace with the path to CA key>'
```

## Verify Kibble is running

```
kubectl logs $(kubectl get pods | grep kibble | awk '{print $1; exit}') -f
```

Should output:

```
2023/06/20 11:33:43 Querying Prometheus
2023/06/20 11:33:43 Found 1 histogram metrics: [temporal_cloud_v0_service_latency_bucket]
...
```

## License
MIT License, please see [LICENSE](LICENSE) for details.
