package tools

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
)

func Hmac(str, key []byte, sha HMacHash) []byte {
	var hmacHash hash.Hash
	switch sha {
	case SHA1:
		hmacHash = hmac.New(sha1.New, key)
	case SHA256:
		hmacHash = hmac.New(sha256.New, key)
	case SHA512:
		hmacHash = hmac.New(sha512.New, key)
	}
	hmacHash.Write(str)

	return hmacHash.Sum(nil)
}

func HmacToString(str, key []byte, sha HMacHash) string {
	return hex.EncodeToString(Hmac(str, key, sha))
}

func Md5(data []byte) []byte {
	m := md5.New()
	m.Write(data)
	return m.Sum(nil)
}

func Md5ToString(data []byte) string {
	return hex.EncodeToString(Md5(data))
}
