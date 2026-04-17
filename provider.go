// Copyright (c) 2013-2026 by Michael Dvorkin and contributors. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package mop

type MarketData struct {
	IsClosed  bool
	Dow       map[string]string
	Nasdaq    map[string]string
	Sp500     map[string]string
	Tokyo     map[string]string
	HongKong  map[string]string
	London    map[string]string
	Frankfurt map[string]string
	Yield     map[string]string
	Oil       map[string]string
	Yen       map[string]string
	Euro      map[string]string
	Gold      map[string]string
}

// StockProvider defines the interface for fetching market and quotes data.
type StockProvider interface {
	FetchMarket() (*MarketData, error)
	FetchQuotes(tickers []string) ([]Stock, error)
}
