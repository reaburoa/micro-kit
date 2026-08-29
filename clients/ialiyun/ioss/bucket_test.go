package ioss

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/reaburoa/micro-kit/cloud/config"
)

func Test_GetObject(t *testing.T) {
	config.InitConfig()

	manger := NewAliyunOSSManger()

	err := manger.RegisterAllClient()
	if err != nil {
		fmt.Println("register oss client err", err)
		return
	}

	bucket, err := manger.GetBucket("us01", "aigc-us02")
	if err != nil {
		fmt.Println("get oss bucket err", err)
		return
	}
	reader, err := bucket.GetObject(context.Background(), "tb/tbl5wK7ykMP9ZNKb/8288/EP1.txt")
	if err != nil {
		fmt.Println("get object from bucket err", err)
		return
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		fmt.Println("read object from bucket err", err)
		return
	}
	fmt.Println(string(data))

	signUrl, err := bucket.SignURL(context.Background(), "tb/tbl5wK7ykMP9ZNKb/8288/EP1.txt", oss.HTTPGet, 7200)
	if err != nil {
		fmt.Println("sign object url from bucket err", err)
		return
	}

	fmt.Println("signUrl==>", signUrl)
}
