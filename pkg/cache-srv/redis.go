package cache_srv

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"mapdns/pkg/log"
)

const RedisPrefix = "mapdns"

func (srv *Server) initRedis() {
	srv.rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

func (srv *Server) CloseRedis() error {
	return srv.rdb.Close()
}

func GetRedisKey(requestType uint16, domain string) string {
	return fmt.Sprintf("%s:%d_%s", RedisPrefix, requestType, domain)
}

func (srv *Server) SetRecord(requestType uint16, domain, address string) error {
	key := GetRedisKey(requestType, domain)
	log.Logger().Debugf("setRecord -> Redis key: %s", key)
	err := srv.rdb.Set(srv.ctx, key, address, 0).Err()
	if err != nil {
		return err
	}
	log.Logger().Debugf("setRecord -> domain (%s, %d): [%s]", domain, int(requestType), address)
	return nil
}

func (srv *Server) GetRecord(requestType uint16, domain string) (string, bool) {
	key := GetRedisKey(requestType, domain)
	log.Logger().Debugf("getRecord -> Redis key: %s", key)
	val, err := srv.rdb.Get(srv.ctx, key).Result()
	if err != nil {
		return "", false
	}
	log.Logger().Debugf("getRecord -> domain (%s, %d): [%s]", domain, int(requestType), val)
	return val, true
}

func (srv *Server) ClearRecords() error {
	var keys []string
	var cursor uint64 = 0
	for {
		var err error
		keys, cursor, err = srv.rdb.Scan(srv.ctx, cursor, fmt.Sprintf("%s:*", RedisPrefix), 100).Result()
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			break
		}
		_, err = srv.rdb.Del(srv.ctx, keys...).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func (srv *Server) DeleteRecord(requestType uint16, domain string) error {
	key := GetRedisKey(requestType, domain)
	log.Logger().Debugf("deleteRecord -> Redis key: %s", key)
	err := srv.rdb.Del(srv.ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}
