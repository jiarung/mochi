package utils

import "strings"

// URL Trie support.

// URL related constant
const (
	URLPathSeparator      = "/"
	URLGinArbitraryPrefix = ":"
	URLGinArbitraryRepl   = "::"
)

// URLTokenizer tokenizes given path according to URL parsing policy.
func URLTokenizer(path string) []string {
	if strings.HasPrefix(path, URLPathSeparator) {
		path = path[1:]
	}
	return strings.Split(path, URLPathSeparator)
}

// URLTokenJoiner joins given paths into single path.
func URLTokenJoiner(paths ...string) string {
	return strings.Join(paths, URLPathSeparator)
}

// URLGinParamKeyFormatter handle key according to Gin path rule.
func URLGinParamKeyFormatter(key string) string {
	if strings.HasPrefix(key, URLGinArbitraryPrefix) {
		return URLGinArbitraryRepl
	}
	return key
}
