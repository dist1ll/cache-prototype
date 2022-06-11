## Read-through TTL cache with gRPC (Prototype)

This module provides a data-provider-agnostic TTL cache for market sentiment data. To run, specify
host, port number and the TTL of client request caching:

```
./main --host=<ip> --port=<port> --ttl=<ttl in minutes>
```

For an example on how to set up the server, see the unit tests in `stock_sentiment_test.go` or and the example
request with a real data provider.

#### Testing

Unit tests with a mocked data provider are defined in `stock_sentiment_test.go`. 

```
go test -v
```

## Structure

For an example on how to set up the server, see `Example_SendTwoRequests` in `examples/basic_requests.go`. 

The gRPC server `NewStockSentimentServer(ttl time.Duration, provider DataProvider)` creates a `StockSentimentServer` that 
aggregates data from a given `DataProvider`. At the moment, only 1 data provider is supported. Additional data providers
can be implemented via the following interface: 

```go
// DataProvider is an abstraction over third-party providers of stock
// sentiment data.
type DataProvider interface {
	// Fetch returns a batch of stocks according to the request parameters.
	Fetch(ctx context.Context, params *RequestParams) (*StockBatch, error)
}
```

Note that like for all handler functions, the deadline and cancel signals of `ctx` must be respected.

## Service Definition

```protobuf
syntax = "proto3";
option go_package = "./";

// Stock contains info about a particular stock and its tracked sentiment.
message Stock {
  string ticker = 1;
  enum Sentiment {
    Bullish = 0;
    Bearish = 1;
  }
  // TODO: sentiment is derived from score, maybe remove redundancy?
  Sentiment sentiment = 2;
  float sentiment_score = 3;
}

// StockBatch stores an array of Stock objects and the corresponding
// date of their recording. Stocks objects are grouped by their
// date, which results in a batch.
message StockBatch {
  string date = 1;
  repeated Stock s = 2;
}

// RequestParams for fetching stock sentiments from StockSentiment service
message RequestParams {
  // If date is not specified, fetch the most recent batch of stocks
  optional string date = 1;
}

// StockSentiment returns sentiment information about stocks, aggregated from
// third-party sources. (At the moment: Tradestie -> wsb@reddit)
service StockSentiment {
  // GetStockSentiments returns a batch of stocks and their sentiment scores.
  rpc GetStockSentiments(RequestParams) returns (StockBatch) {}
}
```

