package httpimproc

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type httpclient interface {
	Do(*http.Request) (*http.Response, error)
}

// URLReader is a Reader for reading external sources
// over HTTP/S
type URLReader struct {
	Timeout   time.Duration
	client    httpclient
	sourceURL *url.URL
}

// NewURLReader creates a reader which fetches a blob from a URL
func NewURLReader(source *url.URL) *URLReader {
	return &URLReader{
		Timeout:   time.Second * 20,
		client:    &http.Client{},
		sourceURL: source,
	}
}

// ReadBlob requests the specified URL and returns the result
// as a byte array
func (u *URLReader) ReadBlob() ([]byte, error) {
	request, err := http.NewRequest("GET", u.sourceURL.String(), nil)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.TODO(), u.Timeout)
	defer cancel()

	resp, err := u.client.Do(request.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unsuccessful request for URL '%s'", u.sourceURL.String())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
