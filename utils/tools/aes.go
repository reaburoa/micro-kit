package tools

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

type AESEncryptorDecryptor struct {
	Key      []byte
	iv       []byte
	tagSize  int
	authData []byte
	block    cipher.Block
}

func validateKey(key []byte) error {
	keyLen := len(key)
	switch keyLen {
	case AES128KeySize, AES192KeySize, AES256KeySize:
		return nil
	default:
		return fmt.Errorf("invalid key size: %d bytes, must be 16, 24 or 32 bytes", keyLen)
	}
}

func validateIV(iv []byte, mode AESMode) error {
	// ECB模式不需要IV
	if mode == AESModeECB {
		if len(iv) > 0 {
			return errors.New("ECB mode does not use IV")
		}
		return nil
	}

	if len(iv) == 0 {
		return errors.New("IV cannot be empty")
	}

	// 检查IV长度
	expectedSize := getIVSizeForMode(mode)
	if len(iv) != expectedSize {
		return fmt.Errorf("IV must be %d bytes for %v mode, got %d",
			expectedSize, mode, len(iv))
	}

	return nil
}

func getIVSizeForMode(mode AESMode) int {
	switch mode {
	case AESModeCBC, AESModeCTR:
		return AESBlockSize
	case AESModeGCM:
		return 12 // GCM推荐使用12字节nonce
	default:
		return AESBlockSize
	}
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

func pkcs7Unpadding(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	padding := int(data[len(data)-1])
	if padding < 1 || padding > len(data) {
		return nil, errors.New("invalid padding 1")
	}

	// 验证填充字节是否正确
	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return nil, errors.New("invalid padding 2")
		}
	}

	return data[:len(data)-padding], nil
}

