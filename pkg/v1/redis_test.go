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
