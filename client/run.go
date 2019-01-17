package main

import (
	"context"
	"log"

	"github.com/RoanBrand/HerdiusEngineeringTask/auth"
	pb "github.com/RoanBrand/HerdiusEngineeringTask/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func dialServer() (*grpc.ClientConn, error) {
	tlsConfig, err := auth.LoadClientTLS()
	if err != nil {
		return nil, err
	}

	creds := credentials.NewTLS(tlsConfig)
	conn, err := grpc.Dial("localhost:4596", grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func performRequests(conn *grpc.ClientConn) error {
	cl := pb.NewMaxNumberClient(conn)

	str, err := cl.FindMaxNumber(context.Background())
	if err != nil {
		return err
	}

	done := make(chan struct{})
	go func() {
		for {
			in, err := str.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					close(done)
					return
				}
				log.Println("error RX from server:", err)
			}
			log.Println("client RX new max:", in.Max)
		}
	}()

	sendNumber(str, 1)
	sendNumber(str, 5)
	sendNumber(str, 3)
	sendNumber(str, 6)
	sendNumber(str, 2)
	sendNumber(str, 20)

	str.CloseSend()
	<-done
	return nil
}

func sendNumber(stream pb.MaxNumber_FindMaxNumberClient, n int64) {
	err := stream.Send(&pb.Request{In: n})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	conn, err := dialServer()
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	if err := performRequests(conn); err != nil {
		log.Fatal(err)
	}
}
