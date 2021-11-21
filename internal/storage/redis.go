package storage

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"hash/fnv"
	"strconv"
)

const (
	PersistenceTasksKey = "PersistenceTasks"
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

	if err := s.client.Set(KeyPrefix+hash, "true", 0).Err(); err != nil {
		logrus.Warnf("[storage] set visit key[%s] url[%v] failed: %v", KeyPrefix+hash, url, err)
	}
}

func (s *RedisStore) IsVisited(url string) bool {
	h := fnv.New64()
	_, _ = h.Write([]byte(url))
	hash := strconv.FormatUint(h.Sum64(), 16)
	res, err := s.client.Get(KeyPrefix + hash).Result()
	if err != nil {
		logrus.Warnf("[store] get visit key[%s] url[%v]  failed: %v", KeyPrefix+hash, url, err)
		return false
	}
	return res == "true"
}

// PersistenceTasks 持久化tasks，以host保存
func (s *RedisStore) PersistenceTasks(host string, tasks interface{}) error {
	bytes, err := json.Marshal(tasks)
	if err != nil {
		logrus.Errorf("[store] presistence tasks failed on json.Marshal: %v", err)
		return fmt.Errorf("presistence tasks failed")
	}
	logrus.Debugf("[store] presistence tasks, data=%v", bytes)
	if err := s.client.HSet(KeyPrefix+PersistenceTasksKey, host, bytes).Err(); err != nil {
		logrus.Errorf("[store] presistence tasks failed on save to redis: %v", err)
		return fmt.Errorf("presistence tasks failed")
	}
	return nil
}
