# TLS

This samples demonstrates how to configure Transport Layer Security (TLS) to secure network communication with and within Temporal cluster.
The generated certificates are in 
  - PKCS1 format for server
  - PKCS8 and PKCS12 format for client

## Steps to run this sample

1. Generate test certificates with `generate-test-certs.sh`. This will create server and client certificates in the `certs` subdirectory.

```bash
bash generate-test-certs.sh
```

2. Start Temporal with `start-temporal.sh`. This will bring up a Temporal cluster (via `docker-compose`) with the `certs` subdirectory mounted as a volume and Temporal configured to use the test certificates in it to secure network communications.

```bash
bash start-temporal.sh
```

### Disabling Client Authentication
The Temporal Cluster you launched by running the commands above uses mutual TLS (mTLS), meaning that it requires the client and server to authenticate one another by verifying each other's certificates when making a connection. If you would prefer to use TLS (that is, disable the server's veriification of the client's certificate), edit the `docker-compose.yml` file, change the value of `TEMPORAL_TLS_REQUIRE_CLIENT_AUTH` variable from `true` to `false`, and then restart the `start-temporal.sh` script.

#### Connecting to the Cluster via TLS (Command Line)
After disabling client authentication as per the above directions, you could use the `temporal` command to connect to the cluster by specifying options for the path to the CA certificate and the TLS Server Name. The following example shows how to use these options to register a new namespace (`testing`, in this example):

```bash
temporal operator namespace create \
    --tls-ca-path certs/ca.cert \
    --tls-server-name tls-sample \
    testing
```

Here is the corresponding `tctl` command:
```bash
tctl \
    --tls_ca_path certs/ca.cert \
    --tls_server_name tls-sample \
    --namespace testing \
    namespace register
```

#### Connecting to the Cluster via TLS (Go SDK)

The following example shows how to use the Go SDK to create a 
Temporal that can connect to this Temporal Cluster using TLS:

```go
import (
	"go.temporal.io/sdk/client"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
)


func createClient() {
	// load the server's certificate 
	serverPEM, err := os.ReadFile("certs/cluster.pem")
	if err != nil {
		log.Fatalln("failed to load server certificate")
	}

	// add it to a set of certificate authorities
	serverCAPool := x509.NewCertPool()
	if !serverCAPool.AppendCertsFromPEM(serverPEM) {
		log.Fatalln("invalid server cert PEM")
	}

	// configure the TLS connection
	c, err := client.Dial(client.Options{
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{
				RootCAs:      serverCAPool,
				ServerName:   "tls-sample",
			},
		},
	})

	if err != nil {
		log.Fatalln("unable to create client", err)
	}
	defer c.Close()

	// Code that uses the Client would follow
```
