package session

import (
	"fmt"
	"sync"
	"testing"
)

const defaultShardCount = 32

func TestShardMapBasicOperations(t *testing.T) {
	m := NewInt64ShardMap[string](defaultShardCount)

	// Test Set and Get
	m.Set(1, "hello")
	v, ok := m.Get(1)
	if !ok || v != "hello" {
		t.Fatalf("expected 'hello', got '%s', ok=%v", v, ok)
	}

	// Test Has
	if !m.Has(1) {
		t.Fatal("expected Has(1) to be true")
	}
	if m.Has(2) {
		t.Fatal("expected Has(2) to be false")
	}

	// Test Delete
	m.Delete(1)
	_, ok = m.Get(1)
	if ok {
		t.Fatal("expected Get(1) to return false after Delete")
	}
}

func TestShardMapLen(t *testing.T) {
	m := NewStringShardMap[int](16)

	for i := 0; i < 100; i++ {
		m.Set(fmt.Sprintf("key_%d", i), i)
	}

	if m.Len() != 100 {
		t.Fatalf("expected Len() = 100, got %d", m.Len())
	}

	m.Delete("key_0")
	if m.Len() != 99 {
		t.Fatalf("expected Len() = 99, got %d", m.Len())
	}
}

func TestShardMapGetAndDelete(t *testing.T) {
	m := NewInt64ShardMap[string](8)
	m.Set(42, "value")

	v, ok := m.GetAndDelete(42)
	if !ok || v != "value" {
		t.Fatalf("expected 'value', got '%s', ok=%v", v, ok)
	}

	_, ok = m.Get(42)
	if ok {
		t.Fatal("expected key to be deleted after GetAndDelete")
	}

	// GetAndDelete on non-existent key
	_, ok = m.GetAndDelete(999)
	if ok {
		t.Fatal("expected GetAndDelete(999) to return false")
	}
}

func TestShardMapGetOrSet(t *testing.T) {
	m := NewInt64ShardMap[string](8)

	// First call should set the value
	v, existed := m.GetOrSet(1, "first")
	if existed || v != "first" {
		t.Fatalf("expected 'first' and not existed, got '%s', existed=%v", v, existed)
	}

	// Second call should return existing value
	v, existed = m.GetOrSet(1, "second")
	if !existed || v != "first" {
		t.Fatalf("expected 'first' and existed, got '%s', existed=%v", v, existed)
	}
}

func TestShardMapRangeAll(t *testing.T) {
	m := NewInt64ShardMap[int](4)
	for i := int64(0); i < 10; i++ {
		m.Set(i, int(i*10))
	}

	collected := make(map[int64]int)
	m.RangeAll(func(key int64, value int) bool {
		collected[key] = value
		return true
	})

	if len(collected) != 10 {
		t.Fatalf("expected 10 items, got %d", len(collected))
	}

	for i := int64(0); i < 10; i++ {
		if collected[i] != int(i*10) {
			t.Fatalf("expected %d for key %d, got %d", i*10, i, collected[i])
		}
	}
}

func TestShardMapRangeAllEarlyStop(t *testing.T) {
	m := NewInt64ShardMap[int](1) // single shard for deterministic ordering
	for i := int64(0); i < 5; i++ {
		m.Set(i, int(i))
	}

	count := 0
	m.RangeAll(func(key int64, value int) bool {
		count++
		return count < 3 // stop after 3 items
	})

	if count != 3 {
		t.Fatalf("expected exactly 3 iterations, got %d", count)
	}
}

func TestShardMapConcurrency(t *testing.T) {
	m := NewInt64ShardMap[int64](32)
	const goroutines = 100
	const opsPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines * 3) // writers + readers + deleters

	// Writers
	for g := 0; g < goroutines; g++ {
		go func(base int64) {
			defer wg.Done()
			for i := int64(0); i < opsPerGoroutine; i++ {
				m.Set(base+i, base+i)
			}
		}(int64(g * opsPerGoroutine))
	}

	// Readers
	for g := 0; g < goroutines; g++ {
		go func(base int64) {
			defer wg.Done()
			for i := int64(0); i < opsPerGoroutine; i++ {
				m.Get(base + i)
			}
		}(int64(g * opsPerGoroutine))
	}

	// Deleters
	for g := 0; g < goroutines; g++ {
		go func(base int64) {
			defer wg.Done()
			for i := int64(0); i < opsPerGoroutine; i++ {
				m.Delete(base + i)
			}
		}(int64(g * opsPerGoroutine))
	}

	wg.Wait()
	// If we get here without panics/deadlocks, concurrency is handled correctly
}

func TestShardMapWithShard(t *testing.T) {
	m := NewInt64ShardMap[string](8)
	m.Set(1, "hello")
	m.Set(2, "world")

	m.WithShard(1, func(items map[int64]string) {
		items[1] = "modified"
	})

	v, ok := m.Get(1)
	if !ok || v != "modified" {
		t.Fatalf("expected 'modified', got '%s'", v)
	}
}

func TestShardMapWithShardRLock(t *testing.T) {
	m := NewInt64ShardMap[string](8)
	m.Set(1, "test")

	var result string
	m.WithShardRLock(1, func(items map[int64]string) {
		result = items[1]
	})

	if result != "test" {
		t.Fatalf("expected 'test', got '%s'", result)
	}
}

func BenchmarkShardMapSet(b *testing.B) {
	m := NewInt64ShardMap[int64](32)
	b.RunParallel(func(pb *testing.PB) {
		var i int64
		for pb.Next() {
			m.Set(i, i)
			i++
		}
	})
}

func BenchmarkShardMapGet(b *testing.B) {
	m := NewInt64ShardMap[int64](32)
	for i := int64(0); i < 10000; i++ {
		m.Set(i, i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i int64
		for pb.Next() {
			m.Get(i % 10000)
			i++
		}
	})
}

func BenchmarkShardMapMixedReadWrite(b *testing.B) {
	m := NewInt64ShardMap[int64](32)
	for i := int64(0); i < 10000; i++ {
		m.Set(i, i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var i int64
		for pb.Next() {
			if i%10 == 0 {
				m.Set(i%10000, i)
			} else {
				m.Get(i % 10000)
			}
			i++
		}
	})
}
