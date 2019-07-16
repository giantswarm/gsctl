package util

// Truncate takes a string and truncates it (if doTruncate is true)
func Truncate(str string, length int, doTruncate bool) string {
	if !doTruncate || len(str) < length {
		return str
	}
	return str[:(length-1)] + "…"
}
