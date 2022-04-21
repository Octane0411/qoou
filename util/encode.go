package util

import (
	"encoding/base64"
)

// encode string to base64
func EncodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
