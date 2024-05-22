package cache_srv

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"mapdns/pkg/log"
	"sort"
	"strings"
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

func GetAllUpLevelDomain(domain string, isContainSelf bool) []string {
	var result []string
	parts := strings.Split(domain, ".")
	levelDomain := len(parts) - 1
	for i := levelDomain; i >= 0; i-- {
		if !isContainSelf && i == 0 {
			break
		}
		result = append(result, strings.Join(parts[i:], "."))
	}
	return result
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
	//accuracy
	var exists int64
	var err error
	exists, err = srv.rdb.Exists(srv.ctx, key).Result()
	if err != nil {
		return "", false
	}
	address := ""
	log.Logger().Debugf("getRecord -> exists: [%d]", exists)
	if exists > 0 {
		address, err = srv.rdb.Get(srv.ctx, key).Result()
		if err != nil {
			return "", false
		}
	} else {
		levelDomainArr := GetAllUpLevelDomain(domain, false)
		sort.Slice(levelDomainArr, func(i, j int) bool {
			return len(levelDomainArr[i]) < len(levelDomainArr[j])
		})
		log.Logger().Debugf("getRecord -> levelDomainArr: [%v]", levelDomainArr)
		for _, levelDomain := range levelDomainArr {
			rWildDomain := "*." + levelDomain
			levelDomainCacheKey := GetRedisKey(requestType, rWildDomain)
			log.Logger().Debugf("getRecord -> levelDomainCacheKey: [%s]", levelDomainCacheKey)
			exists, err = srv.rdb.Exists(srv.ctx, levelDomainCacheKey).Result()
			if err != nil {
				continue
			}
			if exists > 0 {
				address, err = srv.rdb.Get(srv.ctx, levelDomainCacheKey).Result()
				if err != nil {
					continue
				}
				log.Logger().Debugf("getRecord -> domain (%s, %d): [%s]", rWildDomain, int(requestType), address)
			}
		}
	}
	log.Logger().Debugf("getRecord -> domain (%s, %d): [%s]", domain, int(requestType), address)
	return address, len(address) > 0
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
