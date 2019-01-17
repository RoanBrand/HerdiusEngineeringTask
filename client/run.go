package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

	pb "github.com/RoanBrand/HerdiusEngineeringTask/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	certificate, err := tls.LoadX509KeyPair(
		"cert/client/localhost.crt",
		"cert/client/localhost.key",
	)

	certPool := x509.NewCertPool()
	caF, err := ioutil.ReadFile("cert/MaxNumberRootCA.crt")
	if err != nil {
		log.Fatal("failed to load CA cert:", err)
	}

	if ok := certPool.AppendCertsFromPEM(caF); !ok {
		log.Fatal("failed to append cert")
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
		ServerName:   "localhost",
	})

	conn, err := grpc.Dial("localhost:4596", grpc.WithTransportCredentials(creds))
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
