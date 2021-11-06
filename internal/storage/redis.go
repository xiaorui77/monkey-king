package storage

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"hash/fnv"
	"strconv"
)

type RedisStore struct {
	addr string

	client *redis.Client
}

func NewRedisStore(addr string) (s *RedisStore, err error) {
	c := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if _, err := c.Ping().Result(); err != nil {
		logrus.Errorf("[store] connect to redis-cluster %v failed: %v", addr, err)
		return nil, fmt.Errorf("connect redis failed")
	}
	logrus.Infof("[store] connect to redis-cluster %v successfully", addr)
	return &RedisStore{
		addr:   addr,
		client: c,
	}, nil
}

func (s *RedisStore) Visit(url string) {
	h := fnv.New64()
	_, _ = h.Write([]byte(url))
	hash := strconv.FormatUint(h.Sum64(), 16)

	if err := s.client.Set(KeyPrefix+hash, "true", 0); err != nil {
		logrus.Warnf("[storage] set visit key url(%v) failed: %v", url, err)
	}
}

func (s *RedisStore) IsVisited(url string) bool {
	h := fnv.New64()
	_, _ = h.Write([]byte(url))
	hash := strconv.FormatUint(h.Sum64(), 16)
	res, err := s.client.Get(hash).Result()
	if err != nil {
		logrus.Warnf("[store] get visit key url(%v) failed: %v", url, err)
		return false
	}
	return res == "true"
}
