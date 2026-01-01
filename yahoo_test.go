package mop

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMarketFetch(t *testing.T) {
	// Mock the Yahoo Finance API for market data
	marketHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify expected query parameters (optional but good for robustness)

		// Return sample JSON response
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
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	})

	// Mock for cookie/crumb (simply return success)
	crumbHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("mock_crumb"))
	})

	// Since NewMarket fetches cookies/crumbs internally using hardcoded URLs,
	// we can't easily mock those calls unless we change the URLs in the source code or redirect traffic.
	// HOWEVER, since we are passing a custom httpClient, we can mock the Transport to intercept requests.
	// But `NewMarket` constructs the URL using `marketURL` constant which is hardcoded.
	// And `fetchCrumb` uses `crumbURL` constant.

	// Wait! My previous thought about dependency injection was correct, but I cannot change the constants easily without changing code.
	// But `http.Client` uses `Transport`. I can use a custom Transport to route requests to my test server.

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v7/finance/quote" {
			marketHandler(w, r)
		} else if r.URL.Path == "/v1/test/getcrumb" {
			crumbHandler(w, r)
		} else {
			// Cookie URL or others
			w.Write([]byte("mock_cookie"))
		}
	}))
	defer server.Close()

	// Better approach: Mock the `RoundTripper` of the client to intercept requests and return responses based on the request URL.

	mockClient := &http.Client{
		Transport: &MockTransport{
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
				// Need to handle cookies if fetchCookies relies on Set-Cookie headers.
				// fetchCookies calls https://finance.yahoo.com/
				// It expects Set-Cookie header.

				resp := &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("<html></html>")),
					Header:     make(http.Header),
					Request:    req, // Needed for cookies
				}
				// Simulate A1 cookie
				// getA1Cookie looks for "A1" cookie.
				// cookiejar.Cookies(url) gets cookies.
				// But we are mocking RoundTrip. We don't update jar automatically unless we use real client logic?
				// Wait, the client has a jar. The client calls RoundTrip.
				// The client updates the jar from Set-Cookie in response.

				resp.Header.Set("Set-Cookie", "A1=mock_a1_cookie; Path=/; Domain=.yahoo.com")

				return resp, nil
			},
		},
	}

	market := NewMarket(mockClient)
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
	// Mock client
	mockClient := &http.Client{
		Transport: &MockTransport{
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
		},
	}

	// We need a profile and market
	profile := &Profile{Tickers: []string{"AAPL"}}
	market := NewMarket(mockClient) // Initializes cookies/crumb

	quotes := NewQuotes(market, profile, mockClient)
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

// MockTransport implements http.RoundTripper
type MockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}
