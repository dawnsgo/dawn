package session

import (
	"hash/fnv"
	"sync"
)

// ShardMap 通用分片锁 Map，减少锁竞争
// 使用泛型替代重复的 connShard/userShard/channelShard 实现
type ShardMap[K comparable, V any] struct {
	shards []*shard[K, V]
	count  int
	hasher func(key K) uint32
}

// shard 单个分片
type shard[K comparable, V any] struct {
	sync.RWMutex
	items map[K]V
}

// NewShardMap 创建分片锁 Map
func NewShardMap[K comparable, V any](shardCount int, hasher func(K) uint32) *ShardMap[K, V] {
	if shardCount <= 0 {
		shardCount = defaultShardCount
	}

	m := &ShardMap[K, V]{
		shards: make([]*shard[K, V], shardCount),
		count:  shardCount,
		hasher: hasher,
	}

	for i := 0; i < shardCount; i++ {
		m.shards[i] = &shard[K, V]{
			items: make(map[K]V),
		}
	}

	return m
}

// NewInt64ShardMap 创建 int64 键的分片锁 Map（常用场景的便捷构造函数）
func NewInt64ShardMap[V any](shardCount int) *ShardMap[int64, V] {
	return NewShardMap[int64, V](shardCount, func(key int64) uint32 {
		return uint32(key)
	})
}

// NewStringShardMap 创建 string 键的分片锁 Map（常用场景的便捷构造函数）
func NewStringShardMap[V any](shardCount int) *ShardMap[string, V] {
	return NewShardMap[string, V](shardCount, func(key string) uint32 {
		h := fnv.New32a()
		h.Write([]byte(key))
		return h.Sum32()
	})
}

// getShard 获取键对应的分片
func (m *ShardMap[K, V]) getShard(key K) *shard[K, V] {
	return m.shards[m.hasher(key)%uint32(m.count)]
}

// Get 获取值
func (m *ShardMap[K, V]) Get(key K) (V, bool) {
	s := m.getShard(key)
	s.RLock()
	v, ok := s.items[key]
	s.RUnlock()
	return v, ok
}

// Set 设置值
func (m *ShardMap[K, V]) Set(key K, value V) {
	s := m.getShard(key)
	s.Lock()
	s.items[key] = value
	s.Unlock()
}

// Delete 删除值
func (m *ShardMap[K, V]) Delete(key K) {
	s := m.getShard(key)
	s.Lock()
	delete(s.items, key)
	s.Unlock()
}

// Has 检查键是否存在
func (m *ShardMap[K, V]) Has(key K) bool {
	s := m.getShard(key)
	s.RLock()
	_, ok := s.items[key]
	s.RUnlock()
	return ok
}

// Len 获取所有分片的总元素数
func (m *ShardMap[K, V]) Len() int64 {
	var total int64
	for _, s := range m.shards {
		s.RLock()
		total += int64(len(s.items))
		s.RUnlock()
	}
	return total
}

// GetAndDelete 获取并删除值（原子操作）
func (m *ShardMap[K, V]) GetAndDelete(key K) (V, bool) {
	s := m.getShard(key)
	s.Lock()
	v, ok := s.items[key]
	if ok {
		delete(s.items, key)
	}
	s.Unlock()
	return v, ok
}

// GetOrSet 获取或设置值（如果不存在则设置，返回最终值和是否已存在）
func (m *ShardMap[K, V]) GetOrSet(key K, value V) (V, bool) {
	s := m.getShard(key)
	s.Lock()
	if v, ok := s.items[key]; ok {
		s.Unlock()
		return v, true
	}
	s.items[key] = value
	s.Unlock()
	return value, false
}

// RangeAll 遍历所有分片中的所有元素
// fn 返回 false 时停止遍历
func (m *ShardMap[K, V]) RangeAll(fn func(key K, value V) bool) {
	for _, s := range m.shards {
		s.RLock()
		for k, v := range s.items {
			if !fn(k, v) {
				s.RUnlock()
				return
			}
		}
		s.RUnlock()
	}
}

// WithShard 在指定键的分片锁保护下执行操作（用于复杂的原子操作）
func (m *ShardMap[K, V]) WithShard(key K, fn func(items map[K]V)) {
	s := m.getShard(key)
	s.Lock()
	fn(s.items)
	s.Unlock()
}

// WithShardRLock 在指定键的分片读锁保护下执行操作
func (m *ShardMap[K, V]) WithShardRLock(key K, fn func(items map[K]V)) {
	s := m.getShard(key)
	s.RLock()
	fn(s.items)
	s.RUnlock()
}
