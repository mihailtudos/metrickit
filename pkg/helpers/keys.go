package helpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"
)

const (
	rsaKeySize = 2048
)

// GenerateKeyPair generates a new RSA key pair and saves it to the specified path.
func GenerateKeyPair(p string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	pDir := path.Dir(p)

	// Save private key
	privateFile, err := os.Create(path.Join(pDir, "private.pem"))
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}

	if err = pem.Encode(privateFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	// Save public key
	publicFile, err := os.Create(path.Join(pDir, "public.pem"))
	if err != nil {
		return fmt.Errorf("failed to create public key file: %w", err)
	}

	if err := pem.Encode(publicFile, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	}); err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	return nil
}
