// Package flags defines common flags to be used in several or all commands.
package flags

var (
	// APIEndpoint is the API endpoint URL to use.
	APIEndpoint string

	// AuthToken is the authorization/authentication token to use only for this command.
	AuthToken string

	// Verbose enables chatty output, useful for debugging.
	Verbose bool
)
