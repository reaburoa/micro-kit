package tools

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func Test_AES(t *testing.T) {
	key := []byte("key123key123key1")
	iv := []byte("key123key123key1")

	originStr := "this is a string"
	encBy, iv, err := AESEncrypt([]byte(originStr), key, nil, AESModeCBC)
	fmt.Println("AESEncrypt data ", encBy, err)

	fmt.Println("encrypt", base64.StdEncoding.EncodeToString(encBy))

	oriB, err := AESDecrypt(encBy, key, iv, AESModeCBC)
	fmt.Println("AESDecrypt data ", string(oriB), err)

}

func Test_generateRandomBytes(t *testing.T) {
	b, err := generateRandomBytes(16)
	fmt.Println(b, err)
}
