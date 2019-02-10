package certs

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/xenolf/lego/platform/config/env"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/EngineerBetter/concourse-up/iaas"

	"github.com/square/certstrap/pkix"
	"github.com/xenolf/lego/certificate"
	"github.com/xenolf/lego/challenge"
	"github.com/xenolf/lego/lego"
	"github.com/xenolf/lego/providers/dns/gcloud"
	"github.com/xenolf/lego/providers/dns/route53"
	"github.com/xenolf/lego/registration"
)

// Certs contains certificates and keys
type Certs struct {
	CACert []byte
	Key    []byte
	Cert   []byte
}

// User contains a key, a registration resource, and a sync parameter
type User struct {
	k crypto.PrivateKey
	r *registration.Resource
	sync.Once
}

// GetEmail returns the email for a user
func (u *User) GetEmail() string {
	return "nobody@madeupemailaddress.com"
}

// GetRegistration returns the registration for a user
func (u *User) GetRegistration() *registration.Resource {
	return u.r
}

// GetPrivateKey returns the private key for a user
func (u *User) GetPrivateKey() crypto.PrivateKey {
	u.Do(func() {
		var err error
		u.k, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}
	})
	return u.k
}

func hasIP(x []string) bool {
	for _, v := range x {
		if net.ParseIP(v) != nil {
			return true
		}
	}
	return false
}

func customNewDNSProviderServiceAccount(saFile string, config *gcloud.Config) (*gcloud.DNSProvider, error) {
	if saFile == "" {
		return nil, fmt.Errorf("googlecloud: Service Account file missing")
	}

	dat, err := ioutil.ReadFile(saFile)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: unable to read Service Account file: %v", err)
	}

	// If GCE_PROJECT is non-empty it overrides the project in the service
	// account file.
	project := env.GetOrDefaultString("GCE_PROJECT", "")
	if project == "" {
		// read project id from service account file
		var datJSON struct {
			ProjectID string `json:"project_id"`
		}
		err = json.Unmarshal(dat, &datJSON)
		if err != nil || datJSON.ProjectID == "" {
			return nil, fmt.Errorf("googlecloud: project ID not found in Google Cloud Service Account file")
		}
		project = datJSON.ProjectID
	}

	conf, err := google.JWTConfigFromJSON(dat, dns.NdevClouddnsReadwriteScope)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: unable to acquire config: %v", err)
	}
	client := conf.Client(context.Background())

	config.Project = project
	config.HTTPClient = client

	return gcloud.NewDNSProviderConfig(config)
}

// Generate generates certs for use in a bosh director manifest
func Generate(constructor func(u *User) (*lego.Client, error), caName string, provider iaas.Provider, ipOrDomains ...string) (*Certs, error) {

	if hasIP(ipOrDomains) {
		return generateSelfSigned(caName, ipOrDomains...)
	}
	u := &User{}

	c, err := constructor(u)
	if err != nil {
		return nil, err
	}

	c.Challenge.Remove(challenge.HTTP01)
	c.Challenge.Remove(challenge.TLSALPN01)

	switch provider.IAAS() {
	case iaas.AWS:
		dnsConfig := route53.NewDefaultConfig()
		dnsConfig.PropagationTimeout = 10 * time.Minute
		dnsConfig.PollingInterval = 30 * time.Second
		dnsProvider, err1 := route53.NewDNSProviderConfig(dnsConfig)
		if err1 != nil {
			return nil, err1
		}

		err1 = c.Challenge.SetDNS01Provider(dnsProvider)
		if err1 != nil {
			return nil, err1
		}
	case iaas.GCP:
		dnsConfig := gcloud.NewDefaultConfig()
		dnsConfig.PropagationTimeout = 10 * time.Minute
		dnsConfig.PollingInterval = 30 * time.Second
		dnsProvider, err1 := customNewDNSProviderServiceAccount(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), dnsConfig)
		if err1 != nil {
			return nil, err1
		}
		err1 = c.Challenge.SetDNS01Provider(dnsProvider)
		if err1 != nil {
			return nil, err1
		}
	}
	u.r, err = c.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, err
	}
	request := certificate.ObtainRequest{
		Domains:    ipOrDomains,
		Bundle:     true,
		PrivateKey: nil,
		MustStaple: false,
	}
	certificates, err := c.Certificate.Obtain(request)
	if err != nil {
		return nil, err
	}
	return &Certs{
		CACert: certificates.IssuerCertificate,
		Key:    certificates.PrivateKey,
		Cert:   certificates.Certificate,
	}, nil
}

// Generate generates certs for use in a bosh director manifest
func generateSelfSigned(caName string, ipOrDomains ...string) (*Certs, error) {
	caCert, caKey, err := generateCACert(caName)
	if err != nil {
		return nil, err
	}

	csr, key, err := generateCertificateSigningRequest(ipOrDomains)
	if err != nil {
		return nil, err
	}

	cert, err := signCSR(csr, caCert, caKey)
	if err != nil {
		return nil, err
	}

	caCertByte, err := caCert.Export()
	if err != nil {
		return nil, err
	}

	keyBytes, err := key.ExportPrivate()
	if err != nil {
		return nil, err
	}

	certBytes, err := cert.Export()
	if err != nil {
		return nil, err
	}

	return &Certs{
		CACert: caCertByte,
		Key:    keyBytes,
		Cert:   certBytes,
	}, nil
}

func signCSR(csr *pkix.CertificateSigningRequest, caCert *pkix.Certificate, caCertKey *pkix.Key) (*pkix.Certificate, error) {
	crtOut, err := pkix.CreateCertificateHost(caCert, caCertKey, csr, time.Now().Add(2*365*24*time.Hour))
	if err != nil {
		return nil, err
	}

	return crtOut, nil
}

func generateCACert(caName string) (*pkix.Certificate, *pkix.Key, error) {
	key, err := pkix.CreateRSAKey(4096)
	if err != nil {
		return nil, nil, err
	}

	caCert, err := pkix.CreateCertificateAuthority(
		key,
		"",
		time.Now().Add(10*365*24*time.Hour),
		"",
		"",
		"",
		"",
		caName,
	)
	if err != nil {
		return nil, nil, err
	}

	return caCert, key, nil
}

func generateCertificateSigningRequest(ipOrDomains []string) (*pkix.CertificateSigningRequest, *pkix.Key, error) {
	var name string
	domains := []string{}
	ips := []net.IP{}

	for _, ipOrDomain := range ipOrDomains {
		if name == "" {
			name = ipOrDomain
		}

		if parsedIPs, err := pkix.ParseAndValidateIPs(ipOrDomain); err != nil {
			// if invalid IP address, assume domain instead
			if strings.Contains(err.Error(), "Invalid IP address") {
				domains = append(domains, ipOrDomain)
			} else {
				return nil, nil, err
			}
		} else {
			ips = append(ips, parsedIPs[0])
		}
	}

	key, err := pkix.CreateRSAKey(2048)
	if err != nil {
		return nil, nil, err
	}

	csr, err := pkix.CreateCertificateSigningRequest(
		key,
		"",
		ips,
		domains,
		"",
		"",
		"",
		"",
		name,
	)
	if err != nil {
		return nil, nil, err
	}

	return csr, key, nil
}
