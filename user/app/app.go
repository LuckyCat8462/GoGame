package app

import (
	"common/config"
	"common/discovery"
	"common/logs"
	"context"
	"core/repo"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user/internal/service"
	"user/pb"
)

// Run 启动程序 启动grpc服务 启用http服务  启用日志 启用数据库
func Run(ctx context.Context) error {
	//1.引入日志库 charm bracelet/log
	logs.InitLog(config.Conf.AppName)
	//2. etcd注册中心 grpc服务注册到etcd中 客户端访问的时候 通过etcd获取grpc的地址
	register := discovery.NewRegister()

	//启动grpc服务端
	server := grpc.NewServer()
	go func() { //因为是一个阻塞操作，所以放到协程中
		lis, err := net.Listen("tcp", config.Conf.Grpc.Addr)
		if err != nil {
			logs.Fatal("app.go:user grpc server listen err:%v", err)
		}
		//注册grpc
		err = register.Register(config.Conf.Etcd)
		if err != nil {
			logs.Fatal("app.go:user grpc server register etcd err:%v", err)

		}
		////初始化 数据库管理
		manager := repo.New()
		pb.RegisterUserServiceServer(server, service.NewAccountService(manager))

		//阻塞操作
		err = server.Serve(lis)
		if err != nil {
			logs.Fatal("app.go:user grpc server run err:%v", err)
		} else {
			logs.Info("✅ service user running")
		}

	}()
	//优雅启停：遇到退出、终止、挂断，有一个优雅的结束
	stop := func() {
		fmt.Println("app.go:user grpc server 优雅启停")
		server.Stop()
		register.Close()
		//manager.Close()
		//other
		time.Sleep(3 * time.Second)
		logs.Info("stop app finish")
	}
	//缓冲的channel 信号量signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGHUP)
	for {
		select {
		case <-ctx.Done():
			stop()
			//time out
			return nil
		case s := <-c:
			switch s {
			case syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT:
				stop()
				logs.Info("user app quit")
				return nil
			case syscall.SIGHUP:
				stop()
				logs.Info("hang up!! user app quit")
				return nil
			default:
				return nil
			}
		}
	}

	return nil
}
