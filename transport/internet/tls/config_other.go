//go:build !windows && !confonly
// +build !windows,!confonly

package tls

import (
	"crypto/x509"
	"sync"
)

var rootCerts rootCertsCache

type rootCertsCache struct {
	sync.Mutex
	pool *x509.CertPool
}

func (c *Config) getCertPool() (*x509.CertPool, error) {
	if c.DisableSystemRoot {
		return c.loadSelfCertPool()
	}
	if len(c.Certificate) == 0 {
		return rootCerts.load()
	}
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, newError("system root").AtWarning().Base(err)
	}
	for _, cert := range c.Certificate {
		if cert.Usage == Certificate_AUTHORITY_VERIFY_CLIENT {
			continue
		}

		if !pool.AppendCertsFromPEM(cert.Certificate) {
			return nil, newError("append cert to root").AtWarning().Base(err)
		}
	}
	return pool, err
}

func (c *rootCertsCache) load() (*x509.CertPool, error) {
	c.Lock()
	defer c.Unlock()
	if c.pool != nil {
		return c.pool, nil
	}
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	c.pool = pool
	return pool, nil
}
