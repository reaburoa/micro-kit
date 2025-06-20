package config

import (
	"fmt"
	"testing"

	"github.com/reaburoa/micro-kit/protos"
)

func Test_Get(t *testing.T) {
	InitConfig()

	var mysqlObj protos.Mysql
	err := Get("mysql").Scan(&mysqlObj)
	if err != nil {
		fmt.Println("scan mysql config err", err)
	}
	fmt.Println("mysql config", mysqlObj)
}
