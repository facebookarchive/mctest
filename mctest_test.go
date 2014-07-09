package mctest_test

import (
	"bytes"
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/facebookgo/mctest"
)

func test(t *testing.T, answer []byte) {
	t.Parallel()
	mc := mctest.NewStartedServer(t)
	defer mc.Stop()
	client := mc.Client()

	const key = "1"
	err := client.Set(&memcache.Item{
		Key:   key,
		Value: answer,
	})
	if err != nil {
		t.Fatal(err)
	}

	item, err := client.Get(key)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(answer, item.Value) {
		t.Fatalf("expected %s but got %s", answer, item.Value)
	}
}

// Test that multiple instances don't stomp on each other.
func TestOne(t *testing.T) {
	test(t, []byte("42"))
}

func TestTwo(t *testing.T) {
	test(t, []byte("43"))
}

func TestThree(t *testing.T) {
	test(t, []byte("44"))
}
