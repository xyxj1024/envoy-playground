package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"

	"envoy-sds/cert"
)

const basedir = "/Users/xuanyuanxingjian/Documents/projects/Repos/github-envoy-playground/demo7/variant-1"

func main() {
	certPath := flag.String("cert-path", basedir+"/cert", "path to generate certificates")
	dnsNames := flag.String("dns-names", "localhost", "DNS names for server certificate")
	flag.Parse()

	files := make(map[string][]byte)

	if err := cert.Init(); err != nil {
		log.Fatalf("failed to generate certificates: %v", err)
	}

	rootCrt := cert.GetLoadedRootCert()
	rootCrtBytes := cert.GetLoadedRootCertBytes()
	rootKey := cert.GetLoadedRootKey()
	rootKeyBytes, err := cert.GetLoadedRootKeyBytes()
	if err != nil {
		log.Fatalf("failed to load root key bytes: %v", err)
	}
	_, serverCrtBytes, _, serverKeyBytes, err := cert.GenServerCert(
		strings.Split(*dnsNames, ","),
		rootCrt,
		rootKey,
		cert.CertValidityMax,
	)
	if err != nil {
		log.Fatalf("failed to generate certificates for host: %v", err)
	}
	_, clientCrtBytes, _, clientKeyBytes, err := cert.GenServerCert(
		[]string{"envoy"},
		rootCrt,
		rootKey,
		cert.CertValidityMax,
	)
	if err != nil {
		log.Fatalf("failed to generate certificates for Envoy: %v", err)
	}

	files["ca.crt"] = rootCrtBytes
	files["ca.key"] = rootKeyBytes
	files["server.crt"] = serverCrtBytes
	files["server.key"] = serverKeyBytes
	files["client.crt"] = clientCrtBytes
	files["client.key"] = clientKeyBytes

	const fileMode = fs.FileMode(0o644)

	for fileName, fileContent := range files {
		filePath := path.Join(*certPath, fileName)

		log.Printf("saving file %s\n", filePath)

		if err = os.WriteFile(filePath, fileContent, fileMode); err != nil {
			log.Fatal(err)
		}
	}

	log.Println("certificates generated")
}
