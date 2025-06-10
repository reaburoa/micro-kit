package iredis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/welltop-cn/common/cloud/config"
	"github.com/welltop-cn/common/protos"
	"github.com/welltop-cn/common/utils/log"
)

func RedisClient(key string) (redis.Cmdable, func(), error) {
	rConfig := protos.Redis{}
	if err := config.Get(fmt.Sprintf("redis.%s", key)).Scan(&rConfig); err != nil {
		return nil, nil, err
	}
	c, _ := json.Marshal(&rConfig)
	log.Infof("redis %s config %s", key, string(c))

	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFunc()

	client, shutdown, err := ConnRedis(ctx, &rConfig, key)
	if err != nil {
		return nil, nil, err
	}
	return client, shutdown, nil
}

func ConnRedis(ctx context.Context, rConfig *protos.Redis, key string) (redis.Cmdable, func(), error) {
	client := redis.NewClient(newRedisConfig(rConfig))

	client.AddHook(newHook(key, rConfig))

	err := client.Ping(ctx).Err()
	if err != nil {
		log.Errorf("Ping Redis Failed With %s", err.Error())
		return nil, nil, err
	}
	log.Infof("redis [%s] connect success", key)
	return client, func() {
		client.Close()
	}, nil
}

func newRedisConfig(c *protos.Redis) *redis.Options {
	readTimeout, _ := time.ParseDuration(c.ReadTimeout)
	//writeTimeout, _ := time.ParseDuration(c.WriteTimeout)
	dialTimeout, _ := time.ParseDuration(c.DialTimeout)
	minIdleConn := c.MinIdleConn
	poolTimeTime, _ := time.ParseDuration(c.PoolTimeout)
	//maxIdleTimeout, _ := time.ParseDuration(c.MaxIdleTimeout)
	maxConnAge, _ := time.ParseDuration(c.MaxConnAge)
	maxRetries := c.MaxRetries

	return &redis.Options{
		Addr:        c.Addr,
		Password:    c.Password,
		ReadTimeout: readTimeout,
		//WriteTimeout:    writeTimeout,
		DialTimeout:  dialTimeout,
		DB:           int(c.Db),
		MinIdleConns: int(minIdleConn),
		//ConnMaxIdleTime: maxIdleTimeout,
		PoolTimeout:     poolTimeTime,
		ConnMaxLifetime: maxConnAge,
		MaxRetries:      int(maxRetries),
		PoolSize:        int(c.PoolSize),
		MaxActiveConns:  int(c.MaxConnections),
		MaxIdleConns:    int(c.MaxIdleConn),
	}
}
