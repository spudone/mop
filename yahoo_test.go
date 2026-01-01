package mop

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// MockTransport implements http.RoundTripper
type MockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestMarketFetch(t *testing.T) {
	// Mock http.DefaultTransport
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()

	http.DefaultTransport = &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Route based on URL
			if strings.Contains(req.URL.String(), "getcrumb") {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("mock_crumb")),
					Header:     make(http.Header),
				}, nil
			}
			if strings.Contains(req.URL.String(), "finance/quote") {
				// Return market data
				response := `{
					"quoteResponse": {
						"result": [
							{"regularMarketChange": 10.5, "regularMarketPrice": 34000.0, "regularMarketChangePercent": 0.03},
							{"regularMarketChange": -50.0, "regularMarketPrice": 14000.0, "regularMarketChangePercent": -0.35},
							{"regularMarketChange": 5.0, "regularMarketPrice": 4400.0, "regularMarketChangePercent": 0.11},
							{"regularMarketChange": 100.0, "regularMarketPrice": 28000.0, "regularMarketChangePercent": 0.36},
							{"regularMarketChange": -200.0, "regularMarketPrice": 25000.0, "regularMarketChangePercent": -0.80},
							{"regularMarketChange": 15.0, "regularMarketPrice": 7500.0, "regularMarketChangePercent": 0.20},
							{"regularMarketChange": -10.0, "regularMarketPrice": 15000.0, "regularMarketChangePercent": -0.06},
							{"regularMarketChange": 0.05, "regularMarketPrice": 1.50, "regularMarketChangePercent": 3.33},
							{"regularMarketChange": 1.5, "regularMarketPrice": 75.0, "regularMarketChangePercent": 2.0},
							{"regularMarketChange": -0.5, "regularMarketPrice": 110.0, "regularMarketChangePercent": -0.45},
							{"regularMarketChange": 0.01, "regularMarketPrice": 1.10, "regularMarketChangePercent": 0.9},
							{"regularMarketChange": 10.0, "regularMarketPrice": 1800.0, "regularMarketChangePercent": 0.55}
						]
					}
				}`
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(response)),
					Header:     make(http.Header),
				}, nil
			}

			// Default (cookies)
			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("<html></html>")),
				Header:     make(http.Header),
				Request:    req, // Needed for cookies
			}
			resp.Header.Set("Set-Cookie", "A1=mock_a1_cookie; Path=/; Domain=.yahoo.com")

			return resp, nil
		},
	}

	market := NewMarket()
	market.Fetch()

	if !market.IsClosed {
		// Just a check, IsClosed default is false
	}

	if market.Dow.Latest != "34000.000" {
		t.Errorf("Dow Latest = %s; want 34000.000", market.Dow.Latest)
	}

	// Check one that was negative
	if market.Nasdaq.Change != "-50.000" {
		t.Errorf("Nasdaq Change = %s; want -50.000", market.Nasdaq.Change)
	}
}

func TestQuotesFetch(t *testing.T) {
	// Mock http.DefaultTransport
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()

	http.DefaultTransport = &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.String(), "getcrumb") {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("mock_crumb")),
					Header:     make(http.Header),
				}, nil
			}
			if strings.Contains(req.URL.String(), "finance/quote") {
				response := `{
					"quoteResponse": {
						"result": [
							{
								"symbol": "AAPL",
								"regularMarketPrice": 150.0,
								"regularMarketChange": 2.5,
								"regularMarketChangePercent": 1.6,
								"regularMarketOpen": 148.0,
								"regularMarketDayLow": 147.5,
								"regularMarketDayHigh": 151.0,
								"fiftyTwoWeekLow": 100.0,
								"fiftyTwoWeekHigh": 160.0,
								"regularMarketVolume": 50000000,
								"averageDailyVolume10Day": 45000000,
								"trailingPE": 25.0,
								"trailingAnnualDividendRate": 0.8,
								"trailingAnnualDividendYield": 0.005,
								"marketCap": 2500000000000,
								"currency": "USD"
							}
						]
					}
				}`
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(response)),
					Header:     make(http.Header),
				}, nil
			}

			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("<html></html>")),
				Header:     make(http.Header),
				Request:    req,
			}
			resp.Header.Set("Set-Cookie", "A1=mock_a1_cookie; Path=/; Domain=.yahoo.com")
			return resp, nil
		},
	}

	// We need a profile and market
	profile := &Profile{Tickers: []string{"AAPL"}}
	market := NewMarket() // Initializes cookies/crumb

	quotes := NewQuotes(market, profile)
	quotes.Fetch()

	if len(quotes.stocks) != 1 {
		t.Errorf("Expected 1 stock, got %d", len(quotes.stocks))
	} else {
		stock := quotes.stocks[0]
		if stock.Ticker != "AAPL" {
			t.Errorf("Stock Ticker = %s; want AAPL", stock.Ticker)
		}
		if stock.LastTrade != "150.000" {
			t.Errorf("Stock LastTrade = %s; want 150.000", stock.LastTrade)
		}
	}
}
