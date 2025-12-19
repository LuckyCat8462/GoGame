package database

import (
	"common/config"
	"common/logs"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

// MongoManager 封装可以更有利于我们去增加一些属性，便于根据项目进行操作
type MongoManager struct {
	Cli *mongo.Client
	Db  *mongo.Database
}

func NewMongo() *MongoManager {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI(config.Conf.Database.MongoConf.Url)
	//设置认证，用户名密码
	clientOptions.SetAuth(options.Credential{
		Username: config.Conf.Database.MongoConf.UserName,
		Password: config.Conf.Database.MongoConf.Password,
	})
	//设置最小最大连接池
	clientOptions.SetMinPoolSize(uint64(config.Conf.Database.MongoConf.MinPoolSize))
	clientOptions.SetMaxPoolSize(uint64(config.Conf.Database.MongoConf.MaxPoolSize))
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logs.Fatal("mongo connect err:%v", err)
		return nil
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		logs.Fatal("mongo ping err:%v", err)
		return nil
	}
	m := &MongoManager{
		Cli: client,
	}
	m.Db = m.Cli.Database(config.Conf.Database.MongoConf.Db)
	return m
}

// Close 不用的时候close（良好习惯）
func (m *MongoManager) Close() {
	err := m.Cli.Disconnect(context.TODO())
	if err != nil {
		logs.Error("mongo close err:%v", err)
	}
}
