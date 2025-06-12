package ihttp

import (
	"context"
	"net/http"

	"github.com/reaburoa/micro-kit/cloud/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type IHttpClientTracing struct {
	transport http.RoundTripper
}

func NewIHttpClientTracing(transport http.RoundTripper) *IHttpClientTracing {
	return &IHttpClientTracing{
		transport: transport,
	}
}

func (t *IHttpClientTracing) RoundTrip(req *http.Request) (*http.Response, error) {
	// beforeProcess: Logic before sending the request.
	_, span := t.start(req.Context(), req)

	// Send the request.
	response, err := t.transport.RoundTrip(req)

	// afterProcess: Logic after receiving the response.
	err = t.afterProcess(span, req, response, err)

	return response, err
}

func (t *IHttpClientTracing) start(ctx context.Context, req *http.Request) (context.Context, trace.Span) {
	operation := "httpRequest"
	if tracer.TraceProvider == nil {
		return ctx, nil
	}
	newCtx, span := tracer.TraceProvider.Start(ctx, operation, trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(attribute.String("Host", req.Host))
	span.SetAttributes(attribute.String("Method", req.Method))
	return newCtx, span
}

func (t *IHttpClientTracing) afterProcess(span trace.Span, req *http.Request, resp *http.Response, respErr error) error {
	if tracer.TraceProvider == nil || span == nil {
		return respErr
	}
	var (
		attrs = make([]attribute.KeyValue, 0)
	)
	defer span.End()
	attrs = append(attrs, attribute.Key("http.url").String(req.URL.String()))

	if resp != nil {
		attrs = append(attrs, attribute.Key("http.status_code").Int(resp.StatusCode))
	}

	switch req.Method {
	case http.MethodGet:
		attrs = append(attrs, attribute.Key("http.req").String(req.URL.RawQuery))
	case http.MethodPost:
		contentType := req.Header.Get("Content-Type")
		if contentType == "application/x-www-form-urlencoded" {
			attrs = append(attrs, attribute.Key("http.req").String(req.PostForm.Encode()))
		}
	}

	span.SetAttributes(attrs...)

	statusCode := codes.Ok
	statusDesc := ""
	if respErr != nil {
		span.RecordError(respErr)
		span.AddEvent("httpRequest error", trace.WithAttributes(attribute.String("error", respErr.Error())))
		statusDesc = respErr.Error()
		statusCode = codes.Error
	}
	span.SetStatus(statusCode, statusDesc)

	return respErr
}
