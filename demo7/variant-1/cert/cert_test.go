package cert_test

import (
	"testing"
	"time"

	"envoy-sds/cert"
)

func TestCertCA(t *testing.T) {
	t.Parallel()

	_, _, _, _, err := cert.GenCARoot()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCert(t *testing.T) {
	t.Parallel()

	rootCert, _, rootKey, _, err := cert.GenCARoot()
	if err != nil {
		t.Fatal(err)
	}

	serverCert, _, _, _, err := cert.GenServerCert(
		[]string{"test"},
		rootCert,
		rootKey,
		time.Minute,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = cert.Verify(rootCert, serverCert)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadCert(t *testing.T) {
	t.Parallel()

	if err := cert.Init(); err != nil {
		t.Fatal(err)
	}

	serverCert, _, _, _, err := cert.NewCertificate([]string{"test"}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	err = cert.Verify(cert.GetLoadedRootCert(), serverCert)
	if err != nil {
		t.Fatal(err)
	}
}
