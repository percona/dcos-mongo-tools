// Copyright 2018 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

const (
	dialMongodbTimeout = 3 * time.Second
	syncMongodbTimeout = 1 * time.Minute
)

type SSLConfig struct {
	Enabled            bool
	PEMKeyFile         string
	CAFile             string
	HostnameValidation bool
}

func (cnf *Config) loadCaCertificate() (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(cnf.SSL.CAFile)
	if err != nil {
		return nil, err
	}
	certificates := x509.NewCertPool()
	certificates.AppendCertsFromPEM(caCert)
	return certificates, nil
}

func (cnf *Config) configureSSLDialInfo() error {
	config := &tls.Config{
		InsecureSkipVerify: !cnf.SSL.HostnameValidation,
	}
	if len(cnf.SSL.PEMKeyFile) > 0 {
		log.Debugf("Loading SSL/TLS PEM certificate: %s", cnf.SSL.PEMKeyFile)
		certificates, err := tls.LoadX509KeyPair(cnf.SSL.PEMKeyFile, cnf.SSL.PEMKeyFile)
		if err != nil {
			return fmt.Errorf(
				"Cannot load key pair from '%s' to connect to server '%s'. Got: %v",
				cnf.SSL.PEMKeyFile,
				cnf.DialInfo.Addrs,
				err,
			)
		}
		config.Certificates = []tls.Certificate{certificates}
	}
	if len(cnf.SSL.CAFile) > 0 {
		log.Debugf("Loading SSL/TLS Certificate Authority: %s", cnf.SSL.PEMKeyFile)
		ca, err := cnf.loadCaCertificate()
		if err != nil {
			return fmt.Errorf("Couldn't load client CAs from %s. Got: %s", cnf.SSL.CAFile, err)
		}
		config.RootCAs = ca
	}
	cnf.DialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := tls.Dial("tcp", addr.String(), config)
		if err != nil {
			log.Errorf("Could not connect to %v. Got: %v", addr, err)
			return nil, err
		}
		if config.InsecureSkipVerify {
			err = validateConnection(conn, config)
			if err != nil {
				log.Errorf("Could not disable hostname validation. Got: %v", err)
			}
		}
		return conn, err
	}
	return nil
}

func validateConnection(conn *tls.Conn, tlsConfig *tls.Config) error {
	var err error
	if err = conn.Handshake(); err != nil {
		conn.Close()
		return err
	}

	opts := x509.VerifyOptions{
		Roots:         tlsConfig.RootCAs,
		CurrentTime:   time.Now(),
		DNSName:       "",
		Intermediates: x509.NewCertPool(),
	}

	certs := conn.ConnectionState().PeerCertificates
	for i, cert := range certs {
		if i == 0 {
			continue
		}
		opts.Intermediates.AddCert(cert)
	}

	_, err = certs[0].Verify(opts)
	if err != nil {
		conn.Close()
		return err
	}

	return nil
}