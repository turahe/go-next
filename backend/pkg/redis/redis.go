package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type RedisService struct {
	Client *redis.Client
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	TTL time.Duration
}

// Default cache TTLs
const (
	DefaultTTL   = 30 * time.Minute
	ShortTTL     = 5 * time.Minute
	LongTTL      = 2 * time.Hour
	SessionTTL   = 24 * time.Hour
	PermanentTTL = 0 // No expiration
)

// Cache keys prefixes
const (
	UserCachePrefix     = "user:"
	PostCachePrefix     = "post:"
	CategoryCachePrefix = "category:"
	CommentCachePrefix  = "comment:"
	MediaCachePrefix    = "media:"
	RoleCachePrefix     = "role:"
	TokenCachePrefix    = "token:"
	SearchCachePrefix   = "search:"
	StatsCachePrefix    = "stats:"
)

func NewRedisService(cfg RedisConfig) *RedisService {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &RedisService{Client: client}
}

// Basic operations
func (r *RedisService) Set(ctx context.Context, key string, value interface{}) error {
	return r.Client.Set(ctx, key, value, 0).Err()
}

func (r *RedisService) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisService) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

func (r *RedisService) Delete(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

func (r *RedisService) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.Client.Exists(ctx, key).Result()
	return result > 0, err
}

// JSON operations
func (r *RedisService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return r.SetWithTTL(ctx, key, jsonData, ttl)
}

func (r *RedisService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// Cache operations with automatic JSON handling
func (r *RedisService) SetCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.SetJSON(ctx, key, value, ttl)
}

func (r *RedisService) GetCache(ctx context.Context, key string, dest interface{}) error {
	return r.GetJSON(ctx, key, dest)
}

// Batch operations
func (r *RedisService) MSet(ctx context.Context, pairs map[string]interface{}) error {
	return r.Client.MSet(ctx, pairs).Err()
}

func (r *RedisService) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return r.Client.MGet(ctx, keys...).Result()
}

func (r *RedisService) MDelete(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Pattern-based operations
func (r *RedisService) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := r.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return r.Client.Del(ctx, keys...).Err()
	}
	return nil
}

func (r *RedisService) GetKeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	return r.Client.Keys(ctx, pattern).Result()
}

// Hash operations
func (r *RedisService) HSet(ctx context.Context, key string, field string, value interface{}) error {
	return r.Client.HSet(ctx, key, field, value).Err()
}

func (r *RedisService) HGet(ctx context.Context, key string, field string) (string, error) {
	return r.Client.HGet(ctx, key, field).Result()
}

func (r *RedisService) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

func (r *RedisService) HDelete(ctx context.Context, key string, fields ...string) error {
	return r.Client.HDel(ctx, key, fields...).Err()
}

// List operations
func (r *RedisService) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.LPush(ctx, key, values...).Err()
}

func (r *RedisService) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.RPush(ctx, key, values...).Err()
}

func (r *RedisService) LPop(ctx context.Context, key string) (string, error) {
	return r.Client.LPop(ctx, key).Result()
}

func (r *RedisService) RPop(ctx context.Context, key string) (string, error) {
	return r.Client.RPop(ctx, key).Result()
}

func (r *RedisService) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.LRange(ctx, key, start, stop).Result()
}

// Set operations
func (r *RedisService) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.SAdd(ctx, key, members...).Err()
}

func (r *RedisService) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.SRem(ctx, key, members...).Err()
}

func (r *RedisService) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.Client.SMembers(ctx, key).Result()
}

func (r *RedisService) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.Client.SIsMember(ctx, key, member).Result()
}

// Sorted Set operations
func (r *RedisService) ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	return r.Client.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
}

func (r *RedisService) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.ZRange(ctx, key, start, stop).Result()
}

func (r *RedisService) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.ZRevRange(ctx, key, start, stop).Result()
}

// Cache invalidation helpers
func (r *RedisService) InvalidateUserCache(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("%s%s*", UserCachePrefix, userID)
	return r.DeletePattern(ctx, pattern)
}

func (r *RedisService) InvalidatePostCache(ctx context.Context, postID string) error {
	pattern := fmt.Sprintf("%s%s*", PostCachePrefix, postID)
	return r.DeletePattern(ctx, pattern)
}

func (r *RedisService) InvalidateCategoryCache(ctx context.Context, categoryID string) error {
	pattern := fmt.Sprintf("%s%s*", CategoryCachePrefix, categoryID)
	return r.DeletePattern(ctx, pattern)
}

func (r *RedisService) InvalidateCommentCache(ctx context.Context, commentID string) error {
	pattern := fmt.Sprintf("%s%s*", CommentCachePrefix, commentID)
	return r.DeletePattern(ctx, pattern)
}

func (r *RedisService) InvalidateMediaCache(ctx context.Context, mediaID string) error {
	pattern := fmt.Sprintf("%s%s*", MediaCachePrefix, mediaID)
	return r.DeletePattern(ctx, pattern)
}

func (r *RedisService) InvalidateRoleCache(ctx context.Context, roleID string) error {
	pattern := fmt.Sprintf("%s%s*", RoleCachePrefix, roleID)
	return r.DeletePattern(ctx, pattern)
}

// Utility functions
func (r *RedisService) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

func (r *RedisService) Close() error {
	return r.Client.Close()
}

func (r *RedisService) Incr(ctx context.Context, key string) error {
	return r.Client.Incr(ctx, key).Err()
}
