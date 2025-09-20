package async

import (
	"fmt"
	"testing"
	"time"

	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/utils/log"
)

func Test_RunWithRecover(t *testing.T) {
	config.InitConfig()
	log.InitLogger()

	RunWithRecover(func() {
		fmt.Println("go ...")
		panic("async panic")
	})

	time.Sleep(10 * time.Second)
}
