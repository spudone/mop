// Copyright (c) 2013-2026 by Michael Dvorkin and contributors. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package mop

const noDataIndicator = `N/A`

// Stock stores quote information for the particular stock ticker. The data
// for all the fields except 'Direction' is fetched using Yahoo market API.
type Stock struct {
	Ticker          string `json:"symbol"`                      // Stock ticker.
	LastTrade       string `json:"regularMarketPrice"`          // l1: last trade.
	Change          string `json:"regularMarketChange"`         // c6: change real time.
	ChangePct       string `json:"regularMarketChangePercent"`  // k2: percent change real time.
	Open            string `json:"regularMarketOpen"`           // o: market open price.
	Low             string `json:"regularMarketDayLow"`         // g: day's low.
	High            string `json:"regularMarketDayHigh"`        // h: day's high.
	Low52           string `json:"fiftyTwoWeekLow"`             // j: 52-weeks low.
	High52          string `json:"fiftyTwoWeekHigh"`            // k: 52-weeks high.
	Volume          string `json:"regularMarketVolume"`         // v: volume.
	AvgVolume       string `json:"averageDailyVolume10Day"`     // a2: average volume.
	PeRatio         string `json:"trailingPE"`                  // r2: P/E ration real time.
	PeRatioX        string `json:"trailingPE"`                  // r: P/E ration (fallback when real time is N/A).
	Dividend        string `json:"trailingAnnualDividendRate"`  // d: dividend.
	Yield           string `json:"trailingAnnualDividendYield"` // y: dividend yield.
	MarketCap       string `json:"marketCap"`                   // j3: market cap real time.
	MarketCapX      string `json:"marketCap"`                   // j1: market cap (fallback when real time is N/A).
	Currency        string `json:"currency"`                    // String code for currency of stock.
	Direction       int    // -1 when change is < $0, 0 when change is = $0, 1 when change is > $0.
	PreOpen         string `json:"preMarketChangePercent,omitempty"`
	AfterHours      string `json:"postMarketChangePercent,omitempty"`
	PreOpenColor    string
	AfterHoursColor string
	RowColor        string
}

// Quotes stores relevant pointers as well as the array of stock quotes for
// the tickers we are tracking.
type Quotes struct {
	market   *Market       // Pointer to Market.
	profile  *Profile      // Pointer to Profile.
	stocks   []Stock       // Array of stock quote data.
	errors   string        // Error string if any.
	provider StockProvider // Provider for quotes.
}

// Sets the initial values and returns new Quotes struct.
func NewQuotes(market *Market, profile *Profile, provider StockProvider) *Quotes {
	return &Quotes{
		market:   market,
		profile:  profile,
		errors:   ``,
		provider: provider,
	}
}

// Fetch the latest stock quotes and parse raw fetched data into array of
// []Stock structs.
func (quotes *Quotes) Fetch() (self *Quotes) {
	self = quotes
	if quotes.isReady() {
		stocks, err := quotes.provider.FetchQuotes(quotes.profile.Tickers)
		if err != nil {
			quotes.errors = err.Error()
		} else {
			quotes.errors = ""
			quotes.stocks = stocks
		}
	}

	return quotes
}

// Ok returns two values: 1) boolean indicating whether the error has occurred,
// and 2) the error text itself.
func (quotes *Quotes) Ok() (bool, string) {
	return quotes.errors == ``, quotes.errors
}

// AddTickers saves the list of tickers and refreshes the stock data if new
// tickers have been added. The function gets called from the line editor
// when user adds new stock tickers.
func (quotes *Quotes) AddTickers(tickers []string) (added int, err error) {
	if added, err = quotes.profile.AddTickers(tickers); err == nil && added > 0 {
		quotes.stocks = nil // Force fetch.
	}
	return
}

// RemoveTickers saves the list of tickers and refreshes the stock data if some
// tickers have been removed. The function gets called from the line editor
// when user removes existing stock tickers.
func (quotes *Quotes) RemoveTickers(tickers []string) (removed int, err error) {
	if removed, err = quotes.profile.RemoveTickers(tickers); err == nil && removed > 0 {
		quotes.stocks = nil // Force fetch.
	}
	return
}

// isReady returns true if we haven't fetched the quotes yet *or* the stock
// market is still open and we might want to grab the latest quotes. In both
// cases we make sure the list of requested tickers is not empty.
func (quotes *Quotes) isReady() bool {
	return (quotes.stocks == nil || !quotes.market.IsClosed) && len(quotes.profile.Tickers) > 0
}
