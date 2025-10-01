package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func HashPath(path string) string {
	h := md5.Sum([]byte(path))
	return hex.EncodeToString(h[:])
}
