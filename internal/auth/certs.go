package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/dedene/frontapp-cli/internal/config"
)

// EnsureCertificate returns paths to cert and key files, generating them if needed.
func EnsureCertificate() (certPath, keyPath string, err error) {
	dir, err := config.EnsureDir()
	if err != nil {
		return "", "", fmt.Errorf("ensure config dir: %w", err)
	}

	certPath = filepath.Join(dir, "localhost.crt")
	keyPath = filepath.Join(dir, "localhost.key")

	// Check if both exist
	if fileExists(certPath) && fileExists(keyPath) {
		return certPath, keyPath, nil
	}

	// Generate new certificate
	if genErr := generateCertificate(certPath, keyPath); genErr != nil {
		return "", "", fmt.Errorf("generate certificate: %w", genErr)
	}

	return certPath, keyPath, nil
}

func generateCertificate(certPath, keyPath string) error {
	// Generate RSA key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	// Serial number
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("generate serial: %w", err)
	}

	// Certificate template (valid 10 years for localhost)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"frontcli"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	// Create certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("create certificate: %w", err)
	}

	// Write certificate
	certOut, err := os.Create(certPath) //nolint:gosec // localhost cert path
	if err != nil {
		return fmt.Errorf("create cert file: %w", err)
	}
	defer certOut.Close()

	if encErr := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); encErr != nil {
		return fmt.Errorf("encode certificate: %w", encErr)
	}

	// Write private key (0600 permissions)
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) //nolint:gosec // localhost key path
	if err != nil {
		return fmt.Errorf("create key file: %w", err)
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}

	if encErr := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); encErr != nil {
		return fmt.Errorf("encode key: %w", encErr)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
