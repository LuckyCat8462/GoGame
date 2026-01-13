package dao

import (
	"context"
	"core/models/entity"
	"core/repo"
)

// AccountDao 结构体定义：AccountDao是数据访问对象，遵循DAO设计模式，其中包含repo.Manager的引用，用于访问数据连接管理
type AccountDao struct {
	repo *repo.Manager
}

// SaveAccount 方法定义：AccountDao 结构体的 SaveAccount 方法
func (d *AccountDao) SaveAccount(ctx context.Context, ac *entity.Account) error {
	//从mongodb中获取名为account的集合，定义为table
	table := d.repo.Mongo.Db.Collection("account")
	// 向集合中插入一条文档（记录），将ac传入；InsertOne：MongoDB插入单个文档的方法；ctx：上下文，用于控制超时、取消等
	_, err := table.InsertOne(ctx, ac)
	if err != nil {
		return err
	}
	return nil
}

// NewAccountDao 工厂函数：目的为创建AccountDao实例，接收数据库管理器repo.manager,返回初始好的AccountDao指针
func NewAccountDao(m *repo.Manager) *AccountDao {
	return &AccountDao{
		repo: m,
	}
}
