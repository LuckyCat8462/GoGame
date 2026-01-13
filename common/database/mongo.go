package database

//启动mongo ：sudo systemctl start mongod
//命令行：mongosh

import (
	"common/config"
	"common/logs"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

type MongoManager struct {
	Cli *mongo.Client
	Db  *mongo.Database
}

func NewMongo() *MongoManager {
	//设置client选项
	clientOptions := options.Client().
		ApplyURI(config.Conf.Database.MongoConf.Url).
		SetConnectTimeout(60 * time.Second).        // 连接超时
		SetMaxPoolSize(100).                        // 最大连接池大小
		SetMinPoolSize(10).                         // 最小连接池大小
		SetMaxConnIdleTime(5 * time.Minute).        // 连接最大空闲时间
		SetServerSelectionTimeout(5 * time.Second). // 服务器选择超时
		SetAuth(options.Credential{
			Username: config.Conf.Database.MongoConf.UserName,
			Password: config.Conf.Database.MongoConf.Password,
		})

	//连接mongodb
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//defer disconnectFromMongoDB1(client)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logs.Fatal("mongo连接失败:%v", err)
	}
	fmt.Println("成功连接到 MongoDB!")
	pingMongoDB1(client)

	m := &MongoManager{
		Cli: client,
	}
	m.Db = m.Cli.Database(config.Conf.Database.MongoConf.Db)
	return m
}

// Ping MongoDB 服务器
func pingMongoDB1(client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 使用 Ping 方法检查连接
	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
		logs.Fatal("mongo Ping 失败:", err)
	}

	fmt.Println("Ping 成功: MongoDB 服务正常!")
}

// 断开连接
func disconnectFromMongoDB1(client *mongo.Client) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		logs.Fatal("断开连接失败:", err)
	}

	fmt.Println("❌ 已断开 MongoDB 连接")
}

func (m *MongoManager) Close() {
	err := m.Cli.Disconnect(context.TODO())
	if err != nil {
		logs.Error("mongo close err:%v", err)
	}
}
