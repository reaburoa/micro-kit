package igorm

import (
	"errors"
	"strconv"
	"strings"

	"github.com/welltop-cn/common/cloud/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// gorm callback
const (
	componentGorm      = "gorm"
	gormSpanKey        = "gorm_span"
	callBackBeforeName = "trace:before"
	callBackAfterName  = "trace:after"
)

// before gorm before execute action do something
func before(db *gorm.DB) {
	if tracer.TraceProvider == nil {
		return
	}
	_, span := tracer.TraceProvider.Start(db.Statement.Context, componentGorm, trace.WithSpanKind(trace.SpanKindClient))
	db.InstanceSet(gormSpanKey, span)
}

// after gorm after execute action do something
func after(db *gorm.DB) {
	if tracer.TraceProvider == nil {
		return
	}
	_span, isExist := db.InstanceGet(gormSpanKey)
	if !isExist {
		return
	}
	span, ok := _span.(trace.Span)
	if !ok {
		return
	}
	statusCode := codes.Ok
	statusDesc := ""
	if db.Error != nil && !errors.Is(db.Error, gorm.ErrRecordNotFound) {
		span.RecordError(db.Error)
		span.AddEvent("gorm error", trace.WithAttributes(attribute.String("error", db.Error.Error())))
		statusDesc = db.Error.Error()
		statusCode = codes.Error
	}
	sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	sqlSlice := strings.Split(sql, " ")
	span.SetAttributes(attribute.String("Table", db.Statement.Table))
	span.SetAttributes(attribute.String("Operator", sqlSlice[0]))
	span.SetAttributes(attribute.String("SQL", sql))
	span.SetAttributes(attribute.String("RowsAffected", strconv.Itoa(int(db.Statement.RowsAffected))))
	span.SetName(strings.ToLower(componentGorm + "_" + sqlSlice[0]))
	span.SetStatus(statusCode, statusDesc)
	defer span.End()

}

type tracingPlugin struct{}

var tracingGormPlugin gorm.Plugin = &tracingPlugin{}

func (op *tracingPlugin) Name() string {
	return "tracePlugin"
}

func (op *tracingPlugin) Initialize(db *gorm.DB) (err error) {
	// create
	if err = db.Callback().Create().Before("gorm:before_create").Register(callBackBeforeName, before); err != nil {
		return err
	}
	if err = db.Callback().Create().After("gorm:after_create").Register(callBackAfterName, after); err != nil {
		return err
	}

	// update
	if err = db.Callback().Update().Before("gorm:before_update").Register(callBackBeforeName, before); err != nil {
		return err
	}
	if err = db.Callback().Update().After("gorm:after_update").Register(callBackAfterName, after); err != nil {
		return err
	}

	// query
	if err = db.Callback().Query().Before("gorm:query").Register(callBackBeforeName, before); err != nil {
		return err
	}
	if err = db.Callback().Query().After("gorm:after_query").Register(callBackAfterName, after); err != nil {
		return err
	}

	// delete
	if err = db.Callback().Delete().Before("gorm:before_delete").Register(callBackBeforeName, before); err != nil {
		return err
	}
	if err = db.Callback().Delete().After("gorm:after_delete").Register(callBackAfterName, after); err != nil {
		return err
	}

	// row
	if err = db.Callback().Row().Before("gorm:row").Register(callBackBeforeName, before); err != nil {
		return err
	}
	if err = db.Callback().Row().After("gorm:row").Register(callBackAfterName, after); err != nil {
		return err
	}

	// raw
	if err = db.Callback().Raw().Before("gorm:raw").Register(callBackBeforeName, before); err != nil {
		return err
	}
	if err = db.Callback().Raw().After("gorm:raw").Register(callBackAfterName, after); err != nil {
		return err
	}

	return nil
}
