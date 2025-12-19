package tools

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func Test_AES(t *testing.T) {
	key := []byte("key123key123key1")

	originStr := "this is a string. 挺好的"
	encBy, iv, err := AESEncrypt([]byte(originStr), key, nil, AESModeGCM)
	fmt.Println("AESEncrypt data ", encBy, err)

	fmt.Println("encrypt", base64.StdEncoding.EncodeToString(encBy))

	oriB, err := AESDecrypt(encBy, key, iv, AESModeGCM)
	fmt.Println("AESDecrypt data ", string(oriB), err)
}

func Test_AESGCM(t *testing.T) {
	key := []byte("key123key123key1")

	ed, err := NewAESEncryptorDecryptor(key)
	if err != nil {
		fmt.Println("init AESEncryptorDecryptor object failed, ", err.Error())
		return
	}

	iv, err := generateRandomBytes(getIVSizeForMode(AESModeGCM))
	if err != nil {
		return
	}
	err = ed.SetIV(iv, AESModeGCM)
	if err != nil {
		return
	}
	_ = ed.SetTagSize(16, AESModeGCM)
	ed.SetAuthData([]byte("auth add"))

	originStr := "this is a string. 挺好的"
	encBy, err := ed.Encrypt([]byte(originStr), AESModeGCM)
	fmt.Println("AESEncrypt data ", encBy, err)

	fmt.Println("encrypt", base64.StdEncoding.EncodeToString(encBy))

	oriB, err := ed.Decrypt(encBy, AESModeGCM)
	fmt.Println("AESDecrypt data ", string(oriB), err)
}
