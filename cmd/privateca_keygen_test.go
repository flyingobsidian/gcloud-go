package cmd

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	privateca "google.golang.org/api/privateca/v1"
)

func TestPCAGenerateKeyPair(t *testing.T) {
	cases := []struct {
		alg      string
		privType string
		check    func(t *testing.T, pub any)
	}{
		{"rsa-2048", "RSA PRIVATE KEY", rsaBits(2048)},
		{"rsa-3072", "RSA PRIVATE KEY", rsaBits(3072)},
		{"ec-p256", "EC PRIVATE KEY", ecCurve("P-256")},
		{"ec-p384", "EC PRIVATE KEY", ecCurve("P-384")},
	}
	for _, c := range cases {
		priv, pub, err := pcaGenerateKeyPair(c.alg)
		if err != nil {
			t.Fatalf("%s: %v", c.alg, err)
		}
		if b, _ := pem.Decode(priv); b == nil || b.Type != c.privType {
			t.Errorf("%s: private PEM type = %v, want %s", c.alg, b, c.privType)
		}
		b, _ := pem.Decode(pub)
		if b == nil || b.Type != "PUBLIC KEY" {
			t.Fatalf("%s: public PEM = %v", c.alg, b)
		}
		k, err := x509.ParsePKIXPublicKey(b.Bytes)
		if err != nil {
			t.Fatalf("%s: %v", c.alg, err)
		}
		c.check(t, k)
	}
	if _, _, err := pcaGenerateKeyPair("dsa-1024"); err == nil {
		t.Error("expected error for unsupported algorithm")
	}
}

func rsaBits(n int) func(*testing.T, any) {
	return func(t *testing.T, k any) {
		r, ok := k.(*rsa.PublicKey)
		if !ok || r.N.BitLen() != n {
			t.Errorf("want RSA-%d public key, got %T", n, k)
		}
	}
}

func ecCurve(name string) func(*testing.T, any) {
	return func(t *testing.T, k any) {
		e, ok := k.(*ecdsa.PublicKey)
		if !ok || e.Curve.Params().Name != name {
			t.Errorf("want EC %s public key, got %T", name, k)
		}
	}
}

func TestValidatePCAKeyAlgorithm(t *testing.T) {
	for _, a := range pcaKeyAlgorithms {
		if err := validatePCAKeyAlgorithm(a); err != nil {
			t.Errorf("%s: %v", a, err)
		}
	}
	want := "argument --key-algorithm: Invalid choice: 'rsa-1024'. Valid choices are [rsa-2048, rsa-3072, rsa-4096, ec-p256, ec-p384]."
	if err := validatePCAKeyAlgorithm("rsa-1024"); err == nil || err.Error() != want {
		t.Errorf("got %v, want %q", err, want)
	}
}

func TestApplyPCAGeneratedKey(t *testing.T) {
	out := filepath.Join(t.TempDir(), "key.pem")
	cert := &privateca.Certificate{}
	if err := applyPCAGeneratedKey(cert, "ec-p256", out); err != nil {
		t.Fatal(err)
	}
	if cert.Config == nil || cert.Config.PublicKey == nil || cert.Config.PublicKey.Format != "PEM" {
		t.Fatalf("publicKey not set: %+v", cert.Config)
	}
	raw, err := base64.StdEncoding.DecodeString(cert.Config.PublicKey.Key)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := pem.Decode(raw); b == nil || b.Type != "PUBLIC KEY" {
		t.Errorf("publicKey.key is not a base64 PEM public key")
	}
	st, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o400 {
		t.Errorf("key file mode = %v, want 0400", st.Mode().Perm())
	}
	// A second export overwrites the read-only file.
	if err := applyPCAGeneratedKey(&privateca.Certificate{}, "rsa-2048", out); err != nil {
		t.Errorf("re-export: %v", err)
	}
	if err := applyPCAGeneratedKey(&privateca.Certificate{PemCsr: "x"}, "rsa-2048", out); err == nil {
		t.Error("expected error when body has pemCsr")
	}
}

func TestValidatePCACertKeyFlags(t *testing.T) {
	defer func() {
		flagPCACertGenerateKey = false
		flagPCACertKeyOutputFile = ""
		flagPCACertKeyAlgorithm = "rsa-2048"
	}()
	flagPCACertKeyAlgorithm = "rsa-2048"
	flagPCACertGenerateKey = true
	flagPCACertKeyOutputFile = ""
	if err := validatePCACertKeyFlags(pcaCertCreateCmd); err == nil {
		t.Error("expected error for --generate-key without --key-output-file")
	}
	flagPCACertKeyOutputFile = "k.pem"
	if err := validatePCACertKeyFlags(pcaCertCreateCmd); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	flagPCACertGenerateKey = false
	flagPCACertKeyOutputFile = ""
	if err := validatePCACertKeyFlags(pcaCertCreateCmd); err != nil {
		t.Errorf("no key flags: %v", err)
	}
}
