package constants

const (
	PlatformAndroid = 1
	PlatformIOS     = 2

	PlatformAndroidStr = "android"
	PlatformIOSStr     = "ios"
)

var (
	PlatformCodeDescMap = map[int]string{
		PlatformAndroid: PlatformAndroidStr,
		PlatformIOS:     PlatformIOSStr,
	}

	PlatformDescCodeMap = map[string]int{
		PlatformAndroidStr: PlatformAndroid,
		PlatformIOSStr:     PlatformIOS,
	}
)

func CovertPlatformStr(platform int) string {
	if p, ok := PlatformCodeDescMap[platform]; ok {
		return p
	}
	return PlatformAndroidStr
}

func CovertPlatformInt(platform string) int {
	if code, ok := PlatformDescCodeMap[platform]; ok {
		return code
	}
	return PlatformAndroid
}
