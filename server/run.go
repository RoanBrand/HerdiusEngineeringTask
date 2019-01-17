package main

import (
	"log"
	"net"
	"sync"

	"github.com/RoanBrand/HerdiusEngineeringTask/auth"
	pb "github.com/RoanBrand/HerdiusEngineeringTask/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type server struct {
	clients map[string]int64 // client's pub key -> client's max
	cLock   sync.RWMutex
}

func newServer() *server {
	return &server{clients: make(map[string]int64, 2)}
}

func startServer() error {
	tlsConfig, err := auth.LoadServerTLS()
	if err != nil {
		return err
	}

	l, err := net.Listen("tcp", "localhost:4596")
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{grpc.Creds(credentials.NewTLS(tlsConfig))}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterMaxNumberServer(grpcServer, newServer())

	return grpcServer.Serve(l)
}

// Serves client, provided the new number and client ID.
// Returns true if new number is the new max for client.
func (s *server) serveClient(num int64, cPubKey string) bool {
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

	log.Println("RX from client:", num)

	if num > currMax {
		s.cLock.RLock() // Only need write lock when changing map, not updating key's value.
		s.clients[cPubKey] = num
		s.cLock.RUnlock()
		return true
	}

	return false
}

func (s *server) FindMaxNumber(stream pb.MaxNumber_FindMaxNumberServer) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" { // don't want to import io just for this
				log.Println("client closed connection")
				return nil
			}

			log.Println("error RX from client:", err)
			return err
		}

		key, err := auth.ValidateClient(stream.Context())
		if err != nil {
			log.Println(err)
			return err
		}

		if s.serveClient(in.In, string(key)) {
			err := stream.Send(&pb.Response{Max: in.In})
			if err != nil {
				log.Println("error TX to client:", err)
				return err
			}
		}
	}
}

func main() {
	log.Fatal(startServer())
}
