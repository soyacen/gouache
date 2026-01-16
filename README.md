# Gouache - Go 缓存接口和实现

Gouache 是一个 Go 语言缓存库，实现了一致性缓存方案：延迟双删模式，提供统一的缓存接口和多种实现方式，底层实现包括内存缓存（sync.Map）、Redis、BigCache、LRU、Go-Cache 和 FreeCache 等。

## 特性

- **统一接口**: 定义了标准的 ***Cache*** 和 ***Database*** 接口
- **多种实现**: 提供多种缓存实现，包括:
  - 延迟双删缓存 (`ddd`)
  - 分片缓存 (`sharded`)
  - 防击穿缓存 (`sf`)
  - 基于内存的简单实现 (`sample`)
  - 带过期时间的内存缓存 (`gc`)
  - Redis 分布式缓存 (`redis`)
  - LRU 缓存 (`lru`)
  - BigCache 高性能缓存 (`bc`)
  - FreeCache 高性能缓存 (`fc`)
- **可扩展**: 易于添加新的缓存实现
- **线程安全**: 所有实现都支持并发访问

## 安装

```bash
go get github.com/soyacen/gouache
```

## 核心接口

### Cache 接口

```go
type Cache interface {
    Get(ctx context.Context, key string) (any, error)
    Set(ctx context.Context, key string, val any) error
    Delete(ctx context.Context, key string) error
}
```

### Database 接口

```go
type Database interface {
    Select(ctx context.Context, key string) (any, error)
    Upsert(ctx context.Context, key string, val any) error
    Delete(ctx context.Context, key string) error
}
```

## 使用示例

### 基础使用

```go
import "github.com/soyacen/gouache/sample"

// 创建简单内存缓存
cache := &sample.Cache{}

// 设置值
err := cache.Set(context.Background(), "key", "value")
if err != nil {
    // 处理错误
}

// 获取值
val, err := cache.Get(context.Background(), "key")
if err != nil {
    if errors.Is(err, gouache.ErrCacheMiss) {
        // 缓存未命中
    }
    // 处理其他错误
}

// 删除值
err = cache.Delete(context.Background(), "key")
```


### 延迟双删缓存

```go
import "github.com/soyacen/gouache/ddd"

// 使用延迟双删模式保证缓存与数据库一致性
cache := ddd.New(
    memoryCache,  // 缓存实现
    database,     // 数据库实现
    ddd.WithDelayDuration(500*time.Millisecond), // 延迟时间
)
```

### Redis 实现

```go
import (
    "github.com/soyacen/gouache/redis"
    "github.com/redis/go-redis/v9"
)

rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

cache := &redis.Cache{
    Cache: rdb,
}

// 使用缓存
err := cache.Set(context.Background(), "key", "value")
```

### LRU 缓存

```go
import "github.com/soyacen/gouache/lru"

lruCache, _ := lrucache.New(1000) // 最大容量1000
cache := &lru.Cache{
    Cache: lruCache,
}

err := cache.Set(context.Background(), "key", "value")
```

### FreeCache 实现

```go
import (
    "github.com/soyacen/gouache/fc"
    "github.com/coocood/freecache"
)

freeCache := fc.NewCache(100 * 1024 * 1024) // 100MB
cache := &fc.Cache{
    Cache: freeCache,
}

err := cache.Set(context.Background(), "key", []byte("value"))
```

### 组合使用 - 防击穿缓存

```go
import "github.com/soyacen/gouache/sf"

// 使用 singleflight 包装缓存防止击穿
cache := &sf.Cache{
    Cache: underlyingCache, // 任意其他缓存实现
}
```

## 各实现说明

| 实现 | 描述 | 特点 |
|------|------|------|
| `ddd` | 延迟双删缓存 | 保证缓存与数据库一致性 |
| `sharded` | 分片缓存 | 减少锁竞争，提高并发性能 |
| `sf` | 防击穿缓存 | 使用 singleflight 防止缓存击穿 |
| `sample` | 基于 `sync.Map` 的简单内存缓存 | 轻量、无依赖、线程安全 |
| `gc` | 基于 `patrickmn/go-cache` 的内存缓存 | 支持过期时间、LRU 清理 |
| `lru` | 基于 `hashicorp/golang-lru` 的 LRU 缓存 | 自动淘汰最久未使用项 |
| `bc` | 基于 `allegro/bigcache` 的高性能缓存 | 高并发、低内存占用 |
| `fc` | 基于 `coocood/freecache` 的高性能缓存 | 零GC、高并发 |
| `redis` | Redis 分布式缓存实现 | 支持分布式、持久化 |


## 错误处理

库定义了标准的缓存未命中错误:

```go
var ErrCacheMiss = errors.New("gouache: key not found")
```

使用时应检查此错误以区分缓存未命中和其他错误情况。

## 许可证

MIT
