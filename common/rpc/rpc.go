package rpc

import (
	"common/config"
	"common/discovery"
	"common/logs"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"user/pb"
)

// UserClient RPC客户端
var (
	UserClient pb.UserServiceClient
)

func Init() {
	//etcd解析器：可以在grpc连接的时候 进行触发，通过提供的addr地址 去etcd中进行查找
	//1、创建并注册etcd解析器
	r := discovery.NewResolver(config.Conf.Etcd)
	resolver.Register(r)
	//2、获取用户服务配置
	userDomain := config.Conf.Domain["user"]
	//3、初始化服务客户端
	initClient(userDomain.Name, userDomain.LoadBalance, &UserClient)
}

// 功能：找服务的地址
// 参数：1、name：服务名称-用于服务发现；2、loadBalance：是否启用负载均衡3、client：空接口类型，用于接收具体类型的grpc客户端
func initClient(name string, loadBalance bool, client interface{}) {
	//使用etcd服务发现机制，格式为etcd:///%s；通过name找到服务地址
	addr := fmt.Sprintf("etcd:///%s", name)
	//配置grpc连接选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials())}
	//如果启用负载均衡，添加轮询策略
	//grpc支持多种负载均衡策略，round_robin（轮循），pick_first等
	if loadBalance {
		opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, "round_robin")))
	}
	//建立grpc连接
	conn, err := grpc.DialContext(context.TODO(), addr, opts...)
	if err != nil {
		logs.Fatal("rpc connect etcd err:%v", err)
	}
	//使用类型断言，确认client的具体类型
	switch c := client.(type) {
	//如果是*pb.UserServiceClient这种类型，则为它new一个操作。通过指针赋值，将创建的客户端赋值给传入的指针参数
	case *pb.UserServiceClient:
		*c = pb.NewUserServiceClient(conn)
	default:
		logs.Fatal("unsupported client type")
	}
}
