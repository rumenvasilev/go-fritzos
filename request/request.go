package request

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// genericGetRequest takes session id as parameter and calls the target endpoint
// Result is plain string, to facilitate development of new requests and structs.
func GenericGetRequest(url string) (*http.Response, error) {
	return GenericGetRequestWithContext(context.Background(), url)
}

// genericGetRequestWithContext is the same as genericGetRequest, but accepts context
func GenericGetRequestWithContext(ctx context.Context, url string) (*http.Response, error) {
	// resp, err := http.Get(fmt.Sprintf("%s&sid=%s", url, Session))
	rctx, cancel := context.WithTimeout(ctx, time.Duration(30)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(rctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return HttpRequest(req)
}

// genericPostRequest takes session id as parameter and calls the target endpoint
// Result is plain string, to facilitate development of new requests and structs.
func GenericPostRequest(url string, params url.Values) (*http.Response, error) {
	return GenericPostRequestWithContext(context.Background(), url, params)
}

// genericPostRequestWithContext is the same as genericPostRequest, but accepts context
func GenericPostRequestWithContext(ctx context.Context, url string, params url.Values) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	contentType := "application/x-www-form-urlencoded"
	req.Header.Set("Content-Type", contentType)

	return HttpRequest(req)
}

func HttpRequest(req *http.Request) (*http.Response, error) {
	// Set timeout
	c := &http.Client{}
	return c.Do(req)
}

type ResponseHeader string

const (
	HeaderXML  ResponseHeader = "xml"
	HeaderJSON ResponseHeader = "json"
)

func ValidateHeader(rh ResponseHeader, h http.Header) bool {
	var contentType string
	switch rh {
	case HeaderXML:
		contentType = "text/xml"
	case HeaderJSON:
		contentType = "application/json; charset=utf-8"
	default:
		return false
	}

	log.Println("Debug:", h.Get("Content-Type"))

	return h.Get("Content-Type") == contentType
}
