package tools

import (
	"strings"
	"unicode"
	"unicode/utf8"
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

// FirstUpper 首字母大写
func FirstUpper(str string) string {
	if str == "" {
		return ""
	}

	// 快速路径：ASCII 字符
	if str[0] < utf8.RuneSelf {
		// ASCII 字符
		if 'a' <= str[0] && str[0] <= 'z' {
			// 创建新字符串，只修改第一个字符
			b := []byte(str)
			b[0] -= 'a' - 'A'
			return string(b)
		}
		return str
	}

	// Unicode 路径
	return firstUpperUnicode(str)
}

// FirstLower 首字母小写
func FirstLower(str string) string {
	if str == "" {
		return ""
	}

	// 快速路径：ASCII 字符
	if str[0] < utf8.RuneSelf {
		if 'A' <= str[0] && str[0] <= 'Z' {
			b := []byte(str)
			b[0] += 'a' - 'A'
			return string(b)
		}
		return str
	}

	// Unicode 路径
	return firstLowerUnicode(str)
}

// firstUpperUnicode 处理 Unicode 字符的首字母大写
func firstUpperUnicode(str string) string {
	r, size := utf8.DecodeRuneInString(str)
	if r == utf8.RuneError {
		return str
	}

	upperRune := unicode.ToUpper(r)

	// 如果大小写转换后 rune 长度不变，可以原地替换
	if utf8.RuneLen(upperRune) == size {
		// 构建新字符串
		b := make([]byte, len(str))
		utf8.EncodeRune(b, upperRune)
		copy(b[size:], str[size:])
		return string(b)
	}

	// 长度变化，需要使用 Builder
	var result strings.Builder
	result.Grow(len(str) + 4) // 预留一些额外空间
	result.WriteRune(upperRune)
	result.WriteString(str[size:])
	return result.String()
}

// firstLowerUnicode 处理 Unicode 字符的首字母小写
func firstLowerUnicode(str string) string {
	r, size := utf8.DecodeRuneInString(str)
	if r == utf8.RuneError {
		return str
	}

	lowerRune := unicode.ToLower(r)

	if utf8.RuneLen(lowerRune) == size {
		b := make([]byte, len(str))
		utf8.EncodeRune(b, lowerRune)
		copy(b[size:], str[size:])
		return string(b)
	}

	var result strings.Builder
	result.Grow(len(str) + 4)
	result.WriteRune(lowerRune)
	result.WriteString(str[size:])
	return result.String()
}
