package database

import (
	"common/config"
	"common/logs"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

//systemctl start redis-server
//ps aux | grep redis

type RedisManager struct {
	Cli        *redis.Client        //redis客户端-单机
	ClusterCli *redis.ClusterClient //集群
}

// NewRedis redis初始化函数，创建redis连接
func NewRedis() *RedisManager {
	var clusterCli *redis.ClusterClient
	var cli *redis.Client
	//对配置进行判断
	addrs := config.Conf.Database.RedisConf.ClusterAddrs
	logs.Debug("addr测试:", config.Conf.Database.RedisConf)
	if len(addrs) == 0 {
		////若未提供 addrs -非集群 单节点
		cli = redis.NewClient(&redis.Options{
			Addr:         config.Conf.Database.RedisConf.Addr,
			PoolSize:     config.Conf.Database.RedisConf.PoolSize,
			MinIdleConns: config.Conf.Database.RedisConf.MinIdleConns,
			Password:     config.Conf.Database.RedisConf.Password,
			DB:           config.Conf.Database.RedisConf.DB,
		})
	} else { //若提供了addrs
		clusterCli = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        config.Conf.Database.RedisConf.ClusterAddrs,
			PoolSize:     config.Conf.Database.RedisConf.PoolSize,
			MinIdleConns: config.Conf.Database.RedisConf.MinIdleConns,
			Password:     config.Conf.Database.RedisConf.Password,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //30s超时的上下文
	defer cancel()

	// 测试连接
	//fmt.Println("正在连接 Redis...")
	//fmt.Printf("地址: %s\n", redis.Options{Addr: config.Conf.Database.RedisConf.Addr})

	//进行连接测试
	if clusterCli != nil {
		if err := clusterCli.Ping(ctx).Err(); err != nil {
			logs.Fatal("❌ redis cluster connect err:%v", err)
			return nil
		}
		fmt.Println("✅ redis cluster connect success")
	}
	if cli != nil {
		if err := cli.Ping(ctx).Err(); err != nil {
			logs.Fatal("❌ redis connect err:%v", err)
			return nil
		}
		fmt.Println("✅ redis 连接成功")
	}
	//都没有问题，则返回redis manager
	return &RedisManager{
		Cli:        cli,
		ClusterCli: clusterCli,
	}
}

// Close redis的关闭函数：安全关闭，哪个连接存在关闭哪个，记录关闭时的错误日志，避免空指针异常
func (r *RedisManager) Close() {
	if r.ClusterCli != nil { //如果ClusterCli非空，则进行错误判断
		if err := r.ClusterCli.Close(); err != nil {
			logs.Error("redis cluster close err:%v", err)
		}
	}
	if r.Cli != nil {
		if err := r.Cli.Close(); err != nil {
			logs.Error("redis close err:%v", err)
		}
	}
}

// Set 封装set操作，-expire超时
func (r *RedisManager) Set(ctx context.Context, key, value string, expire time.Duration) error {
	if r.ClusterCli != nil {
		return r.ClusterCli.Set(ctx, key, value, expire).Err()
	}
	if r.Cli != nil {
		return r.Cli.Set(ctx, key, value, expire).Err()
	}
	return nil
}
