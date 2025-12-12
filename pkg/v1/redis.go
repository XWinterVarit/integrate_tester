package v1

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps a go-redis client for test helpers.
type RedisClient struct {
	client *redis.Client
}

// ConnectRedis connects to Redis using go-redis/v9.
func ConnectRedis(addr, password string, db int) *RedisClient {
	RecordAction(fmt.Sprintf("Redis Connect: %s", addr), func() { ConnectRedis(addr, password, db) })
	if IsDryRun() {
		return &RedisClient{}
	}
	Logf(LogTypeRedis, "Connecting to Redis at %s (db=%d)", addr, db)
	c := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	if err := c.Ping(context.Background()).Err(); err != nil {
		Fail("Failed to connect to Redis: %v", err)
	}
	Log(LogTypeRedis, "Connected to Redis", "")
	return &RedisClient{client: c}
}

// Set sets a key with expiration.
func (c *RedisClient) Set(key string, value interface{}, expiration time.Duration) {
	RecordAction(fmt.Sprintf("Redis Set: %s", key), func() { c.Set(key, value, expiration) })
	if IsDryRun() {
		return
	}
	if c.client == nil {
		Fail("RedisClient is not connected")
	}
	Log(LogTypeRedis, fmt.Sprintf("SET %s", key), fmt.Sprintf("value=%v, ttl=%s", value, expiration))
	if err := c.client.Set(context.Background(), key, value, expiration).Err(); err != nil {
		Fail("Failed to set redis key %s: %v", key, err)
	}
}

// Get retrieves a key value.
func (c *RedisClient) Get(key string) string {
	RecordAction(fmt.Sprintf("Redis Get: %s", key), func() { c.Get(key) })
	if IsDryRun() {
		return ""
	}
	if c.client == nil {
		Fail("RedisClient is not connected")
	}
	Logf(LogTypeRedis, "GET %s", key)
	val, err := c.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			Fail("Redis key %s not found", key)
		}
		Fail("Failed to get redis key %s: %v", key, err)
	}
	return val
}

// Del deletes keys.
func (c *RedisClient) Del(keys ...string) {
	RecordAction(fmt.Sprintf("Redis Del: %v", keys), func() { c.Del(keys...) })
	if IsDryRun() {
		return
	}
	if c.client == nil {
		Fail("RedisClient is not connected")
	}
	Log(LogTypeRedis, "DEL keys", fmt.Sprintf("%v", keys))
	if err := c.client.Del(context.Background(), keys...).Err(); err != nil {
		Fail("Failed to delete redis keys %v: %v", keys, err)
	}
}

// ExpectValue asserts that a key has the expected value.
func (c *RedisClient) ExpectValue(key string, expected string) {
	if IsDryRun() {
		return
	}
	val := c.Get(key)
	if val != expected {
		Fail("Redis value mismatch for key %s: expected %s, got %s", key, expected, val)
	}
	Logf(LogTypeExpect, "Redis key %s == %s - PASSED", key, expected)
}

// FlushAll removes all keys from the current database.
func (c *RedisClient) FlushAll() {
	RecordAction("Redis FlushAll", func() { c.FlushAll() })
	if IsDryRun() {
		return
	}
	if c.client == nil {
		Fail("RedisClient is not connected")
	}
	Log(LogTypeRedis, "FLUSHALL", "")
	if err := c.client.FlushDB(context.Background()).Err(); err != nil {
		Fail("Failed to flush redis db: %v", err)
	}
}
