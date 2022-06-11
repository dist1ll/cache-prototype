package examples

import (
	"context"
	"fmt"
	pb "github.com/dist1ll/cache-prototype/stock_sentiment"
	"github.com/dist1ll/cache-prototype/stock_sentiment/provider"
	grpc2 "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"time"
)

func Example_SendTwoRequests() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8100))
	if err != nil {
		log.Fatal(err)
	}
	// start client
	go DemoClient()

	// start server
	grpcServer := grpc2.NewServer([]grpc2.ServerOption{}...)
	pb.RegisterStockSentimentServer(grpcServer,
		// configure cache TTL with Reddit WSB provider
		pb.NewStockSentimentServer(time.Minute*15, provider.NewRedditWSBProvider()))

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal("Server: " + err.Error())
	}
}

// DemoClient demos the service by requesting data twice. The second retrieval is read from
// the cache, which should be apparent from the speed of the console output.
func DemoClient() {
	time.Sleep(time.Millisecond * 500)
	// transport security needs to be set
	dialOpt := grpc2.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc2.Dial("localhost:8100", []grpc2.DialOption{dialOpt}...)
	if err != nil {
		log.Fatal("Client: " + err.Error())
	}
	defer conn.Close()

	client := pb.NewStockSentimentClient(conn)

	var t = "2022-06-10"
	println("request started")
	sent, err := client.GetStockSentiments(context.Background(), &pb.RequestParams{Date: &t})
	if err != nil {
		log.Fatal("Fetch: " + err.Error())
	}
	fmt.Println(sent.S)
	time.Sleep(time.Second)
	println("request started")
	sent, err = client.GetStockSentiments(context.Background(), &pb.RequestParams{Date: &t})
	if err != nil {
		log.Fatal("Fetch: " + err.Error())
	}
	fmt.Println(sent.S)
}
