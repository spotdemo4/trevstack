package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
)

const (
	// developmentCertificateFile is the file in the configuration directory
	// that the development certificates are written to, for native clients
	// on the same machine to trust.
	developmentCertificateFile = "development.pem"

	// developmentCertificateRefresh is how often the development certificate
	// file is brought up to date, which keeps it current within this long of
	// a rotation.
	developmentCertificateRefresh = time.Hour
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

// developmentCertificateWriter keeps a file holding the current and next
// development certificates, which native clients trust as roots.
type developmentCertificateWriter struct {
	path        string
	development *trevrpc.DevelopmentCertificate
	// written is what was last written, so the file is rewritten only after a
	// rotation, and servers sharing it do not take turns replacing it.
	written []byte
}

// write brings the file up to date, replacing it in one rename so that a
// client never reads part of it.
func (w *developmentCertificateWriter) write() error {
	data, err := w.development.PEM()
	if err != nil {
		return err
	}
	if bytes.Equal(data, w.written) {
		return nil
	}

	// CreateTemp creates the file with 0600 permissions.
	tmp, err := os.CreateTemp(filepath.Dir(w.path), ".development-*.pem")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), w.path); err != nil {
		return err
	}
	w.written = data
	return nil
}

// run writes the file every interval until ctx is done. Failures are logged
// rather than returned, since the server works without the file.
func (w *developmentCertificateWriter) run(ctx context.Context, log *slog.Logger, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.write(); err != nil {
				log.WarnContext(ctx, "could not write the development certificate", "path", w.path, "error", err)
			}
		}
	}
}
