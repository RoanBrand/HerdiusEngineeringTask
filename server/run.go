package main

import (
	"errors"
	"log"
	"net"

	pb "github.com/RoanBrand/HerdiusEngineeringTask/proto"
	"google.golang.org/grpc"
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
	l, err := net.Listen("tcp", "localhost:4596")
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMaxNumberServer(grpcServer, newServer())
	return grpcServer.Serve(l) // TODO: TLS?
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
