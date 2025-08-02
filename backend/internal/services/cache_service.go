package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-next/pkg/redis"
)

var GlobalRedisClient *redis.RedisService

type CacheService interface {
	Get(key string, dest interface{}) error
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
	DeletePattern(pattern string) error
	Exists(key string) bool
	Increment(key string) (int64, error)
	SetNX(key string, value interface{}, expiration time.Duration) (bool, error)
}

type cacheService struct{}

func NewCacheService() CacheService {
	return &cacheService{}
}

func (c *cacheService) Get(key string, dest interface{}) error {
	if GlobalRedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	value, err := GlobalRedisClient.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(value), dest)
}

func (c *cacheService) Set(key string, value interface{}, expiration time.Duration) error {
	if GlobalRedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return GlobalRedisClient.SetWithExpiration(ctx, key, string(data), expiration)
}

func (c *cacheService) Delete(key string) error {
	if GlobalRedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	return GlobalRedisClient.Del(ctx, key)
}

func (c *cacheService) DeletePattern(pattern string) error {
	if GlobalRedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	keys, err := GlobalRedisClient.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return GlobalRedisClient.Del(ctx, keys...)
	}
	return nil
}

func (c *cacheService) Exists(key string) bool {
	if GlobalRedisClient == nil {
		return false
	}

	ctx := context.Background()
	result, err := GlobalRedisClient.Exists(ctx, key)
	return err == nil && result > 0
}

func (c *cacheService) Increment(key string) (int64, error) {
	if GlobalRedisClient == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	return GlobalRedisClient.Incr(ctx, key)
}

func (c *cacheService) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	if GlobalRedisClient == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	return GlobalRedisClient.SetNX(ctx, key, string(data), expiration)
}

// Cache keys for different entities
const (
	CacheKeyUser               = "user:%s"
	CacheKeyPost               = "post:%s"
	CacheKeyPosts              = "posts:page:%d:limit:%d"
	CacheKeyCategory           = "category:%s"
	CacheKeyCategories         = "categories"
	CacheKeyComment            = "comment:%s"
	CacheKeyComments           = "comments:post:%s"
	CacheKeyMedia              = "media:%s"
	CacheKeyUserProfile        = "user_profile:%s"
	CacheKeyToken              = "token:%s"
	CacheKeyTokens             = "tokens:user:%s"
	CacheKeyJWTKey             = "jwt_key:%s"
	CacheKeyJWTKeys            = "jwt_keys"
	CacheKeyVerificationToken  = "verification_token:%s"
	CacheKeyVerificationTokens = "verification_tokens:user:%s:type:%s"
	CacheKeyRefreshToken       = "refresh_token:%s"
	CacheKeyRefreshTokens      = "refresh_tokens:user:%s"
)

// Cache durations
const (
	CacheDurationShort  = 5 * time.Minute
	CacheDurationMedium = 30 * time.Minute
	CacheDurationLong   = 2 * time.Hour
	CacheDurationDay    = 24 * time.Hour
	CacheDurationToken  = 1 * time.Hour
	CacheDurationJWTKey = 24 * time.Hour
)

var CacheSvc CacheService = NewCacheService()
