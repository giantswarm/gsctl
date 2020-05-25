package featuresupport

type Provider struct {
	// Name represents the common name of a provider. (e.g. 'aws' or 'azure').
	Name string

	// RequiredVersion represents the required version for a feature to work.
	// The version number can be EQUAL TO or LARGER THAN this required version,
	// in order for a feature to work.
	RequiredVersion string
}
