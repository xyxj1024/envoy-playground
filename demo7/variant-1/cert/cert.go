package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix" // contains shared, low level structures used for ASN.1 parsing and serialization of X.509 certificates, CRL and OCSP
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

const (
	keyBits          = 2048
	CertValidity     = 7 * 24 * time.Hour
	CertValidityYear = 365 * 24 * time.Hour
	CertValidityMax  = 3000 * 24 * time.Hour
	sslMaxPathLen    = 2
)

var (
	caCert      *x509.Certificate
	caCertBytes []byte
	caKey       *rsa.PrivateKey
)

func Init() error {
	var err error

	caCert, caCertBytes, caKey, err = loadCAFromFiles()
	if err != nil {
		caCert, caCertBytes, caKey, _, err = GenCARoot()
	}

	if err != nil {
		return err
	}

	fmt.Printf("root CA:\n%s", string(GetLoadedRootCertBytes()))

	return nil
}

func loadCAFromFiles() (*x509.Certificate, []byte, *rsa.PrivateKey, error) {
	certBytes, err := os.ReadFile("./")
	if err != nil {
		return nil, nil, nil, err
	}

	certBlock, _ := pem.Decode(certBytes)

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, nil, err
	}

	keyBytes, err := os.ReadFile("./")
	if err != nil {
		return nil, nil, nil, err
	}

	keyBlock, _ := pem.Decode(keyBytes)

	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, nil, err
	}

	privateKey, _ := key.(*rsa.PrivateKey)

	return cert, certBytes, privateKey, nil
}

func genCert(
	template, parent *x509.Certificate,
	publicKey *rsa.PublicKey,
	privateKey *rsa.PrivateKey,
) (*x509.Certificate, []byte, error) {
	// certBytes is a DER encoded ASN.1 structure
	certBytes, err := x509.CreateCertificate(rand.Reader, template, parent, publicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	// cert is an X.509 certificate
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, nil, err
	}

	// b is a PEM encoded structure
	b := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}
	certPem := pem.EncodeToMemory(&b)

	return cert, certPem, nil
}

/* Function exportPrivateKey:
 * returns the PEM encoding of a given private key.
 */
func exportPrivateKey(privkey *rsa.PrivateKey) ([]byte, error) {
	// privkeyBytes is privkey in the PKCS #8, ASN.1 DER form
	privkeyBytes, err := x509.MarshalPKCS8PrivateKey(privkey)
	if err != nil {
		return nil, err
	}

	// Return the PEM encoding of privkeyBytes
	privkeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privkeyBytes,
		},
	)

	return privkeyPem, nil
}

func GetLoadedRootCert() *x509.Certificate {
	return caCert
}

func GetLoadedRootCertBytes() []byte {
	return caCertBytes
}

func GetLoadedRootKey() *rsa.PrivateKey {
	return caKey
}

func GetLoadedRootKeyBytes() ([]byte, error) {
	return exportPrivateKey(caKey)
}

func NewCertificate(dnsNames []string, certDuration time.Duration) (*x509.Certificate, []byte, *rsa.PrivateKey, []byte, error) {
	return GenServerCert(dnsNames, caCert, caKey, certDuration)
}

func GenCARoot() (*x509.Certificate, []byte, *rsa.PrivateKey, []byte, error) {
	rootTmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:            []string{"US"},
			Province:           []string{"Missouri"},
			Locality:           []string{"Saint Louis"},
			Organization:       []string{"Washington University"},
			OrganizationalUnit: []string{"CA"},
			CommonName:         "Washington University",
		},
		NotBefore:             time.Now().Add(-10 * time.Second),
		NotAfter:              time.Now().Add(CertValidityMax),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true, // the template is a CA!
		MaxPathLen:            sslMaxPathLen,
	}

	// Obtain an RSA key pair of the given bit size using the random source
	priv, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Obtain a self-signed certificate and its PEM encoding
	rootCert, rootCertBytes, err := genCert(&rootTmpl, &rootTmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Obtain the PEM encoding of the private key
	privBytes, err := exportPrivateKey(priv)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return rootCert, rootCertBytes, priv, privBytes, nil
}

func GenServerCert(
	dnsNames []string, // the request hostname must match the first server name protected by the SSL certificate
	rootCert *x509.Certificate,
	rootKey *rsa.PrivateKey,
	certDuration time.Duration,
) (*x509.Certificate, []byte, *rsa.PrivateKey, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	serverTmpl := x509.Certificate{
		// Serial number must be unique for each certificate issued by a given CA
		SerialNumber: new(big.Int).SetInt64(time.Now().Unix()),
		Subject: pkix.Name{
			Country:            []string{"US"},
			Province:           []string{"Missouri"},
			Locality:           []string{"Saint Louis"},
			Organization:       []string{"Washington University"},
			OrganizationalUnit: []string{"CLIENT"},
			CommonName:         dnsNames[0],
		},
		DNSNames:       dnsNames,
		NotBefore:      time.Now().Add(-10 * time.Second),
		NotAfter:       time.Now().Add(certDuration),
		KeyUsage:       x509.KeyUsageDigitalSignature,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IsCA:           false,
		MaxPathLenZero: true,
	}

	serverCert, serverCertBytes, err := genCert(&serverTmpl, rootCert, &priv.PublicKey, rootKey)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	priBytes, err := exportPrivateKey(priv)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return serverCert, serverCertBytes, priv, priBytes, nil
}

func Verify(root, child *x509.Certificate) error {
	roots := x509.NewCertPool()
	inter := x509.NewCertPool()

	roots.AddCert(root)

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: inter,
	}

	if _, err := child.Verify(opts); err != nil {
		return err
	}

	return nil
}
