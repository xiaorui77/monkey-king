package storage

import "testing"

func TestNewRedisStore(t *testing.T) {
	_, err := NewRedisStore("192.168.43.104:6379")
	if err != nil {
		t.Errorf("new redis store faile: %v", err)
	}
}

func TestRedisStore_Visit(t *testing.T) {
	c, _ := NewRedisStore("127.0.0.1:6379")
	c.Visit("https://baidu.com")
}

func TestRedisStore_IsVisited(t *testing.T) {
	c, _ := NewRedisStore("127.0.0.1:6379")
	_ = c.IsVisited("https://baidu.com")

}
