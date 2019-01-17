package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"sync"

	pb "github.com/RoanBrand/HerdiusEngineeringTask/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

type server struct {
	clients map[string]int64 // client's pub key -> client's max
	cLock   sync.RWMutex
}

func newServer() *server {
	return &server{clients: make(map[string]int64, 2)}
}

func startServer() error {
	certificate, err := tls.LoadX509KeyPair(
		"cert/server/localhost.crt",
		"cert/server/localhost.key",
	)

	certPool := x509.NewCertPool()
	caF, err := ioutil.ReadFile("cert/MaxNumberRootCA.crt")
	if err != nil {
		return errors.New("failed to load CA cert: " + err.Error())
	}

	if ok := certPool.AppendCertsFromPEM(caF); !ok {
		return errors.New("failed to append cert")
	}

	l, err := net.Listen("tcp", "localhost:4596")
	if err != nil {
		return err
	}

	tlsConf := tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	opts := []grpc.ServerOption{grpc.Creds(credentials.NewTLS(&tlsConf))}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterMaxNumberServer(grpcServer, newServer())

	return grpcServer.Serve(l)
}

func (s *server) FindMaxNumber(stream pb.MaxNumber_FindMaxNumberServer) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" { // don't want to import io just for this
				log.Println("client", 1, "closed connection")
				return nil
			}

			log.Println("error RX from client", 1, ":", err)
			return err
		}

		errMsg := errors.New("error validating client")
		peer, ok := peer.FromContext(stream.Context())
		if !ok {
			log.Println(errMsg)
			return errMsg
		}

		tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)
		if !ok {
			log.Println(errMsg)
			return errMsg
		}

		cCerts := tlsInfo.State.PeerCertificates
		if len(cCerts) == 0 {
			log.Println(errMsg)
			return errMsg
		}

		key, _ := x509.MarshalPKIXPublicKey(cCerts[0].PublicKey)
		cPubKey := string(key)

		s.cLock.RLock()
		currMax, ok := s.clients[cPubKey]
		s.cLock.RUnlock()
		if !ok {
			log.Println("new client")
			s.cLock.Lock()
			s.clients[cPubKey] = 0
			s.cLock.Unlock()
			currMax = 0
		}

		log.Println("RX from client", 1, ":", in.In)

		if in.In > currMax {
			s.cLock.RLock() // Only need write lock when changing map, not updating key's value.
			s.clients[cPubKey] = in.In
			s.cLock.RUnlock()
			err := stream.Send(&pb.Response{Max: in.In})
			if err != nil {
				log.Println("error TX to client", 1, ":", err)
				return err
			}
		}
	}
}

func main() {
	log.Fatal(startServer())
}
