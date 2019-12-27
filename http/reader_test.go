package httpimproc

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockhttp struct {
	lastRequest    *http.Request
	returnError    error
	returnResponse *http.Response
}

func (mh *mockhttp) Do(r *http.Request) (*http.Response, error) {
	mh.lastRequest = r

	if mh.returnError != nil {
		return nil, mh.returnError
	}

	return mh.returnResponse, nil
}

func Test_That_ReadBlob_Issues_GET_Request_To_Specified_URL(t *testing.T) {
	mh := &mockhttp{
		returnResponse: &http.Response{
			Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("test"))),
			StatusCode: 200,
		},
	}
	source, _ := url.Parse("https://www.test.com/path")
	reader := &URLReader{
		client:    mh,
		sourceURL: source,
	}

	_, _ = reader.ReadBlob()

	assert.Equal(t, source.String(), mh.lastRequest.URL.String())
}

func Test_That_ReadBlob_Returns_Error_From_HTTP_Call(t *testing.T) {
	mh := &mockhttp{
		returnError: errors.New("an error message"),
	}
	source, _ := url.Parse("https://www.test.com/path")
	reader := &URLReader{
		client:    mh,
		sourceURL: source,
	}

	_, err := reader.ReadBlob()

	assert.Equal(t, mh.returnError.Error(), err.Error())
}

func Test_That_ReadBlob_Returns_Error_On_NonSuccessful_StatusCode(t *testing.T) {
	mh := &mockhttp{
		returnResponse: &http.Response{
			Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("test"))),
			StatusCode: 400,
		},
	}
	source, _ := url.Parse("https://www.test.com/path")
	reader := &URLReader{
		client:    mh,
		sourceURL: source,
	}

	_, err := reader.ReadBlob()

	assert.Error(t, err)
}

func Test_That_ReadBlob_Returns_Bytes_From_Response_Body(t *testing.T) {
	body := []byte("response body")
	mh := &mockhttp{
		returnResponse: &http.Response{
			Body:       ioutil.NopCloser(bytes.NewBuffer(body)),
			StatusCode: 200,
		},
	}
	source, _ := url.Parse("https://www.test.com/path")
	reader := &URLReader{
		client:    mh,
		sourceURL: source,
	}

	resp, _ := reader.ReadBlob()

	assert.Equal(t, string(body), string(resp))
}
