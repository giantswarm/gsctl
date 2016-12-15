package apischema

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/juju/errgo"
)

type Response struct {
	StatusCode int             `json:"status_code"`
	StatusText string          `json:"status_text"`
	Data       json.RawMessage `json:"data,omitempty"`
}

func NewEmptyResponse(statusCode int, statusText string) *Response {
	return &Response{
		StatusCode: statusCode,
		StatusText: statusText,
	}
}

func NewResponse(statusCode int, statusText string, data interface{}) (*Response, error) {
	rawData, err := json.Marshal(data)
	if err != nil {
		return nil, errgo.Mask(err)
	}

	return &Response{
		StatusCode: statusCode,
		StatusText: statusText,
		Data:       rawData,
	}, nil
}

func ParseResponse(resBody *io.ReadCloser) (*Response, error) {
	byteSlice, err := ioutil.ReadAll(*resBody)
	if err != nil {
		return nil, errgo.Mask(err)
	}

	// This is a hack to be able to read from response body twice. Because we
	// need to read the response body to identify the actual status of the
	// response, we consume the stream maybe somebody else would like too. Thus
	// we buffer the response body and write it back to the original response
	// body reference.
	*resBody = ioutil.NopCloser(bytes.NewReader(byteSlice))

	target := Response{}

	if isJSON(string(byteSlice)) {
		if err := json.Unmarshal(byteSlice, &target); err != nil {
			// In case we receive a response we did not expect and cannot read, we just
			// return an error containing the content of the response.
			return nil, newUnexpectedContentError(string(byteSlice))
		}
	} else {
		// In case we receive a response we did not expect and cannot read, we just
		// return an error containing the content of the response.
		return nil, newUnexpectedContentError(string(byteSlice))
	}

	return &target, nil
}

func FromHTTPResponse(resp *http.Response, err error) (*Response, error) {
	if err != nil {
		return nil, err
	}
	return ParseResponse(&resp.Body)
}

func (resp *Response) EnsureStatusCodes(statusCodes ...int) error {
	if containsInt(statusCodes, resp.StatusCode) {
		return nil
	}
	return NewResponseError(resp)
}

func (resp *Response) UnmarshalData(v interface{}) error {
	if err := json.Unmarshal(resp.Data, v); err != nil {
		// In case we receive a data field we did not expect, we just
		// return an error containing the content of the response.
		return newUnexpectedContentError(string(resp.Data))
	}

	return nil
}

func containsInt(list []int, n int) bool {
	for _, listN := range list {
		if listN == n {
			return true
		}
	}

	return false
}
