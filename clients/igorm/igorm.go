package igorm

import (
	"fmt"
	"time"

	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/protos"
	"github.com/reaburoa/micro-kit/utils/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func GormClient(key string) (*gorm.DB, func(), error) {
	cfg := new(protos.Mysql)
	err := config.Get(fmt.Sprintf("mysql.%s", key)).Scan(&cfg)
	if err != nil {
		return nil, nil, err
	}
	log.Infof("mysql %s config %s", key, cfg)

	client, shutdown, err := ConnGorm(key, cfg)
	if err != nil {
		return nil, nil, err
	}

	return client, shutdown, nil
}

func ConnGorm(instance string, cfg *protos.Mysql) (*gorm.DB, func(), error) {
	conn, err := gorm.Open(mysql.Open(cfg.Dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Error("初始化mysql连接失败", log.Err(err), log.Any("error msg", err.Error()))
		return nil, nil, fmt.Errorf("init mysql [%s] failed: %w", instance, err)
	}
	db, err := conn.DB()
	if err != nil {
		log.Error("mysql连接DB失败", log.Err(err), log.Any("error msg", err.Error()))
		return nil, nil, fmt.Errorf("init mysql db [%s] failed: %w", instance, err)
	}
	err = db.Ping()
	if err != nil {
		log.Error("mysql连通性检测失败", log.Err(err), log.Any("error msg", err.Error()))
		return nil, nil, fmt.Errorf("ping mysql db [%s] failed: %w", instance, err)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(int(cfg.MaxIdle))
	}
	if cfg.MaxOpen > 0 {
		db.SetMaxOpenConns(int(cfg.MaxOpen))
	}
	if cfg.MaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
		db.SetConnMaxIdleTime(time.Duration(cfg.MaxLifetime) * time.Second)
	}
	if cfg.IsDebug {
		conn = conn.Debug()
	}

	for _, plugin := range []gorm.Plugin{tracingGormPlugin} {
		err = conn.Use(plugin)
		if err != nil {
			return nil, nil, fmt.Errorf("register plugin failed: %w", err)
		}
	}
	log.Infof("Ping mysql DB [%s] Success", instance)
	return conn, func() {}, nil
}

func GormClientWithTransPlugin(key string) (*gorm.DB, Provider, func(), error) {
	db, clearFunc, err := GormClient(key)
	if err != nil {
		return nil, nil, clearFunc, err
	}
	err = db.Use(&GormTxnPlugin{})
	if err != nil {
		return nil, nil, clearFunc, fmt.Errorf("register trans plugin failed: %w", err)
	}
	transProvider := NewTransactionProvider(db)
	return db, transProvider, clearFunc, nil
}
