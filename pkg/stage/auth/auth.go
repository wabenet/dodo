package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	Hosts                                []string
	Bits                                 int
	Org                                  string
	CertFile, KeyFile, CAFile, CAKeyFile string
}

func DefaultOptions(baseDir string) *Options {
	return &Options{
		Hosts:     []string{""},
		Org:       "dodo.<bootstrap>",
		Bits:      2048,
		CertFile:  filepath.Join(baseDir, "cert.pem"),
		KeyFile:   filepath.Join(baseDir, "key.pem"),
		CAFile:    filepath.Join(baseDir, "ca.pem"),
		CAKeyFile: filepath.Join(baseDir, "ca-key.pem"),
	}
}

func ValidateCertificate(addr string, certDir string) (bool, error) {
	opts := DefaultOptions(certDir)

	caCert, err := ioutil.ReadFile(opts.CAFile)
	if err != nil {
		return false, err
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return false, errors.New("could not read certificates")
	}

	cert, err := ioutil.ReadFile(opts.CertFile)
	if err != nil {
		return false, err
	}

	key, err := ioutil.ReadFile(opts.KeyFile)
	if err != nil {
		return false, err
	}

	keypair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return false, err
	}

	dialer := &net.Dialer{Timeout: 20 * time.Second}
	tlsConfig := &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: false,
		Certificates:       []tls.Certificate{keypair},
	}

	if _, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig); err != nil {
		return false, err
	}

	return true, nil
}

func BootstrapCertificates(targetDir string) error {
	opts := DefaultOptions(targetDir)
	if err := ensureDirectory(targetDir); err != nil {
		return err
	}

	if _, err := os.Stat(opts.CAFile); os.IsNotExist(err) {
		if err := createCACert(opts); err != nil {
			return err
		}
	} else {
		current, err := checkDate(opts.CAFile)
		if err != nil {
			return err
		}
		if !current {
			log.Info("CA certificate is outdated and needs to be regenerated")
			if err := createCACert(opts); err != nil {
				return err
			}
		}
	}

	if _, err := os.Stat(opts.CertFile); os.IsNotExist(err) {
		if err := createCert(opts); err != nil {
			return err
		}
	} else {
		current, err := checkDate(opts.CertFile)
		if err != nil {
			return err
		}
		if !current {
			log.Info("Client certificate is outdated and needs to be regenerated")
			if err := createCert(opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func createCACert(opts *Options) error {
	log.WithFields(log.Fields{"file": opts.CAFile}).Infof("creating CA certificate")
	os.Remove(opts.CAKeyFile)

	template, err := x509Template(opts.Org)
	if err != nil {
		return errors.Wrap(err, "could not geterate CA certificate")
	}

	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign
	template.KeyUsage |= x509.KeyUsageKeyEncipherment
	template.KeyUsage |= x509.KeyUsageKeyAgreement

	rsaKey, err := rsa.GenerateKey(rand.Reader, opts.Bits)
	if err != nil {
		return errors.Wrap(err, "could not generate RSA key pair")
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &rsaKey.PublicKey, rsaKey)
	if err != nil {
		return errors.Wrap(err, "could not generate CA certificate")
	}

	if err := writeCertificate(opts.CAFile, certBytes); err != nil {
		return errors.Wrap(err, "could not write CA certificate")
	}
	if err := writePrivateKey(opts.CAKeyFile, x509.MarshalPKCS1PrivateKey(rsaKey)); err != nil {
		return errors.Wrap(err, "could not write CA key")
	}

	return nil
}

func createCert(opts *Options) error {
	log.WithFields(log.Fields{"file": opts.CertFile}).Info("creating client certificate")
	os.Remove(opts.KeyFile)

	template, err := x509Template(opts.Org)
	if err != nil {
		return errors.Wrap(err, "generating client certificate failed")
	}

	// client
	if len(opts.Hosts) == 1 && opts.Hosts[0] == "" {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
		template.KeyUsage = x509.KeyUsageDigitalSignature
	} else { // server
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
		for _, h := range opts.Hosts {
			if ip := net.ParseIP(h); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			} else {
				template.DNSNames = append(template.DNSNames, h)
			}
		}
	}

	tlsCert, err := tls.LoadX509KeyPair(opts.CAFile, opts.CAKeyFile)
	if err != nil {
		return err
	}
	x509Cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return err
	}

	rsaKey, err := rsa.GenerateKey(rand.Reader, opts.Bits)
	if err != nil {
		return errors.Wrap(err, "could not generate RSA key pair")
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, x509Cert, &rsaKey.PublicKey, tlsCert.PrivateKey)
	if err != nil {
		return err
	}

	if err := writeCertificate(opts.CertFile, certBytes); err != nil {
		return errors.Wrap(err, "could not write client certificate")
	}
	if err := writePrivateKey(opts.KeyFile, x509.MarshalPKCS1PrivateKey(rsaKey)); err != nil {
		return errors.Wrap(err, "could not write client key")
	}

	return nil
}

func x509Template(org string) (*x509.Certificate, error) {
	now := time.Now()
	// need to set notBefore slightly in the past to account for time
	// skew in the VMs otherwise the certs sometimes are not yet valid
	notBefore := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()-5, 0, 0, time.Local)
	notAfter := notBefore.Add(time.Hour * 24 * 1080)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{org},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement,
		BasicConstraintsValid: true,
	}, nil
}

func writeCertificate(path string, bytes []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "could not write to file")
	}
	defer file.Close()
	pem.Encode(file, &pem.Block{Type: "CERTIFICATE", Bytes: bytes})
	return nil
}

func writePrivateKey(path string, bytes []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "could not write to file")
	}
	defer file.Close()
	pem.Encode(file, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: bytes})
	return nil
}

func checkDate(path string) (bool, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}

	pemData, _ := pem.Decode(bytes)
	if pemData == nil {
		return false, errors.New("failed to decode PEM data")
	}

	cert, err := x509.ParseCertificate(pemData.Bytes)
	if err != nil {
		return false, err
	}

	if time.Now().After(cert.NotAfter) {
		return false, nil
	}

	return true, nil
}

func ensureDirectory(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0700); err != nil {
				return errors.Wrap(err, "could not create directory")
			}
		} else {
			return err
		}
	}
	return nil
}
