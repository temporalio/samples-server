package prometheus

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
)

func BuildTLSConfig(clientCert, clientKey, serverRootCACert, serverName string, insecureSkipVerify bool) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		log.Fatalf("failed load key pairs: %s", err)
	}

	// Load server CA if given
	var serverCAPool *x509.CertPool
	if serverRootCACert != "" {
		serverCAPool = x509.NewCertPool()
		b, err := os.ReadFile(serverRootCACert)
		if err != nil {
			return nil, fmt.Errorf("failed reading server CA: %w", err)
		} else if !serverCAPool.AppendCertsFromPEM(b) {
			return nil, fmt.Errorf("server CA PEM file invalid")
		}
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            serverCAPool,
		ServerName:         serverName,
		InsecureSkipVerify: insecureSkipVerify,
	}, nil
}
