package util

import "strings"

// Truncate takes a string and truncates it
func Truncate(str string, length int) string {
	if len(str) < length {
		return str
	}
	return str[:(length-1)] + "…"
}

// CleanKeypairID returns a cleaned version of the KeyPair ID
func CleanKeypairID(str string) string {
	return strings.Replace(str, ":", "", -1)
}
