### TLS
This sample demonstrates how to configure Transport Layer Security (TLS) to secure network communication with and within a Temporal cluster when using intermediate CAs and different certificate chains for cluster and clients.
It also shows how different clients can be given different server certificates when connecting to the same cluster using different server names.
The generated certificates are in PKCS1 and PKCS12 (client-only) format.

### Steps to run this sample
1. Generate test certificates with `generate-certs.sh`. This will create server and client certificates in a `certs` subdirectory.

```bash
bash generate-certs.sh
```

2. Start Temporal with `start-temporal.sh`. This will bring up a Temporal cluster (via `docker-compose`) with the `certs` subdirectory mounted as a volume and Temporal configured to use the test certificates in it to secure network communications.

```bash
bash start-temporal.sh
```

