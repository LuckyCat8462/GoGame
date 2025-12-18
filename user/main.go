package main

import (
	"common/config"
	"common/metrics"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"user/app"
)

// 加载配置时一般会提供一个configfile配置文件，通过命令行来加载

var configFile = flag.String("config", "application.yml", "Path to config file")

func main() {
	//	1-加载配置
	flag.Parse()
	//	解析配置
	config.InitConfig(*configFile)
	fmt.Println(config.Conf)
	//	2-启动监控
	go func() {
		err := metrics.Serve(fmt.Sprintf("0.0.0.0:%d", config.Conf.MetricPort))
		if err != nil {
			panic(err)
		}
	}()

	//	3-启动应用程序（grpc的服务端）
	err := app.Run(context.Background())
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
}
