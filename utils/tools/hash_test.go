package tools

import (
	"fmt"
	"testing"
)

func Test_Md5(t *testing.T) {
	md5str := Md5ToString([]byte("abc134"))
	fmt.Println(md5str)
}

func Test_HmacHash(t *testing.T) {
	md5str := HmacToString([]byte("abc134"), []byte("abc123456"), SHA1)
	fmt.Println(md5str)
}
