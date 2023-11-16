package main

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// genericGetRequest takes session id as parameter and calls the target endpoint
// Result is plain string, to facilitate development of new requests and structs.
func genericGetRequest(url string) (*http.Response, error) {
	return genericGetRequestWithContext(context.Background(), url)
}

// genericGetRequestWithContext is the same as genericGetRequest, but accepts context
func genericGetRequestWithContext(ctx context.Context, url string) (*http.Response, error) {
	// resp, err := http.Get(fmt.Sprintf("%s&sid=%s", url, sessionID))
	rctx, cancel := context.WithTimeout(ctx, time.Duration(30)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(rctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return httpRequest(req)
}

// genericPostRequest takes session id as parameter and calls the target endpoint
// Result is plain string, to facilitate development of new requests and structs.
func genericPostRequest(url string, params url.Values) (*http.Response, error) {
	return genericPostRequestWithContext(context.Background(), url, params)
}

// genericPostRequestWithContext is the same as genericPostRequest, but accepts context
func genericPostRequestWithContext(ctx context.Context, url string, params url.Values) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	contentType := "application/x-www-form-urlencoded"
	req.Header.Set("Content-Type", contentType)

	return httpRequest(req)
}

func httpRequest(req *http.Request) (*http.Response, error) {
	// Set timeout
	c := &http.Client{}
	return c.Do(req)
}
