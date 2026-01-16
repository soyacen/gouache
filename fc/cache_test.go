package fc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/soyacen/gouache"
)

// TestStruct 是用于测试的自定义结构体
type TestStruct struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// 测试基本的Set和Get操作
func TestCache_SetGet(t *testing.T) {
	// 创建一个缓存实例
	cache := &Cache{
		Cache: freecache.NewCache(1024 * 1024), // 1MB缓存
	}

	ctx := context.Background()
	key := "test_key"
	value := []byte("test_value")

	// 设置值
	err := cache.Set(ctx, key, value)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 获取值
	result, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(result.([]byte)) != string(value) {
		t.Errorf("expected %s, got %s", string(value), string(result.([]byte)))
	}

	// 获取不存在的键
	_, err = cache.Get(ctx, "non_existent_key")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, gouache.ErrCacheMiss) {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
}

// 测试使用自定义Marshal和Unmarshal函数
func TestCache_SetGetWithMarshalUnmarshal(t *testing.T) {
	cache := &Cache{
		Cache: freecache.NewCache(1024 * 1024),
		Marshal: func(key string, obj any) ([]byte, error) {
			return json.Marshal(obj)
		},
		Unmarshal: func(key string, data []byte) (any, error) {
			var obj TestStruct
			err := json.Unmarshal(data, &obj)
			return &obj, err
		},
	}

	ctx := context.Background()
	key := "struct_key"
	value := &TestStruct{ID: 1, Name: "test"}

	// 设置结构体值
	err := cache.Set(ctx, key, value)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 获取结构体值
	result, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 类型断言并验证值
	resultStruct, ok := result.(*TestStruct)
	if !ok {
		t.Fatal("expected *TestStruct type")
	}
	if resultStruct.ID != value.ID {
		t.Errorf("expected ID %d, got %d", value.ID, resultStruct.ID)
	}
	if resultStruct.Name != value.Name {
		t.Errorf("expected Name %s, got %s", value.Name, resultStruct.Name)
	}
}

// 测试TTL功能
func TestCache_TTL(t *testing.T) {
	cache := &Cache{
		Cache: freecache.NewCache(1024 * 1024),
		TTL: func(ctx context.Context, key string, val any) (time.Duration, error) {
			return 1 * time.Second, nil // 设置1秒过期时间
		},
	}

	ctx := context.Background()
	key := "expiring_key"
	value := []byte("expiring_value")

	// 设置值
	err := cache.Set(ctx, key, value)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 立即获取应该成功
	result, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(result.([]byte)) != string(value) {
		t.Errorf("expected %s, got %s", string(value), string(result.([]byte)))
	}

	// 等待过期
	time.Sleep(1 * time.Second)

	// 再次获取应该失败
	_, err = cache.Get(ctx, key)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, gouache.ErrCacheMiss) {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
}

// 测试Delete操作
func TestCache_Delete(t *testing.T) {
	cache := &Cache{
		Cache: freecache.NewCache(1024 * 1024),
	}

	ctx := context.Background()
	key := "delete_key"
	value := []byte("delete_value")

	// 设置值
	err := cache.Set(ctx, key, value)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 删除值
	err = cache.Delete(ctx, key)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 尝试获取已删除的值
	_, err = cache.Get(ctx, key)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, gouache.ErrCacheMiss) {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
}

// 测试没有Marshal函数时设置非字节切片值的情况
func TestCache_SetWithoutMarshal(t *testing.T) {
	cache := &Cache{
		Cache: freecache.NewCache(1024 * 1024),
		// 没有设置Marshal函数
	}

	ctx := context.Background()
	key := "struct_key"
	value := &TestStruct{ID: 1, Name: "test"}

	// 应该返回错误
	err := cache.Set(ctx, key, value)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "gouache: Marshal is nil" {
		t.Errorf("expected 'gouache: Marshal is nil', got %v", err)
	}
}

// 测试TTL函数返回错误的情况
func TestCache_TTLWithError(t *testing.T) {
	expectedErr := errors.New("ttl error")
	cache := &Cache{
		Cache: freecache.NewCache(1024 * 1024),
		TTL: func(ctx context.Context, key string, val any) (time.Duration, error) {
			return 0, expectedErr
		},
	}

	ctx := context.Background()
	key := "error_key"
	value := []byte("test_value")

	// 设置值时应该返回TTL函数的错误
	err := cache.Set(ctx, key, value)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

// 测试Unmarshal函数返回错误的情况
func TestCache_GetWithUnmarshalError(t *testing.T) {
	expectedErr := errors.New("unmarshal error")
	cache := &Cache{
		Cache: freecache.NewCache(1024 * 1024),
		Marshal: func(key string, obj any) ([]byte, error) {
			return json.Marshal(obj)
		},
		Unmarshal: func(key string, data []byte) (any, error) {
			return nil, expectedErr
		},
	}

	ctx := context.Background()
	key := "error_key"
	value := []byte("test_value")

	// 先设置一个字节值
	err := cache.Cache.Set([]byte(key), value, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 获取时应该返回Unmarshal函数的错误
	_, err = cache.Get(ctx, key)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

// 测试Cache实现gouache.Cache接口
func TestCache_InterfaceImplementation(t *testing.T) {
	var _ gouache.Cache = (*Cache)(nil)
}
