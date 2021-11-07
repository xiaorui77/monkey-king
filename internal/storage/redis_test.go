package storage

import "testing"

var redisAddr = "192.168.17.1:31272"

func TestNewRedisStore(t *testing.T) {
	_, err := NewRedisStore(redisAddr)
	if err != nil {
		t.Errorf("new redis store faile: %v", err)
	}
}

func TestRedisStore_Visit(t *testing.T) {
	c, _ := NewRedisStore(redisAddr)
	c.Visit("https://baidu.com")
}

func TestRedisStore_IsVisited(t *testing.T) {
	c, _ := NewRedisStore(redisAddr)
	_ = c.IsVisited("https://baidu.com")
}
