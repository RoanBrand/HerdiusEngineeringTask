package main

import (
	"context"
	"log"

	pb "github.com/RoanBrand/HerdiusEngineeringTask/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:4596", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	cl := pb.NewMaxNumberClient(conn)

	str, err := cl.FindMaxNumber(context.Background())
	if err != nil {
		log.Fatal(err)
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
}

func sendNumber(stream pb.MaxNumber_FindMaxNumberClient, n int64) {
	err := stream.Send(&pb.Request{In: n})
	if err != nil {
		log.Fatal(err)
	}
}
