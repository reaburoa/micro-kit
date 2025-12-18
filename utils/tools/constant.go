package tools

// hmacHash
type HMacHash int

const (
	SHA1 HMacHash = iota
	SHA256
	SHA512
)

// AESMode AES加密模式
type AESMode int

const (
	AESModeCBC AESMode = iota // CBC模式
	AESModeCTR                // CTR模式（推荐替代CFB）
	AESModeGCM                // GCM模式（推荐，提供认证加密）
	AESModeECB                // ECB模式（不推荐，仅特殊场景使用）
)

// 常量定义
const (
	AES128KeySize = 16
	AES192KeySize = 24
	AES256KeySize = 32
	AESBlockSize  = 16 // AES块大小固定为16字节
)
