package tools

import (
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
