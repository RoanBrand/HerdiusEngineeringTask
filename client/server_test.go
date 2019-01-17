package main

import (
	"context"
	"testing"

	pb "github.com/RoanBrand/HerdiusEngineeringTask/proto"
)

func TestServer_FindMaxNumber(t *testing.T) {
	conn, err := dialServer()
	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()

	cl := pb.NewMaxNumberClient(conn)

	str, err := cl.FindMaxNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	expectedSequence := []int64{1, 5, 6, 20}
	go func() {
		i := 0
		for {
			in, err := str.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					close(done)
					return
				}
				t.Log("error RX from server:", err)
			}
			if in.Max != expectedSequence[i] {
				t.Fatal("expected sequence should be", expectedSequence)
			}
			i++
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
