package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func HashPath(path string) string {
	h := md5.Sum([]byte(path))
	return hex.EncodeToString(h[:])
}

func Max[T ~int | ~float64](a, b T) T {
	if a > b {
		return a
	}
	return b
}
func Min[T ~int | ~float64](a, b T) T {
	if a < b {
		return a
	}
	return b
}
