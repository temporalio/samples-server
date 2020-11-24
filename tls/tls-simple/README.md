### TLS
This samples demonstrates how to configure Transport Layer Security (TLS) to secure network communication with and within Temporal cluster.

### Steps to run this sample
1. Generate test certificates with `generate-test-certs.sh`. This will create server and client certificates in the `certs` subdirectory.

```bash
bash generate-test-certs.sh
```

2. Start Temporal with `start-temporal.sh`. This will bring up a Temporal cluster (via `docker-compose`) with the `certs` subdirectory mounted as a volume and Temporal configured to use the test certificates in it to secure network communications.

```bash
bash start-temporal.sh
```

