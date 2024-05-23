package configbase

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"
	"time"

	"github.com/lucky-xin/nebula-importer/pkg/client"
	"github.com/lucky-xin/nebula-importer/pkg/errors"
	"github.com/lucky-xin/nebula-importer/pkg/utils"
)

var newClientPool = client.NewPool

const (
	ClientVersion3       = "v3"
	ClientVersionDefault = ClientVersion3
)

type (
	Client struct {
		Version                  string        `yaml:"version" json:"version"`
		Address                  string        `yaml:"address" json:"address"`
		User                     string        `yaml:"user,omitempty" json:"user,omitempty,optional"`
		Password                 string        `yaml:"password,omitempty" json:"password,omitempty,optional"`
		ConcurrencyPerAddress    int           `yaml:"concurrencyPerAddress,omitempty" json:"concurrencyPerAddress,omitempty,optional"`
		ReconnectInitialInterval time.Duration `yaml:"reconnectInitialInterval,omitempty" json:"reconnectInitialInterval,omitempty,optional"`
		Retry                    int           `yaml:"retry,omitempty" json:"retry,omitempty,optional,default=3"`
		RetryInitialInterval     time.Duration `yaml:"retryInitialInterval,omitempty" json:"retryInitialInterval,omitempty,optional,default=200"`
		SSL                      *SSL          `yaml:"ssl,omitempty" json:"ssl,omitempty,optional"`
	}

	SSL struct {
		Enable             bool   `yaml:"enable,omitempty" json:"enable,omitempty,optional"`
		CertPath           string `yaml:"certPath,omitempty" json:"certPath,omitempty,optional"`
		KeyPath            string `yaml:"keyPath,omitempty" json:"keyPath,omitempty,optional"`
		CAPath             string `yaml:"caPath,omitempty" json:"caPath,omitempty,optional"`
		InsecureSkipVerify bool   `yaml:"insecureSkipVerify,omitempty" json:"insecureSkipVerify,omitempty,optional"`
	}
)

// OptimizePath optimizes relative paths base to the configuration file path
func (c *Client) OptimizePath(configPath string) error {
	if c == nil {
		return nil
	}

	if c.SSL != nil && c.SSL.Enable {
		configPathDir := filepath.Dir(configPath)
		c.SSL.CertPath = utils.RelativePathBaseOn(configPathDir, c.SSL.CertPath)
		c.SSL.KeyPath = utils.RelativePathBaseOn(configPathDir, c.SSL.KeyPath)
		c.SSL.CAPath = utils.RelativePathBaseOn(configPathDir, c.SSL.CAPath)
	}

	return nil
}

func (c *Client) BuildClientPool(opts ...client.Option) (client.Pool, error) {
	if c.Version == "" {
		c.Version = ClientVersion3
	}
	tlsConfig, err := c.SSL.BuildConfig()
	if err != nil {
		return nil, err
	}

	options := make([]client.Option, 0, 8+len(opts))
	options = append(
		options,
		client.WithAddress(c.Address),
		client.WithUserPassword(c.User, c.Password),
		client.WithTLSConfig(tlsConfig),
		client.WithReconnectInitialInterval(c.ReconnectInitialInterval),
		client.WithRetry(c.Retry),
		client.WithRetryInitialInterval(c.RetryInitialInterval),
		client.WithConcurrencyPerAddress(c.ConcurrencyPerAddress),
	)
	switch c.Version {
	case ClientVersion3:
		options = append(options, client.WithV3())
	default:
		return nil, errors.ErrUnsupportedClientVersion
	}
	options = append(options, opts...)
	pool := newClientPool(options...)
	return pool, nil
}

func (s *SSL) BuildConfig() (*tls.Config, error) {
	if s == nil || !s.Enable {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		Certificates:       make([]tls.Certificate, 1),
		InsecureSkipVerify: s.InsecureSkipVerify, //nolint:gosec
	}

	rootPEM, err := os.ReadFile(s.CAPath)
	if err != nil {
		return nil, err
	}

	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(rootPEM)
	tlsConfig.RootCAs = rootCAs

	cert, err := tls.LoadX509KeyPair(s.CertPath, s.KeyPath)
	if err != nil {
		return nil, err
	}
	tlsConfig.Certificates[0] = cert

	return tlsConfig, nil
}
