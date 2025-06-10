package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/encoding"
)

const (
	successCode     int32 = 0
	baseContentType       = "application"
)

type BaseHTTPResponse struct {
	ErrorCode *int32      `json:"error_code,omitempty"`
	ErrorMsg  string      `json:"error_msg,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

func Success(v interface{}) *BaseHTTPResponse {
	code := successCode
	return &BaseHTTPResponse{&code, "success", v}
}

func (e *BaseHTTPResponse) Error() string {
	return fmt.Sprintf("HTTPError code: %d message: %s", e.ErrorCode, e.ErrorMsg)
}

func Error(errCode int32, errMsg string) *BaseHTTPResponse {
	return &BaseHTTPResponse{&errCode, errMsg, nil}
}

func CommonResponseFunc(w http.ResponseWriter, request *http.Request, v interface{}) error {
	codec, _ := CodecForRequest(request, "Accept")
	data, err := codec.Marshal(Success(v))
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", ContentType(codec.Name()))
	_, err = w.Write(data)
	return err
}

func CommonErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	// errCode := ierrors.Code(err)
	// errMsg := ierrors.Msg(err)
	var errCode int32 = 0
	errMsg := ""
	codec, _ := CodecForRequest(r, "Accept")
	body, err := codec.Marshal(Error(errCode, errMsg))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", ContentType(codec.Name()))
	if errCode >= 400 && errCode < 500 {
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
