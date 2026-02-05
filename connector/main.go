package main

import (
	"common/config"
	"common/metrics"
	"connector/app"
	"context"
	"fmt"
	"framework/game"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "connector",
	Short: "connector 管理连接，session以及路由请求",
	Long:  `connector 管理连接，session以及路由请求`,
	Run: func(cmd *cobra.Command, args []string) {
	},
	PostRun: func(cmd *cobra.Command, args []string) {
	},
}

var (
	configFile    string
	gameConfigDir string
	serverId      string
)

func init() {
	rootCmd.Flags().StringVar(&configFile, "config", "application.yml", "app config yml file")
	rootCmd.Flags().StringVar(&gameConfigDir, "gameDir", "../config", "game config dir")
	rootCmd.Flags().StringVar(&serverId, "serverId", "", "app server id， required")
	_ = rootCmd.MarkFlagRequired("serverId")
}

//var configFile = flag.String("config", "application.yml", "config file")

// 连接  写一个 websocket的连接  客户端需要连接这个websocket的两个组件
// 1. wsmanager 2.natsClient
// c := connector.Default()
// c.Run()
func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	config.InitConfig(configFile)
	game.InitConfig(gameConfigDir)
	go func() {
		err := metrics.Serve(fmt.Sprintf("0.0.0.0:#{config.Conf.MetricPort"))
		if err != nil {
			fmt.Println("metric serve err:", err)
		}
	}()

	err := app.Run(context.Background(), "connector001")
	if err != nil {
		log.Print(err)
		os.Exit(-1)
	}
}
