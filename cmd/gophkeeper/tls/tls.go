package tls

import (
	"crypto/tls"
	_ "embed"
)

//go:embed cert.pem
var tlsCert []byte

//go:embed key.pem
var tlsKey []byte

func GetTLSConfig() (*tls.Config, error) {
	cer, err := tls.X509KeyPair(tlsCert, tlsKey)
	if err != nil {
		return nil, err
	}

	return &tls.Config{Certificates: []tls.Certificate{cer}}, nil
}
