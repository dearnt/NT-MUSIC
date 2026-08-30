package security

import (
	"net/url"
	"strings"
)

func ValidOrigin(origin string, allowedHosts []string) bool {
	origin = strings.TrimSpace(origin)

	if origin == "" {
		return false
	}

	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	host := strings.ToLower(parsed.Host)

	for _, allowed := range allowedHosts {
		allowed = strings.ToLower(strings.TrimSpace(allowed))

		if allowed != "" && host == allowed {
			return true
		}
	}

	return false
}
