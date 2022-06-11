package main

import (
	"context"
	"fmt"
	pb "github.com/dist1ll/cache-prototype/stock_sentiment"
	"github.com/dist1ll/cache-prototype/stock_sentiment/provider"
	"github.com/stretchr/testify/assert"
	grpc2 "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"testing"
	"time"
)

// setup starts the gRPC server
func setup(port int, ttl time.Duration) (*grpc2.Server, pb.StockSentimentClient, *grpc2.ClientConn, *provider.MockDataProvider) {
	prov := provider.MockDataProvider{}
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatal(err)
	}
	// start server
	grpcServer := grpc2.NewServer([]grpc2.ServerOption{}...)
	pb.RegisterStockSentimentServer(grpcServer,
		// configure cache TTL with mock provider
		pb.NewStockSentimentServer(ttl, &prov))

	// set mock data
	prov.SetData(`[
	{"sentiment":"Bullish","sentiment_score":0.13,"ticker":"GME"},
	{"sentiment":"Bullish","sentiment_score":0.259,"ticker":"IQ"},
	{"sentiment":"Bullish","sentiment_score":0.257,"ticker":"TSLA"},
	{"sentiment":"Bearish","sentiment_score":-0.227,"ticker":"EV"},
	{"sentiment":"Bullish","sentiment_score":0.151,"ticker":"TA"},
	{"sentiment":"Bullish","sentiment_score":0.026,"ticker":"UK"},
	{"sentiment":"Bearish","sentiment_score":-0.415,"ticker":"SQQQ"}
	]`)
	go grpcServer.Serve(lis)
	// transport security needs to be set
	dialOpt := grpc2.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc2.Dial(fmt.Sprintf("localhost:%d", port), []grpc2.DialOption{dialOpt}...)
	if err != nil {
		log.Fatal("Client: " + err.Error())
	}
	client := pb.NewStockSentimentClient(conn)
	return grpcServer, client, conn, &prov
}

// Test if server delivers basic request
func TestBasicRequest(t *testing.T) {
	server, client, conn, _ := setup(8100, time.Minute)
	defer conn.Close()
	defer server.Stop()

	sb, err := client.GetStockSentiments(context.Background(), &pb.RequestParams{})
	assert.Nil(t, err)
	assert.EqualValues(t, 7, len(sb.S))
}

// Test if server delivers cached request
func TestCachedRequest(t *testing.T) {
	server, client, conn, prov := setup(8101, time.Minute)
	defer conn.Close()
	defer server.Stop()

	sb, err := client.GetStockSentiments(context.Background(), &pb.RequestParams{})
	assert.Nil(t, err)
	assert.EqualValues(t, 7, len(sb.S))

	// block data provider, expect data to still be available because of cache
	prov.BlockFetch = true
	sb, err = client.GetStockSentiments(context.Background(), &pb.RequestParams{})
	assert.Nil(t, err)
	assert.EqualValues(t, 7, len(sb.S))

	// but not if the request params are different
	prov.BlockFetch = true
	var date = "2020-02-02"
	sb, err = client.GetStockSentiments(context.Background(), &pb.RequestParams{Date: &date})
	assert.NotNil(t, err)
}

// Test if server correctly fails if data provider is not available on empty cache
func TestProviderNotAvailable(t *testing.T) {
	server, client, conn, prov := setup(8102, time.Minute)
	defer conn.Close()
	defer server.Stop()

	prov.BlockFetch = true
	_, err := client.GetStockSentiments(context.Background(), &pb.RequestParams{})
	assert.NotNil(t, err)
}

// Test if the server evicts stale cache after TTL
func TestCacheEviction(t *testing.T) {
	server, client, conn, prov := setup(8102, time.Millisecond*100)
	defer conn.Close()
	defer server.Stop()

	// populate cache
	sb, _ := client.GetStockSentiments(context.Background(), &pb.RequestParams{})
	assert.EqualValues(t, 7, len(sb.S))

	// wait until ttl expires
	time.Sleep(time.Millisecond * 400)

	// ask for data, expect server to reach out to provider and fail
	prov.BlockFetch = true
	_, err := client.GetStockSentiments(context.Background(), &pb.RequestParams{})
	assert.NotNil(t, err)
}
