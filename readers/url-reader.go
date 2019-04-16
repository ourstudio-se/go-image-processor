package readers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// URLReaderOptions customizes whitelist, blocking calls,
// etc, for a URLReader
type URLReaderOptions struct {
	HostWhitelist []string
	HostBlacklist []string
	Timeout       time.Duration
}

// NewURLReaderOptions creates a default instance of
// URLReaderOptions, with a default timeout of 10 seconds
func NewURLReaderOptions() *URLReaderOptions {
	return &URLReaderOptions{
		Timeout: time.Second * 10,
	}
}

// URLReaderFactory is a proxy to create URLReaders
type URLReaderFactory struct {
	opts *URLReaderOptions
}

// URLReader is a Reader for reading external sources
// over HTTP/S
type URLReader struct {
	opts      *URLReaderOptions
	sourceURL *url.URL
}

// NewURLReaderFactory instantiates factory with
// provided URLReaderOptions, and returns
// a URLReader instance on demand
func NewURLReaderFactory(opts *URLReaderOptions) *URLReaderFactory {
	return &URLReaderFactory{opts}
}

// NewURLReader creates a reader which fetches a blob from a URL
func (f *URLReaderFactory) NewURLReader(source *url.URL) *URLReader {
	return &URLReader{
		opts:      f.opts,
		sourceURL: source,
	}
}

func contains(haystack []string, needle string) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}

	return false
}

// ValidateURL validates the source URL against the specified
// blacklist and whitelist (if any)
func (u *URLReader) ValidateURL() error {
	if contains(u.opts.HostBlacklist, u.sourceURL.Hostname()) {
		return fmt.Errorf("host '%s' is not trusted", u.sourceURL.Hostname())
	}

	if len(u.opts.HostWhitelist) > 0 && !contains(u.opts.HostWhitelist, u.sourceURL.Hostname()) {
		return fmt.Errorf("host '%s' is not trusted", u.sourceURL.Hostname())
	}

	return nil
}

// ReadBlob requests the specified URL and returns the result
// as a byte array
func (u *URLReader) ReadBlob() ([]byte, error) {
	client := &http.Client{
		Timeout: u.opts.Timeout,
	}
	request, err := http.NewRequest("GET", u.sourceURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(request)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the URL '%s' didn't contain a valid image source, it's either missing or returns a bad response", u.sourceURL.String())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
