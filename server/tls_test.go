package main

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
)

func TestLoadTLSConfigUsesDevelopmentCertificate(t *testing.T) {
	config, development, err := loadTLSConfig("", "")
	if err != nil {
		t.Fatalf("loadTLSConfig() error = %v", err)
	}
	if development == nil {
		t.Fatal("loadTLSConfig() returned no development certificate")
	}
	cert, err := config.GetCertificate(nil)
	if err != nil {
		t.Fatalf("GetCertificate() error = %v", err)
	}
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	if err := leaf.VerifyHostname("localhost"); err != nil {
		t.Errorf("VerifyHostname(localhost) error = %v", err)
	}
}

func TestLoadTLSConfigLoadsCertificate(t *testing.T) {
	development, err := trevrpc.NewDevelopmentCertificate()
	if err != nil {
		t.Fatalf("NewDevelopmentCertificate() error = %v", err)
	}
	cert, err := development.TLSConfig().GetCertificate(nil)
	if err != nil {
		t.Fatalf("GetCertificate() error = %v", err)
	}
	key, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")
	writePEM(t, certFile, "CERTIFICATE", cert.Certificate[0])
	writePEM(t, keyFile, "PRIVATE KEY", key)

	config, loadedDevelopment, err := loadTLSConfig(certFile, keyFile)
	if err != nil {
		t.Fatalf("loadTLSConfig() error = %v", err)
	}
	if loadedDevelopment != nil {
		t.Error("loadTLSConfig() returned a development certificate for files")
	}
	if len(config.Certificates) != 1 || string(config.Certificates[0].Certificate[0]) != string(cert.Certificate[0]) {
		t.Error("TLS configuration does not serve the loaded certificate")
	}

	if _, _, err := loadTLSConfig(certFile, filepath.Join(dir, "missing.pem")); err == nil {
		t.Error("loadTLSConfig() with a missing key error = nil")
	}
}

func writePEM(t *testing.T, path, blockType string, der []byte) {
	t.Helper()
	data := pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
