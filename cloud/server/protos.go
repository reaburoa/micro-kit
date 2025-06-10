package server

type Server struct {
	Network   string           `protobuf:"bytes,1,opt,name=network,proto3" json:"network,omitempty"`
	Port      int64            `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"` // 服务监听端口
	Timeout   string           `protobuf:"bytes,3,opt,name=timeout,proto3" json:"timeout,omitempty"`
	Id        string           `protobuf:"bytes,4,opt,name=id,proto3" json:"id,omitempty"`
	Name      string           `protobuf:"bytes,5,opt,name=name,proto3" json:"name,omitempty"`
	Auth      map[string]*Auth `protobuf:"bytes,6,rep,name=auth,proto3" json:"auth,omitempty"`
	GroupName string           `protobuf:"bytes,6,rep,name=group_name,proto3" json:"group_name,omitempty"` // 服务注册groupName，为空则使用默认分组，leo框架默认分组：DEFAULT_GROUP
}

type Auth struct {
	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
}
