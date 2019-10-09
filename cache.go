package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/coocood/freecache"

	"github.com/go-redis/redis"
)

// type 缓存类型
type _type int

const (
	// Memory 内存缓存
	Memory _type = 1
	// Redis 缓存
	Redis _type = 2
)

// Option 缓存参数
type Option struct {
	Type   _type
	Redis  RedisOption
	Memory MemoryOption
}

// RedisOption redis 参数
type RedisOption struct {
	// 服务器地址
	Host string
	// 服务器端口
	Port int
	// 密码
	Password string
	// 数据库索引
	DB int
}

// MemoryOption memory 参数
type MemoryOption struct {
	// memcache 使用大小
	Size int
}

// Cache 缓存
type Cache interface {
	Set(key string, data interface{}, exp time.Duration) error
	Get(key string) ([]byte, error)
	GetString(key string) (string, error)
	Del(key string) error
}

type redisCache struct {
	redisClient *redis.Client
}

// Set .
func (c *redisCache) Set(key string, data interface{}, exp time.Duration) error {
	if _, err := c.redisClient.Ping().Result(); err != nil {
		return err
	}
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err := c.redisClient.Set(key, bs, exp).Err(); err != nil {
		return err
	}
	return nil
}

// Get .
func (c *redisCache) Get(key string) ([]byte, error) {
	if _, err := c.redisClient.Ping().Result(); err != nil {
		return nil, err
	}
	return c.redisClient.Get(key).Bytes()
}

// GetString .
func (c *redisCache) GetString(key string) (string, error) {
	val, err := c.Get(key)
	if err != nil {
		return "", err
	}
	var ret string
	err = json.Unmarshal(val, &ret)
	if err != nil {
		return "", err
	}
	return ret, nil
}

// Del .
func (c *redisCache) Del(key string) error {
	if _, err := c.redisClient.Ping().Result(); err != nil {
		return err
	}
	c.redisClient.Del(key)
	return nil
}

type memoryCache struct {
	memoryClient *freecache.Cache
}

// Set .
func (c *memoryCache) Set(key string, data interface{}, exp time.Duration) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err := c.memoryClient.Set([]byte(key), bs, int(exp.Seconds())); err != nil {
		return err
	}
	return nil
}

// Get .
func (c *memoryCache) Get(key string) ([]byte, error) {
	val, err := c.memoryClient.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	return val, nil
}

// GetString .
func (c *memoryCache) GetString(key string) (string, error) {
	val, err := c.Get(key)
	if err != nil {
		return "", err
	}
	var ret string
	err = json.Unmarshal(val, &ret)
	if err != nil {
		return "", err
	}
	return ret, nil
}

// Del .
func (c *memoryCache) Del(key string) error {
	c.memoryClient.Del([]byte(key))
	return nil
}

// Cacher .
func Cacher(option *Option) Cache {
	if option == nil {
		panic("option is nil")
	}
	if option.Type == Redis {
		return &redisCache{redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%v:%v", option.Redis.Host, option.Redis.Port),
			Password: option.Redis.Password,
			DB:       option.Redis.DB,
		})}
	}
	return &memoryCache{freecache.NewCache(option.Memory.Size)}
}