func NewAESEncryptorDecryptor(key []byte) (*AESEncryptorDecryptor, error) {
	// 验证密钥
	if err := validateKey(key); err != nil {
		return nil, fmt.Errorf("invalid key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	return &AESEncryptorDecryptor{
		block: block,
		Key:   key,
	}, nil
}

// Encrypt 加密数据
func (e *AESEncryptorDecryptor) Encrypt(plaintext []byte, mod AESMode) ([]byte, error) {
	switch mod {
	case AESModeCBC:
		return e.encryptCBC(plaintext)
	case AESModeCTR:
		return e.encryptCTR(plaintext)
	case AESModeGCM:
		return e.encryptGCM(plaintext)
	case AESModeECB:
		return e.encryptECB(plaintext)
	default:
		return nil, errors.New("unsupported encryption mode")
	}
}

func (e *AESEncryptorDecryptor) Decrypt(ciphertext []byte, mod AESMode) ([]byte, error) {
	switch mod {
	case AESModeCBC:
		return e.decryptCBC(ciphertext)
	case AESModeCTR:
		return e.decryptCTR(ciphertext)
	case AESModeGCM:
		return e.decryptGCM(ciphertext)
	case AESModeECB:
		return e.decryptECB(ciphertext)
	default:
		return nil, errors.New("unsupported encryption mode")
	}
}

func (e *AESEncryptorDecryptor) SetIV(iv []byte, mod AESMode) error {
	err := validateIV(iv, mod)
	if err != nil {
		return err
	}
	e.iv = iv
	return nil
}

func (e *AESEncryptorDecryptor) SetTagSize(tagSize int) {
	e.tagSize = tagSize
}

func (e *AESEncryptorDecryptor) SetAuthData(authData []byte) {
	e.authData = authData
}

// EncryptWithIV 加密并返回IV（用于需要存储IV的场景）
func (e *AESEncryptorDecryptor) EncryptWithIV(plaintext []byte, mod AESMode) (ciphertext, iv []byte, err error) {
	// 生成随机IV
	ivSize := getIVSizeForMode(mod)
	iv, err = generateRandomBytes(ivSize)
	if err != nil {
		return nil, nil, err
	}

	tempEncryptor, err := NewAESEncryptorDecryptor(e.Key)
	if err != nil {
		return nil, nil, err
	}
	err = tempEncryptor.SetIV(iv, mod)
	if err != nil {
		return nil, nil, err
	}

	ciphertext, err = tempEncryptor.Encrypt(plaintext, mod)
	if err != nil {
		return nil, nil, err
	}

	return ciphertext, iv, nil
}

// CBC模式加密
func (e *AESEncryptorDecryptor) encryptCBC(plaintext []byte) ([]byte, error) {
	// CBC需要填充
	paddedText := pkcs7Padding(plaintext, e.block.BlockSize())
	ciphertext := make([]byte, len(paddedText))

	mode := cipher.NewCBCEncrypter(e.block, e.iv)
	mode.CryptBlocks(ciphertext, paddedText)

	return ciphertext, nil
}

// CTR模式加密（替代CFB的推荐模式）
func (e *AESEncryptorDecryptor) encryptCTR(plaintext []byte) ([]byte, error) {
	ciphertext := make([]byte, len(plaintext))

	// CTR是流式加密，不需要填充
	stream := cipher.NewCTR(e.block, e.iv)
	stream.XORKeyStream(ciphertext, plaintext)

	return ciphertext, nil
}

// GCM模式加密（推荐）
func (e *AESEncryptorDecryptor) encryptGCM(plaintext []byte) ([]byte, error) {
	// 设置默认的TagSize
	tagSize := e.tagSize
	if tagSize == 0 {
		tagSize = 16 // 默认16字节tag
	}

	// 创建GCM
	aesgcm, err := cipher.NewGCMWithTagSize(e.block, tagSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// 加密（包含认证标签）
	ciphertext := aesgcm.Seal(nil, e.iv, plaintext, e.authData)

	return ciphertext, nil
}

// ECB模式加密（不推荐，仅特殊场景使用）
func (e *AESEncryptorDecryptor) encryptECB(plaintext []byte) ([]byte, error) {
	// 需要填充
	paddedText := pkcs7Padding(plaintext, e.block.BlockSize())
	ciphertext := make([]byte, len(paddedText))

	// ECB模式：逐块加密
	blockSize := e.block.BlockSize()
	for i := 0; i < len(paddedText); i += blockSize {
		e.block.Encrypt(ciphertext[i:i+blockSize], paddedText[i:i+blockSize])
	}

	return ciphertext, nil
}

func (e *AESEncryptorDecryptor) decryptCBC(ciphertext []byte) ([]byte, error) {
	dstCiphertext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(e.block, e.iv)
	mode.CryptBlocks(dstCiphertext, ciphertext)

	return pkcs7Unpadding(dstCiphertext)
}

func (e *AESEncryptorDecryptor) decryptCTR(ciphertext []byte) ([]byte, error) {
	dstCiphertext := make([]byte, len(ciphertext))
	blockMode := cipher.NewCTR(e.block, e.iv)
	blockMode.XORKeyStream(dstCiphertext, ciphertext)

	return dstCiphertext, nil
}

func (e *AESEncryptorDecryptor) decryptGCM(ciphertext []byte) ([]byte, error) {
	// 设置默认的TagSize
	tagSize := e.tagSize
	if tagSize == 0 {
		tagSize = 16 // 默认16字节tag
	}

	// 创建GCM
	aesgcm, err := cipher.NewGCMWithTagSize(e.block, tagSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	// 加密（包含认证标签）
	return aesgcm.Open(nil, e.iv, ciphertext, e.authData)
}

func (e *AESEncryptorDecryptor) decryptECB(ciphertext []byte) ([]byte, error) {
	// 解密
	blockSize := e.block.BlockSize()
	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += blockSize {
		e.block.Decrypt(plaintext[i:i+blockSize], ciphertext[i:i+blockSize])
	}

	// 去除填充
	return pkcs7Unpadding(plaintext)
}

// 加密
func AESEncrypt(plaintext, key, iv []byte, mod AESMode) ([]byte, []byte, error) {
	AESED, err := NewAESEncryptorDecryptor(key)
	if err != nil {
		return nil, nil, err
	}

	if len(iv) <= 0 {
		ivSize := getIVSizeForMode(mod)
		iv, err = generateRandomBytes(ivSize)
		if err != nil {
			return nil, nil, err
		}
	}
	err = AESED.SetIV(iv, mod)
	if err != nil {
		return nil, nil, err
	}

	ciphertext, err := AESED.Encrypt(plaintext, mod)
	if err != nil {
		return nil, nil, err
	}
	return ciphertext, iv, nil
}

// 解密
func AESDecrypt(ciphertext, key, iv []byte, mod AESMode) ([]byte, error) {
	AESED, err := NewAESEncryptorDecryptor(key)
	if err != nil {
		return nil, err
	}
	err = AESED.SetIV(iv, mod)
	if err != nil {
		return nil, err
	}

	return AESED.Decrypt(ciphertext, mod)
}
