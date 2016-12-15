package apischema

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/juju/errgo"
)

func IsStatus(statusCode int, responseBody string) (bool, error) {
	return isStatus(statusCode, "", responseBody)
}

func IsStatusAndReason(statusCode int, reason, responseBody string) (bool, error) {
	return isStatus(statusCode, reason, responseBody)
}

func IsStatusWithRawBody(statusCode int, resBody *io.ReadCloser) (bool, error) {
	return IsStatusAndReasonWithRawBody(statusCode, "", resBody)
}

func IsStatusAndReasonWithRawBody(statusCode int, reason string, resBody *io.ReadCloser) (bool, error) {
	byteSlice, err := ioutil.ReadAll(*resBody)
	if err != nil {
		return false, errgo.Mask(err, errgo.Any)
	}

	// This is a hack to be able to read from response body twice. Because we
	// need to read the response body to identify the actual status of the
	// response, we consume the stream maybe somebody else would like too. Thus
	// we buffer the response body and write it back to the original response
	// body reference.
	*resBody = ioutil.NopCloser(bytes.NewReader(byteSlice))

	return isStatus(statusCode, reason, string(byteSlice))
}

func IsSuccessResponse(statusCode int) bool {
	return statusCode == http.StatusOK
}

func IsFailureResponse(statusCode int) bool {
	return statusCode == http.StatusInternalServerError
}

type ServerResponse struct {
	StatusCode int         `json:"status_code"`
	StatusText string      `json:"status_text"`
	Data       interface{} `json:"data"`
}

func ParseData(resBody *io.ReadCloser, v interface{}) error {
	byteSlice, err := ioutil.ReadAll(*resBody)
	if err != nil {
		return errgo.Mask(err)
	}

	// This is a hack to be able to read from response body twice. Because we
	// need to read the response body to identify the actual status of the
	// response, we consume the stream maybe somebody else would like too. Thus
	// we buffer the response body and write it back to the original response
	// body reference.
	*resBody = ioutil.NopCloser(bytes.NewReader(byteSlice))

	target := ServerResponse{Data: &v}
	if err := json.Unmarshal(byteSlice, &target); err != nil {
		// In case we receive a response we did not expect and cannot read, we just
		// return an error containing the content of the response.
		return newUnexpectedContentError(string(byteSlice))
	}

	return nil
}

//------------------------------------------------------------------------------
// private

func isStatus(statusCode int, reason, responseBody string) (bool, error) {
	var responsePayload ResponsePayload
	if err := json.Unmarshal([]byte(responseBody), &responsePayload); err != nil {
		// In case we receive a response we did not expect and cannot read, we just
		// return an error containing the content of the response.
		return false, newUnexpectedContentError(responseBody)
	}

	if responsePayload.StatusCode == statusCode {
		if reason == "" {
			// We're not looking for a specific reason, so we're done
			return true, nil
		}
		// Match end of status text to ": " + <reason>
		if strings.HasSuffix(responsePayload.StatusText, ": "+reason) {
			return true, nil
		}
	}

	return false, nil
}

// newUnexpectedContentError creates an error containing the given content as messages, unless that is empty.
// In that case a human readable message is returned.
func newUnexpectedContentError(content string) error {
	if content == "" {
		return errgo.New("Unexpected empty response")
	} else {
		return errgo.New(content)
	}
}

func newReason(text, reason string) string {
	if len(reason) > 0 {
		text = text + ": " + reason
	}

	return text
}

func isJSON(s string) bool {
	return isJSONMap(s) || isJSONSlice(s)
}

func isJSONMap(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func isJSONSlice(s string) bool {
	var js []interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
