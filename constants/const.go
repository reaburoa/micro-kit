package constants

const (
	PlatformAndroidCode = 1
	PlatformIOSCode     = 2
	PlatformWEBCode     = 3

	PlatformAndroidDesc = "android"
	PlatformIOSDesc     = "ios"
	PlatformWEBDesc     = "web"
)

var (
	PlatformCodeDescMap = map[int]string{
		PlatformAndroidCode: PlatformAndroidDesc,
		PlatformIOSCode:     PlatformIOSDesc,
		PlatformWEBCode:     PlatformWEBDesc,
	}

	PlatformDescCodeMap = map[string]int{
		PlatformAndroidDesc: PlatformAndroidCode,
		PlatformIOSDesc:     PlatformIOSCode,
		PlatformWEBDesc:     PlatformWEBCode,
	}
)

func CovertPlatformDesc(platform int) string {
	if p, ok := PlatformCodeDescMap[platform]; ok {
		return p
	}
	return PlatformAndroidDesc
}

func CovertPlatformCode(platform string) int {
	if code, ok := PlatformDescCodeMap[platform]; ok {
		return code
	}
	return PlatformAndroidCode
}
