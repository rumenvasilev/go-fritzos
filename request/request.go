package request

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"
)

// genericGetRequest takes session id as parameter and calls the target endpoint
// Result is plain string, to facilitate development of new requests and structs.
func GenericGetRequest(url string) (*http.Response, error) {
	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()
	return GenericGetRequestWithContext(rctx, url)
}

// genericGetRequestWithContext is the same as genericGetRequest, but accepts context
func GenericGetRequestWithContext(ctx context.Context, url string) (*http.Response, error) {
	// resp, err := http.Get(fmt.Sprintf("%s&sid=%s", url, Session))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return HttpRequest(req)
}

// genericPostRequest takes session id as parameter and calls the target endpoint
// Result is plain string, to facilitate development of new requests and structs.
func GenericPostRequest(url string, body io.Reader) (*http.Response, error) {
	return GenericPostRequestWithContext(context.Background(), url, body)
}

// genericPostRequestWithContext is the same as genericPostRequest, but accepts context
func GenericPostRequestWithContext(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
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

type ContentType string

const (
	HeaderXML  ContentType = "text/xml"
	HeaderJSON ContentType = "application/json; charset=utf-8"
)

func ValidateHeader(ct ContentType, h http.Header) bool {
	var contentType string
	switch ct {
	case HeaderXML:
		contentType = string(HeaderXML)
	case HeaderJSON:
		contentType = string(HeaderJSON)
	default:
		return false
	}

	log.Println("Debug:", h.Get("Content-Type"))

	return h.Get("Content-Type") == contentType
}
