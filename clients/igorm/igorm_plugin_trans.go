package igorm

import (
	"context"
	"fmt"
	"runtime"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// 出于对事务处理，将多个表的事务同时处理

// Provider 主要是为了service可以方便的调用事务而做出来的一个结构体
type Provider interface {
	//
	// Transaction
	//  @Description: 开启一个非自动提交的事务, 可以嵌套调用，每次都会从ctx中找是否有合适的事务，如果没有
	//  			  合适的则会新建一个事务，并且注入到ctx，开启一个事务，任意一个fc抛出错误都会进行回滚
	//  @param ctx 上下文，会尝试从ctx中获取事务
	//  @param fc 需要执行的动作，如果返回error，就会直接回滚
	//  @return error
	//
	Transaction(ctx context.Context, fc func(ctx context.Context) error) error
}
type transactionImpl struct {
	mysqlClient *gorm.DB
}

func NewTransactionProvider(mysqlClient *gorm.DB) Provider {
	return &transactionImpl{
		mysqlClient: mysqlClient,
	}
}

func (t *transactionImpl) Transaction(ctx context.Context, fc func(ctx context.Context) error) (err error) {
	txn, ok := t.getTxnFromContext(ctx)
	if ok {
		// 有事务了，继续执行就好了
		return fc(ctx)
	}
	if t.mysqlClient == nil {
		return fmt.Errorf("mysql client is nil")
	}
	txn = t.mysqlClient.Begin()
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			err = errors.Errorf("db transaction panic: %v, stack: \n%s", e, buf)
		}
		if err != nil {
			_ = txn.Rollback().Error
		}
	}()

	newctx := context.WithValue(ctx, getDbTxnID(t.mysqlClient), txn)
	err = fc(newctx)
	if err == nil {
		err = txn.Commit().Error
	}
	return
}

func (t *transactionImpl) getTxnFromContext(ctx context.Context) (*gorm.DB, bool) {
	txAny := ctx.Value(getDbTxnID(t.mysqlClient))
	tx, ok := txAny.(*gorm.DB)
	if !ok {
		return nil, false
	}
	return tx, true
}

// 插件会用到
type dbTxnID string

func getDbTxnID(db *gorm.DB) dbTxnID {
	if db == nil {
		return ""
	}
	return dbTxnID(fmt.Sprintf("%p", db))
}

func getTxnFromContext(ctx context.Context, txnID dbTxnID) (*gorm.DB, bool) {
	if ctx == nil {
		return nil, false
	}
	txAny := ctx.Value(txnID)
	tx, ok := txAny.(*gorm.DB)
	if !ok {
		return nil, false
	}
	return tx, true
}

const pluginName = "gorm:txn-plugin"

type GormTxnPlugin struct {
	*gorm.DB
}

func (txn *GormTxnPlugin) Name() string {
	return pluginName
}

func (txn *GormTxnPlugin) Initialize(db *gorm.DB) error {
	txn.DB = db
	txn.registerCallbacks(db)
	return nil
}

func (txn *GormTxnPlugin) registerCallbacks(db *gorm.DB) {
	txn.DB = db
	txn.Callback().Create().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Update().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Delete().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Raw().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Query().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Row().Before("*").Register(pluginName, txn.beginTxnIfRequired)
}

func (txn *GormTxnPlugin) beginTxnIfRequired(db *gorm.DB) {
	txnDB, ok := getTxnFromContext(db.Statement.Context, getDbTxnID(txn.DB))
	if ok {
		// 有事务走事务
		db.Statement.ConnPool = txnDB.Statement.ConnPool
		return
	}
	//没有事务不操作
	return
}
