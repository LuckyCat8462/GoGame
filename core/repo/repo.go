package repo

import "common/database"

// Manager 建立一个数据库操作的统一管理器
type Manager struct {
	Mongo *database.MongoManager
	Redis *database.RedisManager
}

func (m *Manager) Close() {
	if m.Mongo != nil {
		m.Mongo.Close()
	}
	if m.Redis != nil {
		m.Redis.Close()
	}
}

func New() *Manager {
	return &Manager{
		Mongo: database.NewMongo(),
		Redis: database.NewRedis(),
	}
}
