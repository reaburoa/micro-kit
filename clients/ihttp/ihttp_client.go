package ihttp

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	// 默认Transport
	defaultTrans = &http.Transport{
		ResponseHeaderTimeout: 90 * time.Second,
		Proxy:                 http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: 90 * time.Second,
			Timeout:   3 * time.Second,
		}).DialContext,
		MaxIdleConns:          20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		MaxIdleConnsPerHost:   5,
		MaxConnsPerHost:       10,
		ExpectContinueTimeout: 500 * time.Millisecond,
	}
)

type IHttp struct {
	url             string
	method          string
	files           map[string]string // 上传文件form表单名以及文件路径
	fileContentType string
	request         *http.Request
	client          *http.Client
	resp            *http.Response
	params          url.Values
	userAgent       string
	retry           int
	retryDelay      time.Duration
	body            io.Reader
	gzip            bool
	timeout         time.Duration
}

func NewIHttp(urlPath, method string, opts ...IHttpOptions) (*IHttp, error) {
	u, er := url.Parse(urlPath)
	if er != nil {
		return nil, er
	}
	hClient := &IHttp{
		url:    urlPath,
		method: method,
		request: &http.Request{
			URL:        u,
			Method:     method,
			Header:     make(http.Header),
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Host:       u.Host,
		},
		files:  make(map[string]string, 5),
		client: &http.Client{},
		resp:   &http.Response{},
		params: make(url.Values, 5),
	}
	if len(opts) > 0 {
		for _, o := range opts {
			o(hClient)
		}
	}
	if hClient.client.Transport == nil {
		hClient.client.Transport = defaultTrans
	}
	hClient.client.Transport = NewIHttpClientTracing(hClient.client.Transport)
	return hClient, nil
}

// Get
// 提供便捷发起Get请求的接口
func Get(ctx context.Context, url string, params url.Values, timeout time.Duration, header map[string]string) (*http.Response, error) {
	client, err := NewIHttp(url, http.MethodGet, WithTimeout(timeout))
	if err != nil {
		return nil, err
	}
	if len(header) > 0 {
		client = client.MultiHeader(header)
	}
	return client.MultiParams(params).Response(ctx)
}

// Post
func Post(ctx context.Context, url string, body url.Values, timeout time.Duration, header map[string]string) (*http.Response, error) {
	client, err := NewIHttp(url, http.MethodPost, WithTimeout(timeout))
	if err != nil {
		return nil, err
	}
	if len(header) > 0 {
		client = client.MultiHeader(header)
	}
	return client.MultiParams(body).Response(ctx)
}

// PostJson
// 方法会自动将自设设置header头 Content-Type: application/json
func PostJson(ctx context.Context, url string, body io.Reader, timeout time.Duration, header map[string]string) (*http.Response, error) {
	client, err := NewIHttp(url, http.MethodPost, WithTimeout(timeout))
	if err != nil {
		return nil, err
	}
	if len(header) > 0 {
		client = client.MultiHeader(header)
	}
	return client.JsonBody(body).Response(ctx)
}

func (h *IHttp) getResponse(ctx context.Context) (*http.Response, error) {
	if h.resp.StatusCode != 0 {
		return h.resp, nil
	}
	resp, err := h.DoRequest(ctx)
	if err != nil {
		return nil, err
	}
	h.resp = resp

	return resp, nil
}

func (h *IHttp) Header(key, value string) *IHttp {
	h.request.Header.Set(key, value)
	return h
}

func (h *IHttp) MultiHeader(header map[string]string) *IHttp {
	if len(header) > 0 {
		for key, value := range header {
			h.request.Header.Set(key, value)
		}
	}

	return h
}

func (h *IHttp) Param(key, value string) *IHttp {
	h.params.Set(key, value)

	return h
}

func (h *IHttp) MultiParams(params url.Values) *IHttp {
	if len(params) > 0 {
		h.params = params
		h.request.Form = params
	}

	return h
}

func (h *IHttp) BodyWithReader(body io.Reader) *IHttp {
	h.request.Body = io.NopCloser(body)
	return h
}

func (h *IHttp) PostFile(formName, filename string) *IHttp {
	h.files[formName] = filename

	return h
}

func (h *IHttp) SetFileContentType(contentType string) *IHttp {
	h.fileContentType = contentType

	return h
}

func (h *IHttp) fileWriter(bodyWriter *multipart.Writer, fieldName, fileName string) (io.Writer, error) {
	m := make(textproto.MIMEHeader)
	m.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			fieldName, // 参数名为file
			filepath.Base(fileName)))
	// 设置文件格式
	if h.fileContentType == "" { // 文件的数据格式默认为数据流
		m.Set("Content-Type", "application/octet-stream")
	} else {
		m.Set("Content-Type", h.fileContentType)
	}
	return bodyWriter.CreatePart(m)
}

