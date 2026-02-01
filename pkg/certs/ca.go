package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"os"

	"github.com/redat00/peekl/pkg/models"
)

func CreateCertificateAuthority(params models.CertificateAuthorityParameters, outCertFilePath string, outKeyFilePath string) error {
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

	// Create private key
	curve := elliptic.P384()
	caKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return err
	}

	// Create certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, &ca, &ca, &caKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	// Save private key
	marshalledPrivKey, err := x509.MarshalPKCS8PrivateKey(caKey)
	if err != nil {
		return err
	}
	caPrivKeyOut, err := os.Create(outKeyFilePath)
	if err != nil {
		return err
	}
	if err := pem.Encode(caPrivKeyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: marshalledPrivKey}); err != nil {
		return err
	}

	// Save certificate
	caCertFileOut, err := os.Create(outCertFilePath)
	if err != nil {
		return err
	}
	if err := pem.Encode(caCertFileOut, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes}); err != nil {
		return err
	}

	return nil
}
