package tools

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"strings"
	"unsafe"
)

func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func IsUrl(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

// ParseToSlice 将以 separator 分隔的字符串转换为slice,convert 函数可以自己提供，函数根据 convert函数返回结果组成slice返回给调用者
func ParseToSlice[T any](paramStr, separator string, convert func(interface{}) T) []T {
	if paramStr == "" {
		return []T{}
	}
	paramSlice := strings.Split(paramStr, separator)
	result := make([]T, 0, len(paramSlice))
	for _, val := range paramSlice {
		result = append(result, convert(val))
	}
	return result
}

func FirstUpper(str string) string {
	if str == "" {
		return ""
	}
	upperStr := strings.ToUpper(str)
	if len(str) == 1 {
		return upperStr
	}
	return upperStr[:1] + str[1:]
}

func FirstLower(str string) string {
	if str == "" {
		return ""
	}
	lowerStr := strings.ToLower(str)
	if len(str) == 1 {
		return lowerStr
	}
	return lowerStr[:1] + str[1:]
}

func Hmac(str, key, sha string) []byte {
	var hmacHash hash.Hash
	switch strings.ToUpper(sha) {
	case SHA1:
		hmacHash = hmac.New(sha1.New, []byte(key))
	case SHA256:
		hmacHash = hmac.New(sha256.New, []byte(key))
	case SHA512:
		hmacHash = hmac.New(sha512.New, []byte(key))
	}
	hmacHash.Write([]byte(str))

	return hmacHash.Sum(nil)
}

func HmacString(str, key, sha string) string {
	hmacByte := Hmac(str, key, sha)
	return hex.EncodeToString(hmacByte)
}

func Md5(data string) string {
	m := md5.New()
	m.Write([]byte(data))
	sign := m.Sum(nil)
	return hex.EncodeToString(sign)
}
