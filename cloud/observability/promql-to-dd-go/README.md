PromQL to Datadog Go
=================

The goal of this Golang implementation is to demonstrate the minimum work necessary to read recently generated metrics from a Temporal Cloud account using the Prometheus API and import them into DataDog while handling some common edge and error cases.

Destination metrics could be modified to match a DataDog environment's naming conventions and metrics types by modifying the DataDog API calls as needed.

**These examples are provided as-is, without support. They are intended as reference material only.**

# Running locally

## Prerequisites

* Go 1.20+
* A Datadog API key exported as `DD_API_KEY` in your shell
* You have [configured your Temporal account with CA certificate](https://docs.temporal.io/cloud/how-to-monitor-temporal-cloud-metrics)

## Build the binary

```
make
```

## Running

```
./promqltodd \
  --prom-endpoint https://<temporal-account-id>.tmprl.cloud/prometheus \
  --client-cert <replace with the path to CA cert> \
  --client-key <replace with the path to CA key>
```

# Install promqltodd on a Kubernetes cluster

## Prerequisites

In addition to the Datadog API key and Temporal Cloud CA certs, you have:
* the system configured to access a Kubernetes cluster (e. g. [AWS EKS](https://aws.amazon.com/eks/), [kind](https://kind.sigs.k8s.io/), or [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)), and
* your machine below clis installed:
  - [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/), and
  - [Helm v3](https://helm.sh)

## Install promqltodd with Helm Chart

Run below command from the `helm-charts` folder.

```
helm install promqltodd . \
  --set prom_endpoint=https://<temporal-account-id>.tmprl.cloud/prometheus \
  --set dd_api_key=${DD_API_KEY} \
  --set-file 'ca_cert=<replace with the path to CA cert>' \
  --set-file 'ca_key=<replace with the path to CA key>'
```

## Verify promqltodd is running

```
kubectl logs $(kubectl get pods | grep promqltodd | awk '{print $1; exit}') -f
```

Should output:

```
2023/06/20 11:33:43 Querying Prometheus
2023/06/20 11:33:43 Found 1 histogram metrics: [temporal_cloud_v0_service_latency_bucket]
...
```

## License
MIT License, please see [LICENSE](LICENSE) for details.
