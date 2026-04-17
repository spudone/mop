// Copyright (c) 2013-2026 by Michael Dvorkin and contributors. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package mop

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type YahooProvider struct {
	cookies string
	crumb   string
	errors  string
}

// NewYahooProvider creates a new instance of YahooProvider.
func NewYahooProvider() *YahooProvider {
	return &YahooProvider{}
}

// Initialize ensures that the provider has the necessary cookies and crumb
// required for making authenticated requests to Yahoo Finance.
func (yp *YahooProvider) Initialize() error {
	var err error
	if yp.cookies == "" {
		yp.cookies, err = fetchCookies()
		if err != nil {
			yp.errors = fmt.Sprintf("Error fetching cookies: %v", err)
			return err
		}
	}
	if yp.crumb == "" {
		yp.crumb, err = fetchCrumb(yp.cookies)
		if err != nil {
			yp.errors = fmt.Sprintf("Error fetching crumb: %v", err)
			return err
		}
	}
	return nil
}

// FetchMarket retrieves the broader market indices (Dow, NASDAQ, etc.) and
// commodities (Oil, Gold, etc.) data from Yahoo Finance.
func (yp *YahooProvider) FetchMarket() (*MarketData, error) {
	if err := yp.Initialize(); err != nil {
		return nil, err
	}

	symbols := `^DJI,^IXIC,^GSPC,^N225,^HSI,^FTSE,^GDAXI,^TNX,CL=F,JPY=X,EUR=X,GC=F`
	base := `https://query1.finance.yahoo.com/v7/finance/quote`
	params := `&range=1d&interval=5m&indicators=close&includeTimestamps=false` +
		`&includePrePost=false&corsDomain=finance.yahoo.com&.tsrc=finance`
	url := fmt.Sprintf(`%s?crumb=%s&symbols=%s%s`, base, yp.crumb, symbols, params)

	client := http.Client{}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	request.Header = http.Header{
		"Accept":          {"*/*"},
		"Accept-Language": {"en-US,en;q=0.5"},
		"Connection":      {"keep-alive"},
		"Content-Type":    {"application/json"},
		"Cookie":          {yp.cookies},
		"Host":            {"query1.finance.yahoo.com"},
		"Origin":          {"https://finance.yahoo.com"},
		"Referer":         {"https://finance.yahoo.com"},
		"Sec-Fetch-Dest":  {"empty"},
		"Sec-Fetch-Mode":  {"cors"},
		"Sec-Fetch-Site":  {"same-site"},
		"TE":              {"trailers"},
		"User-Agent":      {userAgent},
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return yp.extractMarket(body)
}

// assignMarket is a helper that extracts price and change data for a single
// market item at a specific position in the API response results.
func assignMarket(results []map[string]interface{}, position int, changeAsPercent bool) map[string]string {
	out := make(map[string]string)
	if len(results) <= position {
		return out
	}
	if val, ok := results[position]["regularMarketChange"].(float64); ok {
		out[`change`] = float2Str(val)
	} else {
		out[`change`] = "N/A"
	}
	if val, ok := results[position]["regularMarketPrice"].(float64); ok {
		out[`latest`] = float2Str(val)
	} else {
		out[`latest`] = "N/A"
	}

	if changeAsPercent {
		if val, ok := results[position]["regularMarketChangePercent"].(float64); ok {
			out[`change`] = float2Str(val) + `%`
		}
	} else {
		if val, ok := results[position]["regularMarketChangePercent"].(float64); ok {
			out[`percent`] = float2Str(val)
		} else {
			out[`percent`] = "N/A"
		}
	}
	return out
}

// extractMarket parses the raw JSON body from the market API and maps it
// into the internal MarketData structure.
func (yp *YahooProvider) extractMarket(body []byte) (*MarketData, error) {
	d := map[string]map[string][]map[string]interface{}{}
	err := json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}
	results, ok := d["quoteResponse"]["result"]
	if !ok || len(results) == 0 {
		return nil, fmt.Errorf("no results found")
	}

	market := &MarketData{}
	market.Dow = assignMarket(results, 0, false)
	market.Nasdaq = assignMarket(results, 1, false)
	market.Sp500 = assignMarket(results, 2, false)
	market.Tokyo = assignMarket(results, 3, false)
	market.HongKong = assignMarket(results, 4, false)
	market.London = assignMarket(results, 5, false)
	market.Frankfurt = assignMarket(results, 6, false)
	market.Yield = make(map[string]string)
	market.Yield[`name`] = `10-year Yield`
	yieldMap := assignMarket(results, 7, false)
	for k, v := range yieldMap {
		market.Yield[k] = v
	}

	market.Oil = assignMarket(results, 8, true)
	market.Yen = assignMarket(results, 9, true)
	market.Euro = assignMarket(results, 10, true)
	market.Gold = assignMarket(results, 11, true)

	return market, nil
}

// chunkTickers splits a large slice of tickers into smaller chunks to avoid
// hitting URL length limits when making API requests.
func chunkTickers(tickers []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(tickers); i += chunkSize {
		end := i + chunkSize
		if end > len(tickers) {
			end = len(tickers)
		}
		chunks = append(chunks, tickers[i:end])
	}
	return chunks
}

