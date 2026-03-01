package v1

import (
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
)

func TestRedisHelpers(t *testing.T) {
	// start in-memory redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := ConnectRedis(mr.Addr(), "", 0)

	client.Set("foo", "bar", time.Minute)
	client.ExpectValue("foo", "bar")

	client.Set("num", 123, 0)
	if got := client.Get("num"); got != "123" {
		t.Fatalf("expected num=123, got %s", got)
	}

	client.Del("foo")
	if _, err := mr.Get("foo"); err == nil {
		t.Fatalf("expected foo to be deleted")
	}

	client.FlushAll()
	if keys := mr.Keys(); len(keys) != 0 {
		t.Fatalf("expected empty db after flush, got %v", keys)
	}
}

func TestRedisHashHelpers(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := ConnectRedis(mr.Addr(), "", 0)

	// HSet and HGet
	client.HSet("user:1", "name", "Alice")
	client.HSet("user:1", "age", "30")

	if got := client.HGet("user:1", "name"); got != "Alice" {
		t.Fatalf("expected name=Alice, got %s", got)
	}
	if got := client.HGet("user:1", "age"); got != "30" {
		t.Fatalf("expected age=30, got %s", got)
	}

	// HIncrement
	client.HSet("counters", "visits", "10")
	result := client.HIncrement("counters", "visits", 5)
	if result != 15 {
		t.Fatalf("expected visits=15 after increment, got %d", result)
	}
	if got := client.HGet("counters", "visits"); got != "15" {
		t.Fatalf("expected visits=15 in redis, got %s", got)
	}

	// HIncrement on non-existing field starts from 0
	result = client.HIncrement("counters", "newfield", 3)
	if result != 3 {
		t.Fatalf("expected newfield=3, got %d", result)
	}
}
