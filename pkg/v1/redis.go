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

// ExpectFound asserts that a key exists in Redis.
func (c *RedisClient) ExpectFound(key string) {
	if IsDryRun() {
		return
	}
	if c.client == nil {
		Fail("RedisClient is not connected")
	}
	exists, err := c.client.Exists(context.Background(), key).Result()
	if err != nil {
		Fail("Failed to check existence of redis key %s: %v", key, err)
	}
	if exists == 0 {
		Fail("Expected redis key %s to exist, but it was not found", key)
	}
	Logf(LogTypeExpect, "Redis key %s exists - PASSED", key)
}

// ExpectNotFound asserts that a key does not exist in Redis.
func (c *RedisClient) ExpectNotFound(key string) {
	if IsDryRun() {
		return
	}
	if c.client == nil {
		Fail("RedisClient is not connected")
	}
	exists, err := c.client.Exists(context.Background(), key).Result()
	if err != nil {
		Fail("Failed to check existence of redis key %s: %v", key, err)
	}
	if exists != 0 {
		Fail("Expected redis key %s to not exist, but it was found", key)
	}
	Logf(LogTypeExpect, "Redis key %s does not exist - PASSED", key)
}

// HSet sets a field in a hash.
func (c *RedisClient) HSet(key, field string, value interface{}) {
	RecordAction(fmt.Sprintf("Redis HSet: %s %s", key, field), func() { c.HSet(key, field, value) })
	if IsDryRun() {
		return
	}
	if c.client == nil {
		Fail("RedisClient is not connected")
	}
	Log(LogTypeRedis, fmt.Sprintf("HSET %s %s", key, field), fmt.Sprintf("value=%v", value))
	if err := c.client.HSet(context.Background(), key, field, value).Err(); err != nil {
		Fail("Failed to hset redis key %s field %s: %v", key, field, err)
	}
}

// HGet retrieves a field value from a hash.
func (c *RedisClient) HGet(key, field string) string {
	RecordAction(fmt.Sprintf("Redis HGet: %s %s", key, field), func() { c.HGet(key, field) })
	if IsDryRun() {
		return ""
	}
	if c.client == nil {
		Fail("RedisClient is not connected")
	}
	Logf(LogTypeRedis, "HGET %s %s", key, field)
	val, err := c.client.HGet(context.Background(), key, field).Result()
	if err != nil {
		if err == redis.Nil {
			Fail("Redis hash key %s field %s not found", key, field)
		}
		Fail("Failed to hget redis key %s field %s: %v", key, field, err)
	}
	return val
}

// HIncrement increments a hash field by the given integer amount.
func (c *RedisClient) HIncrement(key, field string, increment int64) int64 {
	RecordAction(fmt.Sprintf("Redis HIncrement: %s %s by %d", key, field, increment), func() {
		c.HIncrement(key, field, increment)
	})
	if IsDryRun() {
		return 0
	}
	if c.client == nil {
		Fail("RedisClient is not connected")
	}
	Logf(LogTypeRedis, "HINCRBY %s %s %d", key, field, increment)
	val, err := c.client.HIncrBy(context.Background(), key, field, increment).Result()
	if err != nil {
		Fail("Failed to hincrby redis key %s field %s: %v", key, field, err)
	}
	return val
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
