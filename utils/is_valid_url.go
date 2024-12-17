package utils

import (
	"regexp"
)

func IsValidURL(url string) bool {
	regex := `^https?://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=%]+$`
	re := regexp.MustCompile(regex)
	return re.MatchString(url)
}
