package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	//"math/big"
	"os"
	"path/filepath"

	"github.com/redat00/peekl/pkg/models"
)

func CreateCertificateAuthority(caFolder string, params models.CertificateAuthorityParameters) error {
	ca := x509.Certificate{
		Subject: pkix.Name{
			Organization:  []string{"Peekl"},
			Country:       []string{"FR"},
			Province:      []string{""},
			Locality:      []string{"Paris"},
			StreetAddress: []string{"Le Marais"},
			PostalCode:    []string{"75004"},
		},
		NotBefore:             params.NotBefore,
		NotAfter:              params.NotAfter,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	curve := elliptic.P384()
	caKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return err
	}

	marshalledPrivKey, err := x509.MarshalECPrivateKey(caKey)
	if err != nil {
		return err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, &ca, &ca, &caKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	caPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: marshalledPrivKey,
	})

	caPath := filepath.Join(caFolder, "ca.pem")
	keyPath := filepath.Join(caFolder, "ca.key")

	err = os.WriteFile(caPath, caPem, 0600)
	err = os.WriteFile(keyPath, caPrivKeyPem, 0600)

	return nil
}
