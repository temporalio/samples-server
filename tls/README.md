These are three separate examples for Transport Layer Security (TLS) setup of Temporal. See their individual READMEs for more detail.

- `/client-only`: These scripts generate only the client-side certificates, along with their keys and configuration files. For Alpine Linux and Mac.
- `/tls-simple`: This samples demonstrates how to configure TLS to secure network communication with and within Temporal cluster.
- `/tls-full`: This sample demonstrates how to configure TLS to secure network communication with and within a Temporal cluster when using intermediate CAs and different certificate chains for cluster and clients. It also shows how different clients can be given different server certificates when connecting to the same cluster using different server names.
