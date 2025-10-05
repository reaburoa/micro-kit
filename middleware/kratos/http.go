package kratos

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/reaburoa/micro-kit/errors"
	"github.com/samber/lo"
)

const (
	successCode     int32 = 0
	baseContentType       = "application"
)

var (
	originResponseUrlPath = make([]string, 0, 10)
)

type BaseHTTPResponse struct {
	Code    int32       `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(v interface{}) *BaseHTTPResponse {
	code := successCode
	return &BaseHTTPResponse{code, "success", v}
}

func (e *BaseHTTPResponse) Error() string {
	return fmt.Sprintf("code: %d message: %s", e.Code, e.Message)
}

func Error(errCode int32, errMsg string) *BaseHTTPResponse {
	return &BaseHTTPResponse{errCode, errMsg, nil}
}

func RegisterOriginResponseUrlPath(urlPath []string) {
	originResponseUrlPath = append(originResponseUrlPath, urlPath...)
}

func CommonResponseFunc(w http.ResponseWriter, request *http.Request, v interface{}) error {
	codec, _ := CodecForRequest(request, "Accept")
	data, err := codec.Marshal(Success(v))
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", ContentType(codec.Name()))
	if lo.Contains(originResponseUrlPath, request.URL.Path) {
		var dataRsp = make(map[string]interface{}, 5)
		_ = codec.Unmarshal(data, &dataRsp)
		if d, ok := dataRsp["data"]; ok {
			dataMap, _ := d.(map[string]interface{})
			_, err = w.Write([]byte(dataMap["data"].(string)))
			return err
		}
	}
	_, err = w.Write(data)
	return err
}

func CommonErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	errCode := errors.Code(err)
	errMsg := errors.Message(err)
	codec, _ := CodecForRequest(r, "Accept")
	body, err := codec.Marshal(Error(int32(errCode), errMsg))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", ContentType(codec.Name()))
	if errCode >= 400 && errCode <= 500 {
		w.WriteHeader(int(errCode))
	}
	_, _ = w.Write(body)
}

func CodecForRequest(r *http.Request, name string) (encoding.Codec, bool) {
	for _, accept := range r.Header[name] {
		codec := encoding.GetCodec(ContentSubtype(accept))
		if codec != nil {
			return codec, true
		}
	}
	return encoding.GetCodec("json"), false
}

func ContentType(subtype string) string {
	return strings.Join([]string{baseContentType, subtype}, "/")
}

func ContentSubtype(contentType string) string {
	left := strings.Index(contentType, "/")
	if left == -1 {
		return ""
	}
	right := strings.Index(contentType, ";")
	if right == -1 {
		right = len(contentType)
	}
	if right < left {
		return ""
	}
	return contentType[left+1 : right]
}
