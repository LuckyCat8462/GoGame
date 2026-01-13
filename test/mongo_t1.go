package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

func main() {
	// 连接到 MongoDB
	client := connectToMongoDB()
	defer disconnectFromMongoDB(client)

	// Ping 数据库
	pingMongoDB(client)

}

// 连接到 MongoDB
func connectToMongoDB() *mongo.Client {
	// MongoDB 连接字符串
	// 格式: mongodb://用户名:密码@主机:端口/?连接选项
	mongoURI := "mongodb://localhost:27017"

	// 创建连接选项
	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetConnectTimeout(10 * time.Second).       // 连接超时
		SetMaxPoolSize(100).                       // 最大连接池大小
		SetMinPoolSize(10).                        // 最小连接池大小
		SetMaxConnIdleTime(5 * time.Minute).       // 连接最大空闲时间
		SetServerSelectionTimeout(5 * time.Second) // 服务器选择超时

	// 建立连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("连接失败:", err)
	}

	fmt.Println("成功连接到 MongoDB!")
	return client
}

// Ping MongoDB 服务器
func pingMongoDB(client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 使用 Ping 方法检查连接
	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("Ping 失败:", err)
	}

	fmt.Println("Ping 成功: MongoDB 服务正常!")
}

// 断开连接
func disconnectFromMongoDB(client *mongo.Client) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		log.Fatal("断开连接失败:", err)
	}

	fmt.Println("已断开 MongoDB 连接")
}
