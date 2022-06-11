package provider

import (
	"context"
	"encoding/json"
	pb "github.com/dist1ll/cache-prototype/stock_sentiment"
	"io"
	"net/http"
	"time"
)

// RedditWSBProvider provides market sentiment data sourced from comments in
// the wallstreetbets subreddit. Sentiment is taken from Tradestie API.
type redditWSBProvider struct{}

func NewRedditWSBProvider() pb.DataProvider {
	return &redditWSBProvider{}
}

// TradestieWSB_Stock is the json schema of a single entry of the Tradestie API.
// Example: https://tradestie.com/api/v1/apps/reddit?date=2022-06-11
type TradestieWSB_Stock struct {
	Comments       int `json:"no_of_comments"`
	Sentiment      string
	SentimentScore float32 `json:"sentiment_score"`
	Ticker         string
}

func (p *redditWSBProvider) Fetch(ctx context.Context, params *pb.RequestParams) (*pb.StockBatch, error) {
	client := &http.Client{Timeout: time.Second * 5}
	req, err := http.NewRequest("GET", "https://tradestie.com/api/v1/apps/reddit?data="+params.GetDateOrNow(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req = req.WithContext(ctx)
	// GET request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// convert
	var out []TradestieWSB_Stock
	err = json.Unmarshal(b, &out)
	if err != nil {
		return nil, err
	}
	return normalizeWSBData(out)
}

// normalizeWSBData transforms a slice of TradestieWSB_Stock objects into a proper StockBatch
func normalizeWSBData(arr []TradestieWSB_Stock) (*pb.StockBatch, error) {
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
