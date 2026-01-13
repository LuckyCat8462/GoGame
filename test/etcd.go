// 创建一个简单的测试程序 test_etcd.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func main() {
	// 从环境变量或参数获取 etcd 地址
	endpoints := []string{"localhost:2379"}
	if len(os.Args) > 1 {
		endpoints = os.Args[1:]
	}

	fmt.Printf("检查 etcd 服务: %v\n", endpoints)

	// 尝试连接
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
		// 添加 TLS 配置（如果需要）
		// TLS: &tls.Config{...},
	})

	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer cli.Close()

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 方法1：获取集群信息
	resp, err := cli.MemberList(ctx)
	if err != nil {
		log.Fatalf("获取成员列表失败: %v", err)
	}

	fmt.Printf("✓ 连接成功！集群有 %d 个成员:\n", len(resp.Members))
	for _, m := range resp.Members {
		fmt.Printf("  - ID: %x, Name: %s, URLs: %v\n",
			m.ID, m.Name, m.PeerURLs)
	}

	// 方法2：检查每个端点的状态
	fmt.Println("\n检查各端点状态:")
	for _, ep := range endpoints {
		statusCtx, statusCancel := context.WithTimeout(context.Background(), 3*time.Second)
		status, err := cli.Status(statusCtx, ep)
		statusCancel()

		if err != nil {
			fmt.Printf("  %s: ✗ 不可用 (%v)\n", ep, err)
		} else {
			fmt.Printf("  %s: ✓ 可用 (版本: %s, 数据库大小: %d)\n",
				ep, status.Version, status.DbSize)
		}
	}

	fmt.Println("\netcd 服务运行正常！")
}
