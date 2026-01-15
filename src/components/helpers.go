package analytics_worker

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type CertConfig struct {
	Organization  string
	OrganizationalUnit string
	CommonName    string
	Email         string
	ValidityHours int
}

type CertInfo struct {
	CertPEM   string
	KeyPEM    string
	CertPath  string
	KeyPath   string
}

func generateCert(config CertConfig) (*CertInfo, error) {
	dnsNames := []string{config.CommonName}
	if config.Email!= "" {
		dnsNames = append(dnsNames, config.Email)
	}
	if config.OrganizationalUnit!= "" {
		dnsNames = append(dnsNames, config.OrganizationalUnit)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err!= nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{config.Organization},
			CommonName:   config.CommonName,
		},
		DNSNames:         dnsNames,
		NotBefore:        time.Now(),
		NotAfter:         time.Now().Add(time.Duration(config.ValidityHours) * time.Hour),
		BasicConstraintsValid: true,
		IsCA: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err!= nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	certPath := filepath.Join(os.TempDir(), "cert.pem")
	keyPath := filepath.Join(os.TempDir(), "key.pem")

	if err := os.WriteFile(certPath, certPEM, 0600); err!= nil {
		return nil, err
	}
	if err := os.WriteFile(keyPath, keyPEM, 0600); err!= nil {
		return nil, err
	}

	return &CertInfo{
		CertPEM:   string(certPEM),
		KeyPEM:    string(keyPEM),
		CertPath:  certPath,
		KeyPath:   keyPath,
	}, nil
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err!= nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok &&!ipnet.IP.IsLoopback() {
			if ipnet.IP.To4()!= nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no local IP address found")
}