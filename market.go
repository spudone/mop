// Copyright (c) 2013-2026 by Michael Dvorkin and contributors. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package mop

// Market stores current market information displayed in the top three lines of
// the screen. The market data is fetched and parsed from the HTML page above.
type Market struct {
	*MarketData
	errors   string        // Error(s), if any.
	provider StockProvider // Provider to fetch market data.
}

// Returns new initialized Market struct.
func NewMarket(provider StockProvider) *Market {
	market := &Market{
		MarketData: &MarketData{
			IsClosed:  false,
			Dow:       make(map[string]string),
			Nasdaq:    make(map[string]string),
			Sp500:     make(map[string]string),
			Tokyo:     make(map[string]string),
			HongKong:  make(map[string]string),
			London:    make(map[string]string),
			Frankfurt: make(map[string]string),
			Yield:     make(map[string]string),
			Oil:       make(map[string]string),
			Yen:       make(map[string]string),
			Euro:      make(map[string]string),
			Gold:      make(map[string]string),
		},
		errors:   "",
		provider: provider,
	}

	return market
}

// Fetch requests market data from the provider.
// If download or data parsing fails Fetch populates 'market.errors'.
func (market *Market) Fetch() (self *Market) {
	self = market // <-- This ensures we return correct market after recover() from panic().
	
	marketData, err := market.provider.FetchMarket()
	if err != nil {
		market.errors = err.Error()
	} else {
		market.errors = ""
		market.MarketData = marketData
	}

	return market
}

// Ok returns two values: 1) boolean indicating whether the error has occurred,
// and 2) the error text itself.
func (market *Market) Ok() (bool, string) {
	return market.errors == ``, market.errors
}
