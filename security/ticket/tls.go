// Copyright 2018 GRAIL, Inc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package ticket

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/grailbio/base/security/keycrypt"
	"github.com/grailbio/base/security/tls/certificateauthority"
	"v.io/x/lib/vlog"
)

const driftMargin = 10 * time.Minute

func (b *TlsCertAuthorityBuilder) newTlsClientTicket() (TicketTlsClientTicket, error) {
	tlsCredentials, err := b.genTlsCredentials()

	if err != nil {
		return TicketTlsClientTicket{}, err
	}

	return TicketTlsClientTicket{
		Value: TlsClientTicket{
			Credentials: tlsCredentials,
		},
	}, nil
}

func (b *TlsCertAuthorityBuilder) newTlsServerTicket() (TicketTlsServerTicket, error) {
	tlsCredentials, err := b.genTlsCredentials()

	if err != nil {
		return TicketTlsServerTicket{}, err
	}

	return TicketTlsServerTicket{
		Value: TlsServerTicket{
			Credentials: tlsCredentials,
		},
	}, nil
}

func (b *TlsCertAuthorityBuilder) newDockerTicket() (TicketDockerTicket, error) {
	tlsCredentials, err := b.genTlsCredentials()

	if err != nil {
		return TicketDockerTicket{}, err
	}

	return TicketDockerTicket{
		Value: DockerTicket{
			Credentials: tlsCredentials,
		},
	}, nil
}

func (b *TlsCertAuthorityBuilder) newDockerServerTicket() (TicketDockerServerTicket, error) {
	tlsCredentials, err := b.genTlsCredentialsWithKeyUsage([]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})

	if err != nil {
		return TicketDockerServerTicket{}, err
	}

	return TicketDockerServerTicket{
		Value: DockerServerTicket{
			Credentials: tlsCredentials,
		},
	}, nil
}

func (b *TlsCertAuthorityBuilder) newDockerClientTicket() (TicketDockerClientTicket, error) {
	tlsCredentials, err := b.genTlsCredentialsWithKeyUsage([]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})

	if err != nil {
		return TicketDockerClientTicket{}, err
	}

	return TicketDockerClientTicket{
		Value: DockerClientTicket{
			Credentials: tlsCredentials,
		},
	}, nil
}

func (b *TlsCertAuthorityBuilder) genTlsCredentials() (TlsCredentials, error) {
	return b.genTlsCredentialsWithKeyUsage([]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth})
}

func (b *TlsCertAuthorityBuilder) genTlsCredentialsWithKeyUsage(keyUsage []x509.ExtKeyUsage) (TlsCredentials, error) {
	vlog.Infof("TlsCertAuthorityBuilder: %+v", b)
	empty := TlsCredentials{}

	secret, err := keycrypt.Lookup(b.Authority)
	if err != nil {
		return empty, err
	}
	authority := certificateauthority.CertificateAuthority{DriftMargin: driftMargin, Signer: secret}
	if err := authority.Init(); err != nil {
		return empty, err
	}
	ttl := time.Duration(b.TtlSec) * time.Second
	cert, key, err := authority.IssueWithKeyUsage(b.CommonName, ttl, nil, b.San, keyUsage)
	if err != nil {
		return empty, err
	}

	r := TlsCredentials{}
	r.AuthorityCert, err = encode(&pem.Block{Type: "CERTIFICATE", Bytes: authority.Cert.Raw})
	if err != nil {
		return empty, err
	}
	r.Cert, err = encode(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	if err != nil {
		return empty, err
	}
	r.Key, err = encode(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err != nil {
		return empty, err
	}
	return r, nil
}

func encode(block *pem.Block) (string, error) {
	var w bytes.Buffer
	if err := pem.Encode(&w, block); err != nil {
		return "", err
	}
	return w.String(), nil
}
