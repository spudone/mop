// Copyright (c) 2013-2026 by Michael Dvorkin and contributors. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package mop

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	crumbURL     = "https://query1.finance.yahoo.com/v1/test/getcrumb"
	cookieURL    = "https://finance.yahoo.com/"
	userAgent    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/113.0"
	euConsentURL = "https://consent.yahoo.com/v2/collectConsent?sessionId="
)

// fetchCrumb retrieves a unique "crumb" string from Yahoo Finance, which is
// required as a security parameter for API requests.
func fetchCrumb(cookies string) (string, error) {
	client := http.Client{}
	request, err := http.NewRequest(http.MethodGet, crumbURL, nil)
	if err != nil {
		return "", err
	}

	request.Header = http.Header{
		"Accept":          {"*/*"},
		"Accept-Encoding": {"gzip, deflate, br"},
		"Accept-Language": {"en-US,en;q=0.5"},
		"Connection":      {"keep-alive"},
		"Content-Type":    {"text/plain"},
		"Cookie":          {cookies},
		"Host":            {"query1.finance.yahoo.com"},
		"Sec-Fetch-Dest":  {"empty"},
		"Sec-Fetch-Mode":  {"cors"},
		"Sec-Fetch-Site":  {"same-site"},
		"TE":              {"trailers"},
		"User-Agent":      {userAgent},
	}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// fetchCookies performs the initial handshake with Yahoo Finance to obtain
// the necessary authentication cookies (like the A1 cookie). It handles
// redirected consent flows if necessary.
func fetchCookies() (string, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// Get the session ID from the first request
	request, err := http.NewRequest(http.MethodGet, cookieURL, nil)
	if err != nil {
		return "", err
	}

	request.Header = http.Header{
		"Authority":                 {"finance.yahoo.com"},
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Accept-Language":           {"en-US,en;q=0.9"},
		"Sec-Fetch-Dest":            {"document"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-Site":            {"none"},
		"Sec-Fetch-User":            {"?1"},
		"Upgrade-Insecure-Requests": {"1"},
		"User-Agent":                {userAgent},
	}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	cookies := jar.Cookies(response.Request.URL)
	cookieA1 := getA1Cookie(cookies)
	if cookieA1 != "" {
		return cookieA1, nil
	}

	// first pass failed - try EU shenanigans
	sessionRegex := regexp.MustCompile("sessionId=(?:([A-Za-z0-9_-]*))")
	sessionID := sessionRegex.FindStringSubmatch(response.Request.URL.RawQuery)[1]

	csrfRegex := regexp.MustCompile("gcrumb=(?:([A-Za-z0-9_]*))")
	csrfToken := csrfRegex.FindStringSubmatch(response.Request.Response.Request.URL.RawQuery)[1]

	gucsCookie := jar.Cookies(response.Request.URL)
	gucsCookieString := ""
	for _, cookie := range gucsCookie {
		gucsCookieString += cookie.Name + "=" + cookie.Value + "; "
	}
	gucsCookieString = strings.TrimSuffix(gucsCookieString, "; ")

	if len(gucsCookie) == 0 {
		return "", fmt.Errorf("fetchCookies: no gucsCookie found")
	}

	form := url.Values{}
	form.Add("csrfToken", csrfToken)
	form.Add("sessionId", sessionID)
	form.Add("namespace", "yahoo")
	form.Add("agree", "agree")
	request2, err := http.NewRequest(http.MethodPost, euConsentURL+sessionID, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	contentLength := strconv.FormatInt(int64(len(form.Encode())), 10)

	request2.Header = http.Header{
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Accept-Language":           {"en-US,en;q=0.9"},
		"Connection":                {"keep-alive"},
		"Cookie":                    {gucsCookieString},
		"Content-Length":            {contentLength},
		"Content-Type":              {"application/x-www-form-urlencoded"},
		"DNT":                       {"1"},
		"Host":                      {"consent.yahoo.com"},
		"Origin":                    {"https://consent.yahoo.com"},
		"Referer":                   {euConsentURL + sessionID},
		"Sec-Fetch-Dest":            {"document"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-Site":            {"same-origin"},
		"Sec-Fetch-User":            {"?1"},
		"Upgrade-Insecure-Requests": {"1"},
		"User-Agent":                {userAgent},
	}

	response2, err := client.Do(request2)
	if err != nil {
		return "", err
	}
	defer response2.Body.Close()

	cookies = jar.Cookies(response2.Request.URL)
	cookieA1 = getA1Cookie(cookies)
	if cookieA1 != "" {
		return cookieA1, nil
	} else {
		return "", fmt.Errorf("fetchCookies: failed to obtain A1 cookie")
	}
}

// getA1Cookie is a helper that extracts the "A1" cookie from a slice of
// http.Cookies and formats it as a string.
func getA1Cookie(cookies []*http.Cookie) string {
	for _, cookie := range cookies {
		if cookie.Name == "A1" {
			return cookie.Name + "=" + cookie.Value + "; "
		}
	}
	return ""
}
