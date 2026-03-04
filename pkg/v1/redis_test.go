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

func TestRedisExpectFound(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := ConnectRedis(mr.Addr(), "", 0)

	client.Set("existing", "value", time.Minute)
	client.ExpectFound("existing")
}

func TestRedisExpectNotFound(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := ConnectRedis(mr.Addr(), "", 0)

	client.ExpectNotFound("missing")
}

func TestRedisSetJsonField(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := ConnectRedis(mr.Addr(), "", 0)

	// Store initial JSON
	client.Set("eeeee", `{"a":{"b":[{"c":"old"}]}}`, 0)

	// Set nested field
	client.SetJsonField("eeeee", "a.b[0].c", "something")

	// Verify via ExpectJsonField
	client.ExpectJsonField("eeeee", "a.b[0].c", "something")

	// Set a top-level field
	client.Set("obj", `{"name":"alice","score":10}`, 0)
	client.SetJsonField("obj", "score", 99)
	client.ExpectJsonField("obj", "score", float64(99))

	// Set a field two levels deep
	client.Set("deep", `{"x":{"y":{"z":"before"}}}`, 0)
	client.SetJsonField("deep", "x.y.z", "after")
	client.ExpectJsonField("deep", "x.y.z", "after")
}

func TestRedisSetJsonFieldFailsOnMissingKey(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := ConnectRedis(mr.Addr(), "", 0)

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		client.SetJsonField("nonexistent", "a.b", "val")
	}()

	if !panicked {
		t.Fatal("expected Fail (panic) for missing key")
	}
}

func TestRedisExpectJsonFieldFailsOnDecodeError(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := ConnectRedis(mr.Addr(), "", 0)
	client.Set("bad", "not-json", 0)

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		client.ExpectJsonField("bad", "a", "x")
	}()

	if !panicked {
		t.Fatal("expected Fail (panic) for decode error")
	}
}

func TestRedisHSetJsonField(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := ConnectRedis(mr.Addr(), "", 0)

	// Store initial JSON in a hash field
	client.HSet("user:1", "profile", `{"name":"alice","score":10,"address":{"city":"Bangkok"}}`)

	// Set a top-level field
	client.HSetJsonField("user:1", "profile", "score", 99)
	client.HExpectJsonField("user:1", "profile", "score", float64(99))

	// Set a nested field
	client.HSetJsonField("user:1", "profile", "address.city", "Chiang Mai")
	client.HExpectJsonField("user:1", "profile", "address.city", "Chiang Mai")

	// Set a field with array index
	client.HSet("user:2", "data", `{"items":[{"id":1},{"id":2}]}`)
	client.HSetJsonField("user:2", "data", "items[0].id", 42)
	client.HExpectJsonField("user:2", "data", "items[0].id", float64(42))

	// Set deep nested path
	client.HSet("user:3", "config", `{"a":{"b":[{"c":"old"}]}}`)
	client.HSetJsonField("user:3", "config", "a.b[0].c", "new")
	client.HExpectJsonField("user:3", "config", "a.b[0].c", "new")
}

func TestRedisHSetJsonFieldFailsOnMissingField(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := ConnectRedis(mr.Addr(), "", 0)

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		client.HSetJsonField("nonexistent", "field", "a.b", "val")
	}()

	if !panicked {
		t.Fatal("expected Fail (panic) for missing hash key/field")
	}
}

func TestRedisHExpectJsonFieldFailsOnDecodeError(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := ConnectRedis(mr.Addr(), "", 0)
	client.HSet("myhash", "badjson", "not-json")

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		client.HExpectJsonField("myhash", "badjson", "a", "x")
	}()

	if !panicked {
		t.Fatal("expected Fail (panic) for decode error")
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
