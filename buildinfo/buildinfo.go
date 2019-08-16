// Package buildinfo defines variables which are set by the Go linker on build time using ldflags.
package buildinfo

const (
	// Placeholder is the text we use to pre-set our variables.
	Placeholder = "Not set - please use a binary release or use 'make' to build gsctl."

	// VersionPlaceholder is the default we use for the verison number.
	VersionPlaceholder = "Not set - only available in a binary release."
)

var (
	// BuildDate is a string representing when the binary is built.
	BuildDate = Placeholder
	// Commit is the commit SHA hash representing the state of the repository.
	Commit = Placeholder
	// Version is the semantic version number of the build.
	Version = VersionPlaceholder
)
