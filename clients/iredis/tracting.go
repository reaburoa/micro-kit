package iredis

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/welltop-cn/common/cloud/tracer"
	"github.com/welltop-cn/common/protos"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	MaxStatLen = 50
)

type hook struct {
	optName string
	cfg     *protos.Redis
}

func newHook(name string, cfg *protos.Redis) hook {
	return hook{
		optName: name,
		cfg:     cfg,
	}
}

func (h hook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h hook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if tracer.TraceProvider == nil {
			return next(ctx, cmd)
		}
		optName := "redis_" + cmd.Name()
		ctx, span := tracer.TraceProvider.Start(ctx, optName, trace.WithSpanKind(trace.SpanKindClient))
		span.SetAttributes(attribute.String("DBType", "redis"))
		span.SetAttributes(attribute.String("Database", strconv.Itoa(int(h.cfg.Db))))
		span.SetAttributes(attribute.String("Addr", h.cfg.Addr))

		redisErr := next(ctx, cmd)

		var stat string
		for _, v := range cmd.Args() {
			if s, ok := v.(string); ok && len(stat) < MaxStatLen {
				stat = stat + " " + s
			}
		}
		span.SetAttributes(attribute.String("Statement", stat))
		statusCode := codes.Ok
		statusDesc := ""
		if err := cmd.Err(); err != nil {
			if err != redis.Nil {
				span.RecordError(err)
				statusCode = codes.Error
				statusDesc = err.Error()
				span.AddEvent("redis error", trace.WithAttributes(attribute.String("error", err.Error())))
			}
		}
		span.SetStatus(statusCode, statusDesc)
		span.End()

		return redisErr
	}
}
func (h hook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		if tracer.TraceProvider == nil {
			return next(ctx, cmds)
		}
		pipeName := make([]string, 0, len(cmds))
		for _, cmd := range cmds {
			pipeName = append(pipeName, cmd.Name())
		}
		optName := "redis_pipeline_" + strings.Join(pipeName, "_")
		ctx, span := tracer.TraceProvider.Start(ctx, optName, trace.WithSpanKind(trace.SpanKindClient))
		span.SetAttributes(attribute.String("DBType", "redis"))
		span.SetAttributes(attribute.String("Database", strconv.Itoa(int(h.cfg.Db))))
		span.SetAttributes(attribute.String("Addr", h.cfg.Addr))

		redisErr := next(ctx, cmds)

		statRet := make([]string, 0, len(cmds))
		for i, cmd := range cmds {
			stat := fmt.Sprintf("redis_pipeline_%d", i)
			for _, v := range cmd.Args() {
				if s, ok := v.(string); ok && len(stat) < MaxStatLen {
					stat = stat + s
				}
			}
			statRet = append(statRet, stat)
		}
		statusCode := codes.Ok
		statusDesc := ""
		span.SetAttributes(attribute.String("Statement", strings.Join(statRet, " ")))
		if err := cmds[0].Err(); err != nil {
			if err != redis.Nil {
				statusCode = codes.Error
				statusDesc = err.Error()
				span.RecordError(err)
				span.AddEvent("redis pipeline error", trace.WithAttributes(attribute.String("error", err.Error())))
			}
		}
		span.SetStatus(statusCode, statusDesc)

		return redisErr
	}
}
