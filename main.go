package main

import (
	"flag"
	"fmt"
	pb "github.com/dist1ll/kaspa/stock_sentiment"
	"github.com/dist1ll/kaspa/stock_sentiment/provider"
	grpc2 "google.golang.org/grpc"
	"log"
	"net"
	"time"
)

var CFG_HOST string
var CFG_PORT int
var CFG_TTL int

func init() {
	flag.StringVar(&CFG_HOST, "host", "0.0.0.0", "The address of this server. Default 0.0.0.0")
	flag.IntVar(&CFG_PORT, "port", 8100, "The port of this server. Default: 8100")
	flag.IntVar(&CFG_TTL, "ttl", 15, "The TTL for client request caching in minutes. Default: 15 minutes")
}

func main() {
	flag.Parse()

	log.Println("Starting Server.")
	log.Printf("Host: %s", CFG_HOST)
	log.Printf("Port: %d", CFG_PORT)
	log.Printf("TTL:  %d Minutes", CFG_TTL)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", CFG_HOST, CFG_PORT))
	if err != nil {
		log.Fatal(err)
	}

	// start server
	grpcServer := grpc2.NewServer([]grpc2.ServerOption{}...)
	pb.RegisterStockSentimentServer(grpcServer,
		// configure cache TTL with Reddit WSB provider
		pb.NewStockSentimentServer(time.Minute*time.Duration(CFG_TTL), provider.NewRedditWSBProvider()))

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}
