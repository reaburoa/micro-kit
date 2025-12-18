package tools

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkFirstUpperFast(b *testing.B) {
	str := "hello world this is a test string for benchmarking"
	for i := 0; i < b.N; i++ {
		FirstUpper(str)
	}
}

func Test_Upper(t *testing.T) {
	// 运行简单的性能测试
	longStr := "this is a very long string that we want to test performance with"

	// 优化版本
	start := time.Now()
	for i := 0; i < 1000000; i++ {
		FirstUpper(longStr)
	}
	fmt.Printf("Optimized: %v\n", time.Since(start))

	// 高性能版本
	start = time.Now()
	for i := 0; i < 1000000; i++ {
		FirstUpper(longStr)
	}
	fmt.Printf("Fast: %v\n", time.Since(start))
}
