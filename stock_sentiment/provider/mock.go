package provider

import (
	"context"
	"encoding/json"
	"errors"
	pb "github.com/dist1ll/cache-prototype/stock_sentiment"
)

// MockDataProvider is a mocking data provider for unit testing our sentiment service
type MockDataProvider struct {
	// If BlockFetch is true, returns an error on fetch.
	BlockFetch bool
	Data       *pb.StockBatch
	// Number of times Fetch has been called
	FetchCounter int
}

type mockStock struct {
	Sentiment      string
	SentimentScore float32
	Ticker         string
}

// SetData sets the data that the MockDataProvider should return on Fetch requests.
// jsonData is an array with Stock objects, according to the stock_sentiment.proto.
func (m *MockDataProvider) SetData(jsonData string) {
	var stocks []mockStock
	err := json.Unmarshal([]byte(jsonData), &stocks)
	if err != nil {
		panic(err.Error())
	}
	sb, err := convert(stocks)
	if err != nil {
		panic(err.Error())
	}
	m.Data = sb
}

// convert an array of mockStocks to a StockBatch object
func convert(arr []mockStock) (*pb.StockBatch, error) {
	sb := pb.StockBatch{}
	sb.S = make([]*pb.Stock, len(arr))
	for i, val := range arr {
		sentiment, err := pb.SentimentFromString(val.Sentiment)
		if err != nil {
			return nil, err
		}
		sb.S[i] = &pb.Stock{
			Ticker:         val.Ticker,
			Sentiment:      sentiment,
			SentimentScore: val.SentimentScore,
		}
	}
	return &sb, nil
}

func (m *MockDataProvider) Fetch(ctx context.Context, params *pb.RequestParams) (*pb.StockBatch, error) {
	m.FetchCounter += 1
	if m.BlockFetch {
		return nil, errors.New("couldn't access data provider")
	}
	return m.Data, nil
}
