package clienterror

// IsMalformedResponseError checks whether the error is
// "Malformed response", which can mean several things.
func IsMalformedResponseError(err error) bool {
	return err.Error() == "Malformed response"
}
