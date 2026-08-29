package ioss

import (
	"context"
	"fmt"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/reaburoa/micro-kit/cloud/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type BucketClient struct {
	bucket     *oss.Bucket
	bucketName string
}

type OssBucket = BucketClient

func (b *BucketClient) Name() string { return b.bucketName }

func (b *BucketClient) PutObject(ctx context.Context, objectKey string, reader io.Reader, options ...oss.Option) error {
	if b == nil || b.bucket == nil {
		return fmt.Errorf("oss bucket is not initialized")
	}
	ctx, span := startOSSSpan(ctx, "oss-put-object", b.bucketName, objectKey)
	defer span.End()
	if err := b.bucket.PutObject(objectKey, reader, options...); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "")
	_ = ctx
	return nil
}

func (b *BucketClient) GetObject(ctx context.Context, objectKey string, options ...oss.Option) (io.ReadCloser, error) {
	if b == nil || b.bucket == nil {
		return nil, fmt.Errorf("oss bucket is not initialized")
	}
	ctx, span := startOSSSpan(ctx, "oss-get-object", b.bucketName, objectKey)
	defer span.End()
	reader, err := b.bucket.GetObject(objectKey, options...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetStatus(codes.Ok, "")
	_ = ctx
	return reader, nil
}

func (b *BucketClient) DeleteObject(ctx context.Context, objectKey string, options ...oss.Option) error {
	if b == nil || b.bucket == nil {
		return fmt.Errorf("oss bucket is not initialized")
	}
	ctx, span := startOSSSpan(ctx, "oss-delete-object", b.bucketName, objectKey)
	defer span.End()
	if err := b.bucket.DeleteObject(objectKey, options...); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "")
	_ = ctx
	return nil
}

func (b *BucketClient) SignURL(ctx context.Context, objectKey string, method oss.HTTPMethod, expiry int64, options ...oss.Option) (string, error) {
	if b == nil || b.bucket == nil {
		return "", fmt.Errorf("oss bucket is not initialized")
	}
	ctx, span := startOSSSpan(ctx, "oss-sign-url", b.bucketName, objectKey)
	defer span.End()
	url, err := b.bucket.SignURL(objectKey, method, expiry, options...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	_ = ctx
	return url, nil
}

func startOSSSpan(ctx context.Context, operation, bucket, objectKey string) (context.Context, trace.Span) {
	if tracer.TraceProvider == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	ctx, span := tracer.TraceProvider.Start(ctx, operation, trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("component", "aliyun-oss"),
		attribute.String("cloud.provider", "aliyun"),
		attribute.String("bucket", bucket),
		attribute.String("object.key", objectKey),
	)
	return ctx, span
}
