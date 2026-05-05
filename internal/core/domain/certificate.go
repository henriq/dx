package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// CertificateType determines the x509 key usage extensions for an issued certificate.
type CertificateType string

const (
	CertificateTypeServer CertificateType = "server"
	CertificateTypeClient CertificateType = "client"
)

// K8sSecretType controls the Kubernetes secret type used to store the certificate.
type K8sSecretType string

const (
	K8sSecretTypeTLS    K8sSecretType = "tls"
	K8sSecretTypeOpaque K8sSecretType = "opaque"
)

// CertificateRequest defines a certificate to be issued and stored as a Kubernetes secret.
type CertificateRequest struct {
	Type      CertificateType `yaml:"type"`
	DNSNames  []string        `yaml:"dnsNames"`
	K8sSecret K8sSecretConfig `yaml:"k8sSecret"`
}

// K8sSecretConfig defines how the issued certificate is stored in Kubernetes.
type K8sSecretConfig struct {
	Name string            `yaml:"name"`
	Type K8sSecretType     `yaml:"type"`
	Keys *OpaqueSecretKeys `yaml:"keys,omitempty"`
}

// OpaqueSecretKeys defines custom key names for opaque Kubernetes secrets.
type OpaqueSecretKeys struct {
	PrivateKey string `yaml:"privateKey"`
	Cert       string `yaml:"cert"`
	CA         string `yaml:"ca"`
}

// IssuedCertificate holds a newly issued certificate and its private key in PEM format.
type IssuedCertificate struct {
	CertPEM []byte
	KeyPEM  []byte
	CAPEM   []byte
}

// ServiceCertificates pairs a service name with its certificate requests.
// Used by CertificateProvisioner to decouple certificate provisioning from
// the full Service type.
type ServiceCertificates struct {
	Name         string
	Certificates []CertificateRequest
}

// ProvisionedCertificate pairs a certificate request with its secret data,
// ready to be rendered as a K8s Secret manifest in a Helm wrapper chart.
type ProvisionedCertificate struct {
	Request CertificateRequest
	Data    map[string][]byte
}

