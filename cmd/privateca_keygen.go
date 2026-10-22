package cmd

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	privateca "google.golang.org/api/privateca/v1"
)

// pcaKeyAlgorithms lists the --key-algorithm choices of
// `gcloud privateca certificates create --generate-key`, in the order of
// gcloud-python's key_generation.KeyAlgorithm enum (588.0.0).
var pcaKeyAlgorithms = []string{"rsa-2048", "rsa-3072", "rsa-4096", "ec-p256", "ec-p384"}

// pcaKeyOutputWarning mirrors key_generation.KEY_OUTPUT_WARNING.
const pcaKeyOutputWarning = `A private key was exported to %s.

Possession of this key file could allow anybody to act as this certificate's
subject. Please make sure that you store this key file in a secure location at
all times, and ensure that only authorized users have access to it.`

// pcaGenerateKeyPair generates a key pair for a --key-algorithm choice and
// returns the private key as a "traditional OpenSSL" PEM (PKCS#1 for RSA,
// SEC 1 for EC) and the public key as a SubjectPublicKeyInfo PEM, matching
// gcloud-python's KeyGenForAlgorithm.
func pcaGenerateKeyPair(algorithm string) (privPEM, pubPEM []byte, err error) {
	var (
		signer   crypto.Signer
		privDER  []byte
		privType string
	)
	switch algorithm {
	case "rsa-2048", "rsa-3072", "rsa-4096":
		bits := map[string]int{"rsa-2048": 2048, "rsa-3072": 3072, "rsa-4096": 4096}[algorithm]
		k, gerr := rsa.GenerateKey(rand.Reader, bits)
		if gerr != nil {
			return nil, nil, gerr
		}
		signer, privDER, privType = k, x509.MarshalPKCS1PrivateKey(k), "RSA PRIVATE KEY"
	case "ec-p256", "ec-p384":
		curve := elliptic.P256()
		if algorithm == "ec-p384" {
			curve = elliptic.P384()
		}
		k, gerr := ecdsa.GenerateKey(curve, rand.Reader)
		if gerr != nil {
			return nil, nil, gerr
		}
		der, merr := x509.MarshalECPrivateKey(k)
		if merr != nil {
			return nil, nil, merr
		}
		signer, privDER, privType = k, der, "EC PRIVATE KEY"
	default:
		return nil, nil, fmt.Errorf("Invalid value for [--key-algorithm]: Unsupported key algorithm '%s'.", algorithm)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(signer.Public())
	if err != nil {
		return nil, nil, err
	}
	privPEM = pem.EncodeToMemory(&pem.Block{Type: privType, Bytes: privDER})
	pubPEM = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return privPEM, pubPEM, nil
}

// validatePCAKeyAlgorithm rejects --key-algorithm values outside the choice
// list with gcloud's argparse wording.
func validatePCAKeyAlgorithm(algorithm string) error {
	for _, a := range pcaKeyAlgorithms {
		if a == algorithm {
			return nil
		}
	}
	return fmt.Errorf("argument --key-algorithm: Invalid choice: '%s'. Valid choices are [%s].",
		algorithm, strings.Join(pcaKeyAlgorithms, ", "))
}

// exportPCAPrivateKey writes the private key owner-read-only and prints
// gcloud-python's KEY_OUTPUT_WARNING to stderr.
func exportPCAPrivateKey(path string, privPEM []byte) error {
	_ = os.Remove(path) // a previous 0400 export would block O_TRUNC
	if err := os.WriteFile(path, privPEM, 0o600); err != nil {
		return fmt.Errorf("Error writing to private key output file named '%s'", path)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		return fmt.Errorf("Error writing to private key output file named '%s'", path)
	}
	fmt.Fprintf(os.Stderr, "WARNING: "+pcaKeyOutputWarning+"\n", path)
	return nil
}

// applyPCAGeneratedKey generates a key pair, exports the private key to
// keyOutputFile, and sets cert.config.publicKey to the PEM public key, as
// gcloud-python does for `certificates create --generate-key`.
func applyPCAGeneratedKey(cert *privateca.Certificate, algorithm, keyOutputFile string) error {
	if cert.PemCsr != "" {
		return fmt.Errorf("argument --generate-key: --generate-key cannot be used with a pemCsr in --config-file")
	}
	privPEM, pubPEM, err := pcaGenerateKeyPair(algorithm)
	if err != nil {
		return err
	}
	if err := exportPCAPrivateKey(keyOutputFile, privPEM); err != nil {
		return err
	}
	if cert.Config == nil {
		cert.Config = &privateca.CertificateConfig{}
	}
	cert.Config.PublicKey = &privateca.PublicKey{
		Format: "PEM",
		Key:    base64.StdEncoding.EncodeToString(pubPEM),
	}
	return nil
}
