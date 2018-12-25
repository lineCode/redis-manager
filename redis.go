package redis_manager

import (
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

type RedisConfig struct {
	ID                 int    `json:"id" yaml:"id"`
	Name               string `json:"name" yaml:"name"`
	Host               string `json:"host" yaml:"host"`
	Port               int    `json:"port" yaml:"port"`
	Password           string `json:"password" yaml:"password"`
	DB                 int    `json:"db" yaml:"db"`
	ConnectTimeout     int    `json:"connect_timeout" yaml:"connect_timeout"`
	ReadTimeout        int    `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout       int    `json:"write_timeout" yaml:"write_timeout"`
	RefreshDuration    int    `json:"refresh_duration" yaml:"refresh_duration"`
	ScanRowNumEachLoop int64  `json:"scan_row_num_each_loop" yaml:"scan_row_num_each_loop"`
	PreLoad            bool   `json:"preload" yaml:"preload"`
}

type RedisBaseCfg struct {
	ID   int    `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
}

func newRedisConn(config RedisConfig) (*redis.Client, error) {
	if config.Port <= 0 {
		config.Port = 6379
	}
	if config.DB < 0 {
		config.DB = 0
	}
	if config.ConnectTimeout <= 0 {
		config.ConnectTimeout = 1
	}
	if config.ReadTimeout <= 0 {
		config.ReadTimeout = 10
	}
	if config.WriteTimeout <= 0 {
		config.WriteTimeout = 10
	}
	conn := redis.NewClient(&redis.Options{
		Addr:         config.Host + ":" + strconv.Itoa(config.Port),
		Password:     config.Password,
		DialTimeout:  time.Second * time.Duration(config.ConnectTimeout),
		ReadTimeout:  time.Second * time.Duration(config.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(config.WriteTimeout),
		DB:           config.DB,
	})
	return conn, conn.Ping().Err()
}
