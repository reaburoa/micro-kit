package iredis

import (
	"context"
	"fmt"
	"testing"

	"github.com/welltop-cn/common/protos"
)

func Test_redis(t *testing.T) {
	client, shutdown, _ := ConnRedis(context.Background(), &protos.Redis{
		Addr:     "free-test.redis.rds.aliyuncs.com:6379",
		Password: "jOub9uga3sun2miK",
		Db:       9,
	}, "account")

	defer shutdown()

	ctx := context.Background()
	res, err := client.HIncrBy(ctx, "foo:123", "foo", 1).Result()
	fmt.Println("HIncrBy", res, err)
}
