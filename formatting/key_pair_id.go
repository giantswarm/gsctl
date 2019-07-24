package formatting

import "strings"

// CleanKeypairID returns a cleaned version of the KeyPair ID
func CleanKeypairID(str string) string {
	return strings.Replace(str, ":", "", -1)
}