func (h *IHttp) buildURL(paramBody string) bool {
	// build GET url with query string
	if h.request.Method == "GET" && len(paramBody) > 0 {
		if strings.Contains(h.url, "?") {
			h.url += "&" + paramBody
		} else {
			h.url = h.url + "?" + paramBody
		}
		return true
	}

	// build POST/PUT/PATCH url and body
	if (h.request.Method == "POST" || h.request.Method == "PUT" || h.request.Method == "PATCH" || h.request.Method == "DELETE") && h.request.Body == nil {
		// with files
		if len(h.files) > 0 {
			pr, pw := io.Pipe()
			bodyWriter := multipart.NewWriter(pw)
			go func() {
				for formName, filename := range h.files {
					fileWriter, err := h.fileWriter(bodyWriter, formName, filename)
					if err != nil {
						log.Println("http_client init file obj failed", err)
						return
					}
					fh, err := os.Open(filename)
					if err != nil {
						log.Println("http_client open local file failed", err)
						return
					}
					// io copy
					_, err = io.Copy(fileWriter, fh)
					_ = fh.Close()
					if err != nil {
						log.Println("http_client upload local file failed", err)
						return
					}
				}
				for k, v := range h.params {
					for _, vv := range v {
						_ = bodyWriter.WriteField(k, vv)
					}
				}
				_ = bodyWriter.Close()
				_ = pw.Close()
			}()
			h.Header("Content-Type", bodyWriter.FormDataContentType())
			h.request.Body = io.NopCloser(pr)
			h.Header("Transfer-Encoding", "chunked")
			return false
		}

		// with params
		if len(paramBody) > 0 {
			h.Header("Content-Type", "application/x-www-form-urlencoded")
			h.Body(paramBody)
		}
	}

	return false
}

func (h *IHttp) Body(data interface{}) *IHttp {
	switch t := data.(type) {
	case string:
		bf := bytes.NewBufferString(t)
		h.request.Body = io.NopCloser(bf)
		h.request.ContentLength = int64(len(t))
	case []byte:
		bf := bytes.NewBuffer(t)
		h.request.Body = io.NopCloser(bf)
		h.request.ContentLength = int64(len(t))
	}
	return h
}

// XMLBody adds request raw body encoding by XML.
func (h *IHttp) XMLBody(obj interface{}) (*IHttp, error) {
	if h.request.Body == nil && obj != nil {
		xmlBody, err := xml.Marshal(obj)
		if err != nil {
			return h, err
		}
		h.request.Body = io.NopCloser(bytes.NewReader(xmlBody))
		h.request.ContentLength = int64(len(xmlBody))
		h.request.Header.Set("Content-Type", "application/xml")
	}

	return h, nil
}

func (h *IHttp) JsonBody(body io.Reader) *IHttp {
	if h.request.Body == nil && body != nil {
		h.request.Body = io.NopCloser(body)
		h.request.Header.Set("Content-Type", "application/json")
	}
	return h
}

func (h *IHttp) DoRequest(ctx context.Context) (resp *http.Response, err error) {
	var paramBody string
	if len(h.params) > 0 {
		paramBody = h.params.Encode()
	}
	h.request.Form = h.params
	h.request.PostForm = h.params
	if h.buildURL(paramBody) {
		u, er := url.Parse(h.url)
		if er != nil {
			return nil, er
		}
		h.request.URL = u
	}

	if h.userAgent != "" && h.request.Header.Get("User-Agent") == "" {
		h.Header("User-Agent", h.userAgent)
	}
	for i := 0; i <= h.retry; i++ {
		resp, err = h.client.Do(h.request.WithContext(ctx))
		if err == nil {
			break
		}
		if h.retry > 0 && h.retryDelay > 0 {
			time.Sleep(h.retryDelay)
			continue
		}
		break
	}
	return resp, err
}

func (h *IHttp) Response(ctx context.Context) (*http.Response, error) {
	return h.getResponse(ctx)
}

// ToFile saves the body data in response to one file.
// it calls Response inner.
func (h *IHttp) ToFile(ctx context.Context, filename string) error {
	resp, err := h.getResponse(ctx)
	if err != nil {
		return err
	}
	if resp.Body == nil {
		return nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	err = pathExistAndMkdir(filename)
	if err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// Check that the file directory exists, there is no automatically created
func pathExistAndMkdir(filename string) (err error) {
	filename = path.Dir(filename)
	_, err = os.Stat(filename)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(filename, os.ModePerm)
		if err == nil {
			return nil
		}
	}
	return err
}
