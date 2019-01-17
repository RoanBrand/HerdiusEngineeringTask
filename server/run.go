package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"
	"net"

	pb "github.com/RoanBrand/HerdiusEngineeringTask/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var knownClients = []string{"1"}

type server struct {
	clients map[string]int64 // client's pub key -> client's max
	// seems no locks needed as we will only read once populated and compare and update numbers synchronously.
}

func newServer() *server {
	s := server{clients: make(map[string]int64, len(knownClients))}
	for _, c := range knownClients {
		s.clients[c] = 0
	}

	return &s
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

		log.Println("RX from client", 1, ":", in.In)

		currMax, ok := s.clients["1"]
		if !ok {
			err := errors.New("error: unknown client")
			log.Println(err)
			return err
		}

		if in.In > currMax {
			s.clients["1"] = in.In
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
