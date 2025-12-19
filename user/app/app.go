package app

import (
	"common/config"
	"common/discovery"
	"common/logs"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run 启动程序	grpc、http、日志、数据库
func Run(ctx context.Context) error {
	//1.做一个日志库 info error fatal debug
	logs.InitLog(config.Conf.AppName)
	//2. etcd注册中心 grpc服务注册到etcd中 客户端访问的时候 通过etcd获取grpc的地址
	register := discovery.NewRegister()
	//加一个ctx上下文，例如需要设置超时的时候，就需要这个东西
	//	启动grpc服务端
	server := grpc.NewServer()
	//放到一个协程中去go func
	go func() {
		listen, err := net.Listen("tcp", config.Conf.Grpc.Addr)
		if err != nil {
			logs.Fatal("failed to grpc listen: %v", err)
		}
		err = register.Register(config.Conf.Etcd)
		if err != nil {
			logs.Fatal("注册失败 grpc register: %v", err)

		}

		//注册grpc service 需要数据库 mongo redis
		//初始化 数据库管理
		//manager := repo.New()
		//阻塞操作
		err = server.Serve(listen)
		if err != nil {
			log.Fatalf("failed to grpc serve run: %v", err)
		}
	}()

	//	优雅启停，遇到中断信号、推出、终止、挂断
	stop := func() {
		server.Stop()
		register.Close()

		time.Sleep(3 * time.Second)
		fmt.Println("stop app finish")
	}

	chanel := make(chan os.Signal, 1)
	//signal.Notify监听信号：SIGINT中断，SIGTERM终止,SIGQUIT退出,SIGHUP挂断
	signal.Notify(chanel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	for {
		select {
		//监听上下文
		case <-ctx.Done():
			stop()
			//time out
			return nil
		case sig := <-chanel:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				stop()
				log.Println("user app quit")
				return nil
			case syscall.SIGHUP:
				stop()
				log.Println("hang up ! user app quit")
				return nil
			default:
				return nil
			}

		}
	}
}
