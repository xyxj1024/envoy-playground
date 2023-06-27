package tls

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"envoy-swarm-control/pkg/storage"
)

type Certificate struct {
	storage.Storage
}

const (
	CertificateExtension = "pem"
	PrivateKeyExtension  = "key"
)

func (c *Certificate) PutCertificate(domain string, sans []string, publicChain, privateKey []byte) (err error) {
	fileName := getCertificateFilename(domain, sans)
	err = c.PutFile(fmt.Sprintf("%s.%s", fileName, CertificateExtension), publicChain)
	if err != nil {
		return err
	}

	return c.PutFile(fmt.Sprintf("%s.%s", fileName, PrivateKeyExtension), privateKey)
}

/* Function GetCertificate:
 * returns the .pem and .key files for SDS provider.
 */
func (c *Certificate) GetCertificate(domain string, sans []string) (publicChain, privateKey []byte, err error) {
	fileName := getCertificateFilename(domain, sans)
	publicChain, err = c.GetFile(fmt.Sprintf("%s.%s", fileName, CertificateExtension))
	if err != nil {
		return nil, nil, err
	}

	privateKey, err = c.GetFile(fmt.Sprintf("%s.%s", fileName, PrivateKeyExtension))
	if err != nil {
		return nil, nil, err
	}

	return publicChain, privateKey, err
}

func getCertificateFilename(primaryDomain string, sans []string) string {
	filename := strings.ToLower(primaryDomain)

	sortedDomains := make([]string, len(sans))
	_ = copy(sortedDomains, sans)
	sort.Strings(sortedDomains)

	sum := sha256.Sum256([]byte(strings.Join(sortedDomains, "")))
	hash := base64.StdEncoding.EncodeToString(sum[:])

	return strings.NewReplacer("/", "", "\\", "").Replace(fmt.Sprintf("%s-%s", filename, hash[:16]))
}
