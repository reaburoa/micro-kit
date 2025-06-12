package ihttp

import (
	"net/http"
	"time"
)

type IHttpOptions func(client *IHttp)

func WithTimeout(timeout time.Duration) IHttpOptions {
	return func(client *IHttp) {
		if timeout > 0 {
			client.client.Timeout = timeout
		}
	}
}

func WithHttpClient(hClient *http.Client) IHttpOptions {
	return func(client *IHttp) {
		client.client = hClient
	}
}

func WithGzipOn(gzipOn bool) IHttpOptions {
	return func(client *IHttp) {
		client.gzip = gzipOn
	}
}

func WithCookie(jar http.CookieJar) IHttpOptions {
	return func(client *IHttp) {
		client.client.Jar = jar
	}
}

func WithUserAgent(ua string) IHttpOptions {
	return func(client *IHttp) {
		client.userAgent = ua
	}
}

func WithRetries(retry int, retryDelay time.Duration) IHttpOptions {
	return func(client *IHttp) {
		client.retry = retry
		client.retryDelay = retryDelay
	}
}
