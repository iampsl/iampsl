package chat

import (
	"chatserver/global"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

//RedisMgr 管理多个redis连接池
type RedisMgr struct {
	pools map[string]*redis.Pool
}

var redismgr = RedisMgr{pools: make(map[string]*redis.Pool)}

//Get  从指定池中得到连接
func (mgr *RedisMgr) Get(name string) (redis.Conn, error) {
	p, b := mgr.pools[name]
	if !b {
		return nil, fmt.Errorf("not found redis pool:%s", name)
	}
	return p.Get(), nil
}

func initRedisMgr() {
	for _, v := range global.AppConfig.RedisConfig.RedisSplice {
		_, b := redismgr.pools[v.Name]
		if b {
			log.Fatalf("pool %s have already\n", v.Name)
		}
		pool := &redis.Pool{
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", v.IP+":"+strconv.Itoa(int(v.Port)))
				if err != nil {
					return nil, err
				}
				if _, err := c.Do("AUTH", v.Passwd); err != nil {
					c.Close()
					return nil, err
				}
				if _, err := c.Do("SELECT", v.DB); err != nil {
					c.Close()
					return nil, err
				}
				return c, nil
			},
			MaxActive:   int(v.MaxConnect),
			MaxIdle:     int(v.MaxConnect),
			Wait:        true,
			IdleTimeout: 10 * time.Minute,
		}
		redismgr.pools[v.Name] = pool
	}
}
