package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Redis 连接配置
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := 0 // 默认使用 0 号数据库

	// 构建 Redis 连接地址
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试连接
	fmt.Println("正在连接 Redis...")
	fmt.Printf("地址: %s\n", redisAddr)
	fmt.Printf("数据库: %d\n", redisDB)

	// 1. Ping 测试
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("❌ Redis 连接失败: %v", err)
	}
	fmt.Printf("✅ Ping 响应: %s\n", pong)

	// 2. 设置测试键值
	testKey := "test_connection"
	testValue := fmt.Sprintf("测试时间: %s", time.Now().Format("2006-01-02 15:04:05"))

	err = rdb.Set(ctx, testKey, testValue, 30*time.Second).Err()
	if err != nil {
		log.Fatalf("❌ 设置键值失败: %v", err)
	}
	fmt.Printf("✅ 设置键值成功: %s -> %s\n", testKey, testValue)

	// 3. 获取测试键值
	val, err := rdb.Get(ctx, testKey).Result()
	if err != nil {
		log.Fatalf("❌ 获取键值失败: %v", err)
	}
	fmt.Printf("✅ 获取键值成功: %s -> %s\n", testKey, val)

	// 4. 获取 TTL
	ttl, err := rdb.TTL(ctx, testKey).Result()
	if err != nil {
		log.Fatalf("❌ 获取 TTL 失败: %v", err)
	}
	fmt.Printf("✅ 键的 TTL: %v\n", ttl)

	// 5. 获取 Redis 信息
	info, err := rdb.Info(ctx).Result()
	if err != nil {
		log.Printf("⚠️  获取 Redis 信息失败: %v", err)
	} else {
		// 显示部分重要信息
		fmt.Println("✅ Redis 服务器信息:")
		fmt.Println(info)
	}

	// 6. 清理测试键
	delResult, err := rdb.Del(ctx, testKey).Result()
	if err != nil {
		log.Printf("⚠️  清理测试键失败: %v", err)
	} else {
		fmt.Printf("✅ 清理测试键成功，删除数量: %d\n", delResult)
	}

	fmt.Println("\n🎉 Redis 连接测试完成！所有操作成功。")

	// 关闭连接
	if err := rdb.Close(); err != nil {
		log.Printf("⚠️  关闭 Redis 连接时出错: %v", err)
	}
}

// 获取环境变量，如果没有则使用默认值
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
