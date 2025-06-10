package env

var (
	serviceId string // 服务名称
)

// ServiceName return the microservice name.
func ServiceName() string {
	return serviceId
}

// SetServiceName.
func SetServiceName(appId string) {
	serviceId = appId
}
