package storage

import (
	"fmt"
	"github.com/xiaorui77/goutils/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Storage interface {
	GetDB() *gorm.DB
}

type storage struct {
	db *gorm.DB
}

func (s *storage) GetDB() *gorm.DB {
	return s.db
}

func NewStorage(addr string) Storage {
	dsn := fmt.Sprintf("root:123456@tcp(%s)/monkey-king?charset=utf8mb4&parseTime=True&loc=Local", addr)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logx.Fatalf("connect DB failed: %v", err)
	}
	return &storage{db: db}
}