// FetchQuotes retrieves detailed stock quote information for the requested
// list of tickers from Yahoo Finance.
func (yp *YahooProvider) FetchQuotes(tickers []string) ([]Stock, error) {
	if len(tickers) == 0 {
		return []Stock{}, nil
	}
	if err := yp.Initialize(); err != nil {
		return nil, err
	}

	chunks := chunkTickers(tickers, 500)
	var allStocks []Stock

	for _, chunk := range chunks {
		symbols := strings.Join(chunk, `,`)
		base := `https://query1.finance.yahoo.com/v7/finance/quote`
		params := `&range=1d&interval=5m&indicators=close&includeTimestamps=false` +
			`&includePrePost=false&corsDomain=finance.yahoo.com&.tsrc=finance`
		url := fmt.Sprintf(`%s?crumb=%s&symbols=%s%s`, base, yp.crumb, symbols, params)

		client := http.Client{}
		request, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		request.Header = http.Header{
			"Accept":          {"*/*"},
			"Accept-Language": {"en-US,en;q=0.5"},
			"Connection":      {"keep-alive"},
			"Content-Type":    {"application/json"},
			"Cookie":          {yp.cookies},
			"Host":            {"query1.finance.yahoo.com"},
			"Origin":          {"https://finance.yahoo.com"},
			"Referer":         {"https://finance.yahoo.com"},
			"Sec-Fetch-Dest":  {"empty"},
			"Sec-Fetch-Mode":  {"cors"},
			"Sec-Fetch-Site":  {"same-site"},
			"TE":              {"trailers"},
			"User-Agent":      {userAgent},
		}

		response, err := client.Do(request)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return nil, err
		}

		stocks, err := yp.parseQuotes(body)
		if err != nil {
			return nil, err
		}
		allStocks = append(allStocks, stocks...)
	}

	return allStocks, nil
}

// parseQuotes unmarshals the raw JSON response from Yahoo and converts each
// result into a Stock struct, including calculating color indicators.
func (yp *YahooProvider) parseQuotes(body []byte) ([]Stock, error) {
	d := map[string]map[string][]map[string]interface{}{}
	err := json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}
	results, ok := d["quoteResponse"]["result"]
	if !ok {
		return nil, fmt.Errorf("no results found")
	}

	stocks := make([]Stock, len(results))
	for i, raw := range results {
		result := map[string]string{}
		for k, v := range raw {
			switch v := v.(type) {
			case string:
				result[k] = v
			case float64:
				result[k] = float2Str(v)
			default:
				result[k] = fmt.Sprintf("%v", v)
			}
		}
		stocks[i].Ticker = result["symbol"]
		stocks[i].LastTrade = result["regularMarketPrice"]
		stocks[i].Change = result["regularMarketChange"]
		stocks[i].ChangePct = result["regularMarketChangePercent"]
		stocks[i].Open = result["regularMarketOpen"]
		stocks[i].Low = result["regularMarketDayLow"]
		stocks[i].High = result["regularMarketDayHigh"]
		stocks[i].Low52 = result["fiftyTwoWeekLow"]
		stocks[i].High52 = result["fiftyTwoWeekHigh"]
		stocks[i].Volume = result["regularMarketVolume"]
		stocks[i].AvgVolume = result["averageDailyVolume10Day"]
		stocks[i].PeRatio = result["trailingPE"]
		stocks[i].PeRatioX = result["trailingPE"]
		stocks[i].Dividend = result["trailingAnnualDividendRate"]

		val, err := strconv.ParseFloat(result["trailingAnnualDividendYield"], 64)
		if err != nil {
			stocks[i].Yield = "N/A"
		} else {
			stocks[i].Yield = strconv.FormatFloat(val*100, 'f', 2, 64)
		}

		stocks[i].MarketCap = result["marketCap"]
		stocks[i].MarketCapX = result["marketCap"]
		stocks[i].Currency = result["currency"]
		stocks[i].PreOpen = result["preMarketChangePercent"]
		stocks[i].AfterHours = result["postMarketChangePercent"]

		adv, err := strconv.ParseFloat(stocks[i].Change, 64)
		stocks[i].Direction = 0
		stocks[i].RowColor = ""
		if err == nil {
			if adv < 0.0 {
				stocks[i].Direction = -1
				stocks[i].RowColor = "loss"
			} else if adv > 0.0 {
				stocks[i].Direction = 1
				stocks[i].RowColor = "gain"
			}
		}

		if pre, err := strconv.ParseFloat(stocks[i].PreOpen, 64); err == nil {
			if pre < 0.0 {
				stocks[i].PreOpenColor = "loss"
			} else if pre > 0.0 {
				stocks[i].PreOpenColor = "gain"
			}
		}

		if aft, err := strconv.ParseFloat(stocks[i].AfterHours, 64); err == nil {
			if aft < 0.0 {
				stocks[i].AfterHoursColor = "loss"
			} else if aft > 0.0 {
				stocks[i].AfterHoursColor = "gain"
			}
		}
	}
	return stocks, nil
}

// float2Str converts a float64 to a human-readable string with units (K, M, B, T)
// for large numbers, keeping 3 decimal places.
func float2Str(v float64) string {
	unit := ""
	switch {
	case v > 1.0e12:
		v /= 1.0e12
		unit = "T"
	case v > 1.0e9:
		v /= 1.0e9
		unit = "B"
	case v > 1.0e6:
		v /= 1.0e6
		unit = "M"
	case v > 1.0e5:
		v /= 1.0e3
		unit = "K"
	default:
		unit = ""
	}
	return fmt.Sprintf("%0.3f%s", v, unit)
}