var dnsNameRegex = regexp.MustCompile(`^(\*\.)?[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?)*$`)
var k8sNameRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9\-\.]*[a-z0-9])?$`)
var k8sSecretKeyRegex = regexp.MustCompile(`^[-._a-zA-Z0-9]+$`)

// allowedDNSSuffixes restricts certificate DNS names to suffixes that cannot
// collide with real public domains. This prevents the private CA from being
// misused to issue certificates for public domains when added to a trust store.
//
// The structural defense relies on the invariant that **no current or future
// public TLD will end in any of these strings**. RFC-reserved TLDs satisfy this
// by definition. The K8s entries (.svc, .cluster.local) are not IANA-reserved
// — they are conventions of Kubernetes — but ICANN policy avoids delegating
// gTLDs that conflict with widely deployed internal names, so treating them as
// effectively reserved is reasonable for a per-developer dev CA. Do **not**
// add suffixes here that could plausibly be delegated as a public TLD (e.g.
// .corp, .home, .lan); that would break the substring-match defense.
//
// Sources:
//   - .localhost     — RFC 6761 (loopback)
//   - .test          — RFC 2606 (testing)
//   - .example       — RFC 2606 (documentation)
//   - .invalid       — RFC 2606 (guaranteed non-resolvable)
//   - .local         — RFC 6762 (mDNS)
//   - .internal      — ICANN (2024, private/internal use)
//   - .home.arpa     — RFC 8375 (home networks)
//   - .svc           — Kubernetes service DNS convention
//   - .cluster.local — Kubernetes default cluster domain
var allowedDNSSuffixes = []string{
	".localhost",
	".test",
	".example",
	".invalid",
	".local",
	".internal",
	".home.arpa",
	".svc",
	".cluster.local",
}

// hasAllowedDNSSuffix checks whether a DNS name (without wildcard prefix) ends
// with one of the allowed reserved suffixes, or is exactly a bare reserved
// suffix. Single-label names (no dots) bypass this check via ValidateDNSNames
// and are handled separately.
func hasAllowedDNSSuffix(name string) bool {
	name = strings.TrimPrefix(name, "*.")
	lower := strings.ToLower(name)
	for _, suffix := range allowedDNSSuffixes {
		if lower == suffix[1:] || strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	return false
}

// labelsWithinLimit checks that every label in a DNS name is at most 63 characters (RFC 1035).
func labelsWithinLimit(name string) bool {
	name = strings.TrimPrefix(name, "*.")
	for _, label := range strings.Split(name, ".") {
		if len(label) > 63 {
			return false
		}
	}
	return true
}

// ValidateDNSNames checks that the given DNS names are well-formed and use
// either a single label (e.g. "my-service" for in-cluster K8s addressing) or
// a reserved suffix. The label parameter is used in error messages to identify
// the source.
//
// Wildcard rule: a name "*.X" is rejected when X contains no dots (i.e. a
// wildcard over a single label). RFC 6125 §6.4.3 disallows wildcards in the
// rightmost label, and a wildcard over a bare label has unbounded scope.
func ValidateDNSNames(dnsNames []string, label string) error {
	if len(dnsNames) == 0 {
		return fmt.Errorf("%s has empty dnsNames", label)
	}
	for i, name := range dnsNames {
		if name == "" {
			return fmt.Errorf("%s has empty dnsNames entry at index %d", label, i)
		}
		if len(name) > 253 {
			return fmt.Errorf("%s has dnsNames entry '%s' exceeding 253 characters", label, name)
		}
		if !dnsNameRegex.MatchString(name) {
			return fmt.Errorf("%s has invalid dnsNames entry '%s'", label, name)
		}
		if !labelsWithinLimit(name) {
			return fmt.Errorf(
				"%s has dnsNames entry '%s' with a label exceeding 63 characters",
				label, name,
			)
		}
		if err := validateWildcardScope(name, label); err != nil {
			return err
		}
		stripped := strings.TrimPrefix(name, "*.")
		if !strings.Contains(stripped, ".") {
			continue
		}
		if !hasAllowedDNSSuffix(name) {
			return fmt.Errorf(
				"%s has dnsNames entry '%s' with a non-reserved suffix; "+
					"allowed forms are a single label (e.g. 'my-service') or a name ending in "+
					".localhost, .test, .example, .invalid, .local, .internal, .home.arpa, "+
					".svc, or .cluster.local",
				label, name,
			)
		}
	}
	return nil
}

// validateWildcardScope rejects wildcards whose scope is a whole suffix root:
// a single-label wildcard like "*.svc", or a multi-label wildcard whose
// stripped form is exactly a recognized suffix (e.g. "*.cluster.local",
// "*.home.arpa"). Such wildcards expand across every name under the suffix,
// which is an unbounded SAN for a per-developer dev CA.
func validateWildcardScope(name, label string) error {
	if !strings.HasPrefix(name, "*.") {
		return nil
	}
	stripped := strings.TrimPrefix(name, "*.")
	if !strings.Contains(stripped, ".") || isAllowedDNSSuffixRoot(stripped) {
		return fmt.Errorf(
			"%s has dnsNames entry '%s' with a wildcard over too few labels; "+
				"a wildcard must scope to at least one label below a reserved suffix "+
				"(e.g. '*.foo.svc', not '*.svc' or '*.cluster.local')",
			label, name,
		)
	}
	return nil
}

// isAllowedDNSSuffixRoot reports whether a name (without leading dot) equals
// one of the entries in allowedDNSSuffixes.
func isAllowedDNSSuffixRoot(name string) bool {
	lower := strings.ToLower(name)
	for _, suffix := range allowedDNSSuffixes {
		if lower == suffix[1:] {
			return true
		}
	}
	return false
}

// Validate checks the certificate request for correctness.
func (c *CertificateRequest) Validate(serviceName, contextName string) error {
	if c.Type != CertificateTypeServer && c.Type != CertificateTypeClient {
		return fmt.Errorf(
			"certificate for service '%s' in context '%s' has invalid type '%s' (must be 'server' or 'client')",
			serviceName, contextName, c.Type,
		)
	}

	label := fmt.Sprintf("certificate for service '%s' in context '%s'", serviceName, contextName)
	if err := ValidateDNSNames(c.DNSNames, label); err != nil {
		return err
	}

	if c.K8sSecret.Name == "" {
		return fmt.Errorf(
			"certificate for service '%s' in context '%s' has empty k8sSecret.name",
			serviceName, contextName,
		)
	}
	if len(c.K8sSecret.Name) > 253 || !k8sNameRegex.MatchString(c.K8sSecret.Name) {
		return fmt.Errorf(
			"certificate for service '%s' in context '%s' has invalid k8sSecret.name '%s' (must be a valid Kubernetes name: lowercase alphanumeric, hyphens, or dots)",
			serviceName, contextName, c.K8sSecret.Name,
		)
	}

	if c.K8sSecret.Type != K8sSecretTypeTLS && c.K8sSecret.Type != K8sSecretTypeOpaque {
		return fmt.Errorf(
			"certificate for service '%s' in context '%s' has invalid k8sSecret.type '%s' (must be 'tls' or 'opaque')",
			serviceName, contextName, c.K8sSecret.Type,
		)
	}

	if c.K8sSecret.Type == K8sSecretTypeTLS && c.K8sSecret.Keys != nil {
		return fmt.Errorf(
			"certificate for service '%s' in context '%s' must not specify keys when k8sSecret.type is 'tls'",
			serviceName, contextName,
		)
	}

	if c.K8sSecret.Type == K8sSecretTypeOpaque {
		if c.K8sSecret.Keys == nil {
			return fmt.Errorf(
				"certificate for service '%s' in context '%s' must specify keys when k8sSecret.type is 'opaque'",
				serviceName, contextName,
			)
		}
		if c.K8sSecret.Keys.PrivateKey == "" {
			return fmt.Errorf(
				"certificate for service '%s' in context '%s' has empty k8sSecret.keys.privateKey",
				serviceName, contextName,
			)
		}
		if !k8sSecretKeyRegex.MatchString(c.K8sSecret.Keys.PrivateKey) {
			return fmt.Errorf(
				"certificate for service '%s' in context '%s' has invalid k8sSecret.keys.privateKey '%s' (must match [-._a-zA-Z0-9]+)",
				serviceName, contextName, c.K8sSecret.Keys.PrivateKey,
			)
		}
		if c.K8sSecret.Keys.Cert == "" {
			return fmt.Errorf(
				"certificate for service '%s' in context '%s' has empty k8sSecret.keys.cert",
				serviceName, contextName,
			)
		}
		if !k8sSecretKeyRegex.MatchString(c.K8sSecret.Keys.Cert) {
			return fmt.Errorf(
				"certificate for service '%s' in context '%s' has invalid k8sSecret.keys.cert '%s' (must match [-._a-zA-Z0-9]+)",
				serviceName, contextName, c.K8sSecret.Keys.Cert,
			)
		}
		if c.K8sSecret.Keys.CA == "" {
			return fmt.Errorf(
				"certificate for service '%s' in context '%s' has empty k8sSecret.keys.ca",
				serviceName, contextName,
			)
		}
		if !k8sSecretKeyRegex.MatchString(c.K8sSecret.Keys.CA) {
			return fmt.Errorf(
				"certificate for service '%s' in context '%s' has invalid k8sSecret.keys.ca '%s' (must match [-._a-zA-Z0-9]+)",
				serviceName, contextName, c.K8sSecret.Keys.CA,
			)
		}
	}

	return nil
}
