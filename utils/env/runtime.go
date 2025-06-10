package env

import (
	"os"
	"strings"
)

const (
	RunMode   = "RUN_MODE"   // 运行环境
	RunRegion = "RUN_REGION" // 运行区域
)

type Env string
type Region string

const (
	Debug Env = "debug"
	Test  Env = "test"
	Pre   Env = "pre"
	Prod  Env = "prod"
)

const (
	CN   Region = "cn"
	US01 Region = "us01"
)

var (
	currentEnv    = Debug
	currentRegion = CN
)

func init() {
	runMode := strings.ToLower(os.Getenv(RunMode))
	if runMode == "" {
		runMode = string(Debug)
	}
	currentEnv = Env(runMode)

	region := strings.ToLower(os.Getenv(RunRegion))
	if region == "" {
		region = string(CN)
	}
	currentRegion = Region(region)
}

func GetRuntimeRegion() Region {
	return currentRegion
}

func GetRuntimeEnv() Env {
	return currentEnv
}

func IsRegionCN() bool {
	return GetRuntimeRegion() == CN
}

func IsUs01Pre() bool {
	return GetRuntimeRegion() == US01 && GetRuntimeEnv() == Pre
}

func IsUs01Prod() bool {
	return GetRuntimeRegion() == US01 && GetRuntimeEnv() == Prod
}

func IsRelease() bool {
	return GetRuntimeEnv() == Pre || GetRuntimeEnv() == Prod
}

func IsDebug() bool {
	return GetRuntimeEnv() == Debug
}
