package dao

import (
	"context"
	"core/repo"
	"fmt"
)

const Prefix = "MSQP"
const AccountIdRedisKey = "AccountId"
const AccountIdBegin = 10000

type RedisDao struct {
	repo *repo.Manager
}

// NextAccountId redis自增uid，给一个前缀
func (d *RedisDao) NextAccountId() (string, error) {
	return d.incr(Prefix + ":" + AccountIdRedisKey)
}

// 自增
func (d *RedisDao) incr(key string) (string, error) {
	todo := context.TODO()
	var exist int64
	var err error
	//判断此key是否存在 不存在则set；存在则自增；0 代表不存在
	if d.repo.Redis.Cli != nil {
		exist, err = d.repo.Redis.Cli.Exists(todo, key).Result()
	} else {
		exist, err = d.repo.Redis.ClusterCli.Exists(todo, key).Result()
	}
	//（不存在时）初始化计数器
	if exist == 0 {
		//不存在，则直接设值
		if d.repo.Redis.Cli != nil {
			err = d.repo.Redis.Cli.Set(todo, key, AccountIdBegin, 0).Err()
		} else { //存在
			err = d.repo.Redis.ClusterCli.Set(todo, key, AccountIdBegin, 0).Err()
		}
		if err != nil {
			return "", err
		}
	}

	//执行自增操作
	//Incr命令：将key的值增加1，并返回增加后的值
	//Redis的INCR是原子操作，适合并发场景
	var id int64
	if d.repo.Redis.Cli != nil {
		id, err = d.repo.Redis.Cli.Incr(todo, key).Result()
	} else {
		id, err = d.repo.Redis.ClusterCli.Incr(todo, key).Result()
	}
	if err != nil {
		return "", err
	}

	//返回结果
	return fmt.Sprintf("%d", id), nil
}

func NewRedisDao(m *repo.Manager) *RedisDao {
	return &RedisDao{
		repo: m,
	}
}
