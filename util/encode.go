package util

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
)

// encode string to base64
func EncodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func EncodeMd5(s string) string {
	m := md5.New()
	m.Write([]byte(s))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}
