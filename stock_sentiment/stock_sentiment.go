package __

import (
	"context"
	"errors"
	"github.com/jellydator/ttlcache/v3"
	"time"
)

// DataProvider is an abstraction over third-party providers of stock
// sentiment data.
type DataProvider interface {
	// Fetch returns a batch of stocks according to the request parameters.
	Fetch(ctx context.Context, params *RequestParams) (*StockBatch, error)
}

// StockSentimentImpl is our main implementation of the StockSentimentService from the proto spec.
type StockSentimentImpl struct {
	Timeout   time.Duration
	data      *ttlcache.Cache[string, StockBatch]
	providers []DataProvider
}

// NewStockSentimentServer creates a StockSentimentServer with a TTL request cache. Aggregates
// data from single providers.
func NewStockSentimentServer(ttl time.Duration, provider DataProvider) StockSentimentServer {
	return &StockSentimentImpl{
		providers: []DataProvider{provider},
		data: ttlcache.New[string, StockBatch](
			ttlcache.WithTTL[string, StockBatch](ttl),
		),
	}
}

func (s *StockSentimentImpl) GetStockSentiments(ctx context.Context, req *RequestParams) (*StockBatch, error) {
	// check cache data
	item := s.data.Get(req.GetDateOrNow())
	if item != nil {
		var val StockBatch = item.Value()
		return &val, nil
	}
	// get fresh data and insert into cache
	latest, err := s.pullFreshData(ctx, req)
	if err != nil {
		return nil, errors.New("problem fetching latest sentiment data")
	}
	s.data.Set(req.GetDateOrNow(), *latest, ttlcache.DefaultTTL)
	return latest, nil
}

// pullFreshData returns the latest StockBatch from third-party sources.
func (s *StockSentimentImpl) pullFreshData(ctx context.Context, params *RequestParams) (*StockBatch, error) {
	// TODO: Average/Aggregate(?) over several different data sources?
	return s.providers[0].Fetch(ctx, params)
}

func (s *StockSentimentImpl) mustEmbedUnimplementedStockSentimentServer() {}

// GetDateOrNow returns the date stamp in the object, or if empty returns the
// current time as a date string. Format: YYYY-MM-DD
func (x *RequestParams) GetDateOrNow() string {
	// TODO: if intended to unit test reliably, use mocked time abstraction
	if x.Date == nil {
		return time.Now().Format("2006-01-02")
	}
	return x.GetDate()
}

// SentimentFromString converts a string sentiment identifier ('Bullish' or 'Bearish') into
// the correct sentiment type.
func SentimentFromString(sentiment string) (Stock_Sentiment, error) {
	if sentiment == "Bullish" {
		return Stock_Bullish, nil
	} else if sentiment == "Bearish" {
		return Stock_Bearish, nil
	} else {
		return -1, errors.New("encountered incorrect grammar for sentiment field. Needs to be either 'Bullish' or 'Bearish'")
	}
}
