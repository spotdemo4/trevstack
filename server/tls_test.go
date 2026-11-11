package main

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"runtime"
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
	// Listen presents the development certificate itself, and refuses a TLS
	// configuration alongside it.
	if config != nil {
		t.Error("loadTLSConfig() returned a TLS configuration with a development certificate")
	}
	cert, err := development.GetCertificate(nil)
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

func TestDevelopmentCertificateWriterWritesPEM(t *testing.T) {
	development, err := trevrpc.NewDevelopmentCertificate()
	if err != nil {
		t.Fatalf("NewDevelopmentCertificate() error = %v", err)
	}
	path := filepath.Join(t.TempDir(), developmentCertificateFile)
	writer := &developmentCertificateWriter{path: path, development: development}
	if err := writer.write(); err != nil {
		t.Fatalf("write() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	want, err := development.PEM()
	if err != nil {
		t.Fatalf("PEM() error = %v", err)
	}
	if !bytes.Equal(data, want) {
		t.Error("file does not hold the development certificates")
	}

	// The file must hold the certificate the server presents, which clients
	// verify against it.
	cert, err := development.GetCertificate(nil)
	if err != nil {
		t.Fatalf("GetCertificate() error = %v", err)
	}
	block, _ := pem.Decode(data)
	if block == nil || !bytes.Equal(block.Bytes, cert.Certificate[0]) {
		t.Error("file does not start with the presented certificate")
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Errorf("mode = %04o, want 0600", got)
		}
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("directory holds %d entries, want only the certificate file", len(entries))
	}
}

func TestDevelopmentCertificateWriterSkipsUnchangedCertificates(t *testing.T) {
	development, err := trevrpc.NewDevelopmentCertificate()
	if err != nil {
		t.Fatalf("NewDevelopmentCertificate() error = %v", err)
	}
	path := filepath.Join(t.TempDir(), developmentCertificateFile)
	writer := &developmentCertificateWriter{path: path, development: development}
	if err := writer.write(); err != nil {
		t.Fatalf("write() error = %v", err)
	}

	// Another server sharing the file replaced it; this one leaves it until
	// its own certificates rotate.
	other := []byte("another server's certificates")
	if err := os.WriteFile(path, other, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	if err := writer.write(); err != nil {
		t.Fatalf("write() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !bytes.Equal(data, other) {
		t.Error("write() replaced the file although the certificates had not rotated")
	}
}

func writePEM(t *testing.T, path, blockType string, der []byte) {
	t.Helper()
	data := pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
