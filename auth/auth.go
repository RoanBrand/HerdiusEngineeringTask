package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"io/ioutil"
)

func LoadServerTLS() (*tls.Config, error) {
	certificate, err := tls.LoadX509KeyPair(
		"cert/server/localhost.crt",
		"cert/server/localhost.key",
	)
	if err != nil {
		return nil, err
	}

	certPool, err := loadRootCA()
	if err != nil {
		return nil, err
	}

	tlsConf := tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	return &tlsConf, nil
}

func LoadClientTLS() (*tls.Config, error) {
	certificate, err := tls.LoadX509KeyPair(
		"cert/client/localhost.crt",
		"cert/client/localhost.key",
	)
	if err != nil {
		return nil, err
	}

	certPool, err := loadRootCA()
	if err != nil {
		return nil, err
	}

	tlsConf := tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
		ServerName:   "localhost",
	}

	return &tlsConf, nil
}

// Returns client PubKey for ID.
func ValidateClient(streamCtx context.Context) ([]byte, error) {
	errMsg := errors.New("error validating client")
	peer, ok := peer.FromContext(streamCtx)
	if !ok {
		return nil, errMsg
	}

	tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, errMsg
	}

	cCerts := tlsInfo.State.PeerCertificates
	if len(cCerts) == 0 {
		return nil, errMsg
	}

	return x509.MarshalPKIXPublicKey(cCerts[0].PublicKey)
}

func loadRootCA() (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	caF, err := ioutil.ReadFile("cert/MaxNumberRootCA.crt")
	if err != nil {
		return nil, errors.New("failed to load CA cert: " + err.Error())
	}

	if ok := certPool.AppendCertsFromPEM(caF); !ok {
		return nil, errors.New("failed to append cert")
	}

	return certPool, nil
}
