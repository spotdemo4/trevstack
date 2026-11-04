package main

import (
	"crypto/tls"
	"fmt"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
)

// loadTLSConfig returns the listener's TLS configuration with the certificate
// in certFile and keyFile, or, when both are empty, with a rotating
// development certificate for localhost that browsers trust by its hashes,
// which is returned too.
func loadTLSConfig(certFile, keyFile string) (*tls.Config, *trevrpc.DevelopmentCertificate, error) {
	if certFile != "" || keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, nil, fmt.Errorf("could not load certificate: %w", err)
		}
		return &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS13}, nil, nil
	}

	development, err := trevrpc.NewDevelopmentCertificate()
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate certificate: %w", err)
	}
	return development.TLSConfig(), development, nil
}
