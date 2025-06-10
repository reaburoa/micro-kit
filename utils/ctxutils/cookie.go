package ctxutils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"strings"

	"github.com/elliotchance/phpserialize"
	"github.com/gin-gonic/gin"
	"github.com/welltop-cn/common/utils/tools"
)

const (
	base64Table         = "9saI0oy_HGitgNA8Fk3hfRqC4pmBOuc6Kx5T-2zSZ1VvjQ7DwnLeMUlErYWbdJPX"
	cookieValidationKey = "n5W7GVDlKLvauNXBKBfx9j-tfTx3YT50"
	BearerSchema        = "Bearer "
)

// GetTokenFromCookieOrHeader 从header或者cookie获取 token
func GetTokenFromCookieOrHeader(ctx *gin.Context) string {
	token := html.EscapeString(ctx.GetHeader("AUTHORIZATION"))
	token = strings.TrimPrefix(token, BearerSchema)
	if token == "" {
		TH, _ := GetYiiCookie(ctx, "TH")
		TP, _ := GetYiiCookie(ctx, "TP")
		TS, _ := GetYiiCookie(ctx, "TS")
		if TH == "" || TP == "" || TS == "" {
			token, _ = GetYiiCookie(ctx, "T")
		} else {
			token = TH + "." + TP + "." + TS
		}
	}
	return token
}

func GetYiiCookie(c *gin.Context, key string) (string, error) {
	rowCookie, err := c.Cookie(key)
	if err != nil {
		return "", err
	}
	ok := strings.HasSuffix(rowCookie, `}`)
	if ok {
		value, err := decodeCookieUsingOldMethod(rowCookie)
		//fmt.Printf("key:%v使用旧解密,结果是%v\n", key, value)
		return value, err
	}

	value, err := DecodeFromBase64(rowCookie)
	//fmt.Printf("key:%v使用新解密,结果是%v\n", key, value)
	return value, err

}

func decodeCookieUsingOldMethod(sData string) (value string, err error) {
	mac := hmac.New(sha256.New, []byte(""))
	_, _ = mac.Write([]byte(""))
	test := fmt.Sprintf("%x", mac.Sum(nil))
	hashLength := len(test)
	if len(sData) < hashLength {
		return "", errors.New("cookie长度不对")
	}
	hash := sData[0:hashLength]
	pureData := tools.StringToBytes(sData[hashLength:])
	mac2 := hmac.New(sha256.New, []byte(cookieValidationKey))
	_, _ = mac2.Write(pureData)
	if hash != fmt.Sprintf("%x", mac2.Sum(nil)) {
		return "", errors.New("cookie验证失败")
	}
	var data map[interface{}]interface{}
	err = phpserialize.Unmarshal(pureData, &data)
	if err != nil {
		return "", err
	}
	for k, v := range data {
		if k.(int64) == 1 {
			return v.(string), nil
		}
	}
	return "", nil
}

func DecodeFromBase64(data string) (value string, err error) {
	coder := base64.NewEncoding(base64Table)
	result, err := coder.DecodeString(data)
	if err != nil {
		return "", err
	}
	return tools.BytesToString(result), nil
}

func SetTokenToCookie(ctx *gin.Context, token string) {
	tokenSli := strings.Split(token, ".")
	tokenHeader := tokenSli[0]
	tokenPayload := tokenSli[1]
	tokenSignature := tokenSli[2]
	SetDetailCookie(ctx, "TH", tokenHeader)
	SetDetailCookie(ctx, "TP", tokenPayload)
	SetDetailCookie(ctx, "TS", tokenSignature)
	SetDetailCookie(ctx, "T", token)
}

func SetDetailCookie(ctx *gin.Context, name string, value string) {
	hostName := ctx.Request.Host
	domain := hostName[strings.Index(hostName, "."):]
	SetYiiCookie(ctx, name, value, 0, "/", domain, true, true)
}

func SetYiiCookie(c *gin.Context, name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	value = EncodeFromBase64(value)
	c.SetCookie(name, value, maxAge, path, domain, secure, httpOnly)
}

func EncodeFromBase64(data string) string {
	coder := base64.NewEncoding(base64Table)
	return coder.EncodeToString(tools.StringToBytes(data))
}
