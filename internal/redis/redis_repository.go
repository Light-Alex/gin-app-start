package redis

import (
	"context"
	"fmt"
	"time"

	"gin-app-start/pkg/timeutil"
	"gin-app-start/pkg/trace"

	"github.com/redis/go-redis/v9"
)

type Option func(*option)

type Trace = trace.T

type option struct {
	Trace *trace.Trace
	Redis *trace.Redis
}

func newOption() *option {
	return &option{}
}

// 检查redisClient是否实现了RedisClient的全部接口
var _ RedisRepository = (*redisRepository)(nil)

type RedisRepository interface {
	// Set 设置键值对
	Set(key, value string, expiration time.Duration, options ...Option) error
	// Get 获取键的值
	Get(key string, options ...Option) (string, error)
	// Delete 删除键
	Delete(key string, options ...Option) error
	// Exists 检查键是否存在
	Exists(key string) (bool, error)
	// SetWithExpire 设置带过期时间的键值对
	SetWithExpire(key, value string, expiration time.Duration, options ...Option) error
	// Increment 对数字值进行递增
	Increment(key string, options ...Option) (int64, error)
	// ListRPush 从右侧推入列表元素
	ListRPush(key string, values ...interface{}) error
	// ListLLen 获取列表长度
	ListLLen(key string) (int64, error)
	// ListLPop 从左侧弹出列表元素
	ListLPop(key string) (string, error)
	// ListLRange 获取列表指定范围的元素
	ListLRange(key string, start, stop int64) ([]string, error)
	// SetSAdd 添加元素到集合
	SetSAdd(key string, members ...interface{}) error
	// SetSRem 移除集合中的元素
	SetSRem(key string, members ...interface{}) error
	// SetSMembers 获取集合所有元素
	SetSMembers(key string) ([]string, error)
	// SetSIsMember 检查元素是否在集合中
	SetSIsMember(key string, member interface{}) (bool, error)
	// SetSCard 获取集合元素数量
	SetSCard(key string) (int64, error)
	// SetSRandMember 随机获取集合中的一个元素
	SetSRandMember(key string) (string, error)
	// SetZAdd 添加/更新有序集合中的元素（带分数）
	SetZAdd(key string, members ...redis.Z) error
	// SetZRem 移除有序集合中的元素
	SetZRem(key string, members ...interface{}) error
	// SetZRange 获取有序集合指定范围的元素(按分数升序)
	SetZRange(key string, start, stop int64) ([]string, error)
	// SetZRevRange 获取有序集合指定范围的元素(按分数降序)
	SetZRevRange(key string, start, stop int64) ([]string, error)
	// SetZCard 获取有序集合元素数量
	SetZCard(key string) (int64, error)
	// SetZRangeByScore 获取有序集合指定分数范围内的元素(按分数升序)
	SetZRangeByScore(key string, min, max string, start, stop int64) ([]string, error)
	// SetZRevRangeByScore 获取有序集合指定分数范围内的元素(按分数降序)
	SetZRevRangeByScore(key string, min, max string, start, stop int64) ([]string, error)
	// SetZScore 获取有序集合中元素的分数
	SetZScore(key string, member string) error
	// SetZIncrBy 增加有序集合中元素的分数
	SetZIncrBy(key string, member string, increment float64) error
	// SetZRank 获取有序集合中元素的排名（按分数升序）
	SetZRank(key string, member string) error
	// SetZRevRank 获取有序集合中元素的排名（按分数降序）
	SetZRevRank(key string, member string) error
	// SetHashSet 设置哈希字段
	HashSet(hashKey string, expireTime time.Duration, params HashParams) error
	// SetHashGetAll 获取哈希字段的所有值
	HashGetAll(hashKey string) (map[string]string, error)
	// SetHashGet 获取哈希字段的值
	HashGet(hashKey string, field string) (string, error)
	// GetRedisContext 获取Redis上下文
	GetRedisContext() context.Context
	// GetRedisClient 获取Redis客户端
	GetRedisClient() *redis.Client
	// Close 关闭Redis连接
	Close()
}

// redisRepository 封装Redis客户端
type redisRepository struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisRepository(client *redis.Client, ctx context.Context) RedisRepository {
	return &redisRepository{client: client, ctx: ctx}
}

// Set 设置键值对
func (rc *redisRepository) Set(key, value string, expiration time.Duration, options ...Option) error {
	start := time.Now()
	opt := newOption()
	defer func() {
		if opt.Trace != nil {
			opt.Redis.Timestamp = timeutil.CSTLayoutString()
			opt.Redis.Handle = "Set"
			opt.Redis.Key = key
			opt.Redis.Value = value
			opt.Redis.TTL = expiration.Minutes()
			opt.Redis.CostSeconds = time.Since(start).Seconds()
			opt.Trace.AppendRedis(opt.Redis)
		}
	}()

	for _, f := range options {
		f(opt)
	}

	err := rc.client.Set(rc.ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("redis set %s -> %s failed: %w", key, value, err)
	}
	return nil
}

// Get 获取键的值
func (rc *redisRepository) Get(key string, options ...Option) (string, error) {
	start := time.Now()
	opt := newOption()
	defer func() {
		if opt.Trace != nil {
			opt.Redis.Timestamp = timeutil.CSTLayoutString()
			opt.Redis.Handle = "Get"
			opt.Redis.Key = key
			opt.Redis.CostSeconds = time.Since(start).Seconds()
			opt.Trace.AppendRedis(opt.Redis)
		}
	}()

	for _, f := range options {
		f(opt)
	}

	value, err := rc.client.Get(rc.ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("redis key %s does not exist", key)
	} else if err != nil {
		return "", fmt.Errorf("redis get key %s failed: %v", key, err)
	}
	return value, nil
}

// Delete 删除键
func (rc *redisRepository) Delete(key string, options ...Option) error {
	start := time.Now()
	opt := newOption()
	defer func() {
		if opt.Trace != nil {
			opt.Redis.Timestamp = timeutil.CSTLayoutString()
			opt.Redis.Handle = "Delete"
			opt.Redis.Key = key
			opt.Redis.CostSeconds = time.Since(start).Seconds()
			opt.Trace.AppendRedis(opt.Redis)
		}
	}()

	for _, f := range options {
		f(opt)
	}

	err := rc.client.Del(rc.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis delete key %s failed: %w", key, err)
	}
	return nil
}

// Exists 检查键是否存在
func (rc *redisRepository) Exists(key string) (bool, error) {
	result, err := rc.client.Exists(rc.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis check key %s existence failed: %w", key, err)
	}
	exists := result > 0
	return exists, nil
}

// SetWithExpire 设置带过期时间的键值对
func (rc *redisRepository) SetWithExpire(key, value string, expiration time.Duration, options ...Option) error {
	start := time.Now()
	opt := newOption()
	defer func() {
		if opt.Trace != nil {
			opt.Redis.Timestamp = timeutil.CSTLayoutString()
			opt.Redis.Handle = "SetWithExpire"
			opt.Redis.Key = key
			opt.Redis.Value = value
			opt.Redis.TTL = expiration.Minutes()
			opt.Redis.CostSeconds = time.Since(start).Seconds()
			opt.Trace.AppendRedis(opt.Redis)
		}
	}()

	for _, f := range options {
		f(opt)
	}

	err := rc.client.SetEx(rc.ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("redis set %s -> %s with expiration %v failed: %w", key, value, expiration, err)
	}
	return nil
}

// Increment 对数字值进行递增
func (rc *redisRepository) Increment(key string, options ...Option) (int64, error) {
	start := time.Now()
	opt := newOption()
	defer func() {
		if opt.Trace != nil {
			opt.Redis.Timestamp = timeutil.CSTLayoutString()
			opt.Redis.Handle = "Increment"
			opt.Redis.Key = key
			opt.Redis.CostSeconds = time.Since(start).Seconds()
			opt.Trace.AppendRedis(opt.Redis)
		}
	}()

	for _, f := range options {
		f(opt)
	}

	result, err := rc.client.Incr(rc.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis increment key %s failed: %w", key, err)
	}
	return result, nil
}

// ListRPush 从右侧推入列表元素
func (rc *redisRepository) ListRPush(key string, values ...interface{}) error {
	err := rc.client.RPush(rc.ctx, key, values...).Err()
	if err != nil {
		return fmt.Errorf("redis list rpush %s -> %v failed: %w", key, values, err)
	}
	return nil
}

// ListLLen 获取列表长度
func (rc *redisRepository) ListLLen(key string) (int64, error) {
	length, err := rc.client.LLen(rc.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis list llen %s failed: %w", key, err)
	}
	return length, nil
}

// ListLPop 从左侧弹出列表元素
func (rc *redisRepository) ListLPop(key string) (string, error) {
	value, err := rc.client.LPop(rc.ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("redis list %s is empty", key)
	} else if err != nil {
		return "", fmt.Errorf("redis list lpop %s failed: %w", key, err)
	}
	return value, nil
}

// ListLRange 获取列表指定范围的元素[start, stop]
func (rc *redisRepository) ListLRange(key string, start, stop int64) ([]string, error) {
	items, err := rc.client.LRange(rc.ctx, key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("redis list lrange %s failed: %w", key, err)
	}
	return items, nil
}

// SetSAdd 添加元素到集合
func (rc *redisRepository) SetSAdd(key string, members ...interface{}) error {
	err := rc.client.SAdd(rc.ctx, key, members...).Err()
	if err != nil {
		return fmt.Errorf("redis set sadd %s -> %v failed: %w", key, members, err)
	}
	return nil
}

// SetSRem 移除集合中的元素
func (rc *redisRepository) SetSRem(key string, members ...interface{}) error {
	err := rc.client.SRem(rc.ctx, key, members...).Err()
	if err != nil {
		return fmt.Errorf("redis set srem %s -> %v failed: %w", key, members, err)
	}
	return nil
}

// SetSMembers 获取集合所有元素
func (rc *redisRepository) SetSMembers(key string) ([]string, error) {
	members, err := rc.client.SMembers(rc.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis set smembers %s failed: %w", key, err)
	}
	return members, nil
}

// SetSIsMember 检查元素是否在集合中
func (rc *redisRepository) SetSIsMember(key string, member interface{}) (bool, error) {
	isMember, err := rc.client.SIsMember(rc.ctx, key, member).Result()
	if err != nil {
		return false, fmt.Errorf("redis set smember %s -> %v failed: %w", key, member, err)
	}
	return isMember, nil
}

// SetSCard 获取集合元素数量
func (rc *redisRepository) SetSCard(key string) (int64, error) {
	cardinality, err := rc.client.SCard(rc.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis set scard %s failed: %w", key, err)
	}
	return cardinality, nil
}

// SetSRandMember 随机获取集合中的一个元素
func (rc *redisRepository) SetSRandMember(key string) (string, error) {
	randomMember, err := rc.client.SRandMember(rc.ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("redis set srandmember %s failed: %w", key, err)
	}
	return randomMember, nil
}

// SetZAdd 添加/更新有序集合中的元素（带分数）
func (rc *redisRepository) SetZAdd(key string, members ...redis.Z) error {
	err := rc.client.ZAdd(rc.ctx, key, members...).Err()
	if err != nil {
		return fmt.Errorf("redis set zadd %s -> %v failed: %w", key, members, err)
	}
	return nil
}

// SetZRem 移除有序集合中的元素
func (rc *redisRepository) SetZRem(key string, members ...interface{}) error {
	err := rc.client.ZRem(rc.ctx, key, members...).Err()
	if err != nil {
		return fmt.Errorf("redis set zrem %s -> %v failed: %w", key, members, err)
	}
	return nil
}

// SetZRange 获取有序集合指定范围的元素(按分数升序) [start, stop]
func (rc *redisRepository) SetZRange(key string, start, stop int64) ([]string, error) {
	members, err := rc.client.ZRange(rc.ctx, key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("redis set zrange %s failed: %w", key, err)
	}
	return members, nil
}

// SetZRevRange 获取有序集合指定范围的元素(按分数降序) [start, stop]
func (rc *redisRepository) SetZRevRange(key string, start, stop int64) ([]string, error) {
	members, err := rc.client.ZRevRange(rc.ctx, key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("redis set zrevrange %s failed: %w", key, err)
	}
	return members, nil
}

// SetZCard 获取有序集合元素数量
func (rc *redisRepository) SetZCard(key string) (int64, error) {
	cardinality, err := rc.client.ZCard(rc.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis set zcard %s failed: %w", key, err)
	}
	return cardinality, nil
}

// SetZRangeByScore 获取有序集合指定分数范围内的元素(按分数升序) [min, max] [start, stop]
func (rc *redisRepository) SetZRangeByScore(key string, min, max string, start, stop int64) ([]string, error) {
	if min > max {
		return nil, fmt.Errorf("min[%s] must less than max[%s]", min, max)
	}

	members, err := rc.client.ZRangeByScore(rc.ctx, key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: start,
		Count:  stop - start + 1,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("redis set ZRangeByScore failed: %w", err)
	}
	return members, nil
}

// SetZRevRangeByScore 获取有序集合指定分数范围内的元素(按分数降序) [min, max] [start, stop]
func (rc *redisRepository) SetZRevRangeByScore(key string, min, max string, start, stop int64) ([]string, error) {
	if min > max {
		return nil, fmt.Errorf("min[%s] must less than max[%s]", min, max)
	}

	members, err := rc.client.ZRevRangeByScore(rc.ctx, key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: start,
		Count:  stop - start + 1,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("redis set ZRevRangeByScore failed: %w", err)
	}
	return members, nil
}

// SetZScore 获取有序集合中元素的分数
func (rc *redisRepository) SetZScore(key string, member string) error {
	_, err := rc.client.ZScore(rc.ctx, key, member).Result()
	if err != nil {
		return fmt.Errorf("redis set ZScore failed: %w", err)
	}
	return nil
}

// SetZIncrBy 增加有序集合中元素的分数
func (rc *redisRepository) SetZIncrBy(key string, member string, increment float64) error {
	_, err := rc.client.ZIncrBy(rc.ctx, key, increment, member).Result()
	if err != nil {
		return fmt.Errorf("redis set ZIncrBy failed: %w", err)
	}
	return nil
}

// SetZRank 获取有序集合中元素的排名（按分数升序）
func (rc *redisRepository) SetZRank(key string, member string) error {
	_, err := rc.client.ZRank(rc.ctx, key, member).Result()
	if err != nil {
		return fmt.Errorf("redis set ZRank failed: %w", err)
	}
	return nil
}

// SetZRevRank 获取有序集合中元素的排名（按分数降序）
func (rc *redisRepository) SetZRevRank(key string, member string) error {
	_, err := rc.client.ZRevRank(rc.ctx, key, member).Result()
	if err != nil {
		return fmt.Errorf("redis set ZRevRank failed: %w", err)
	}
	return nil
}

type HashParams struct {
	Options []Option
	Values  []interface{}
}

// SetHashSet 设置哈希字段
func (rc *redisRepository) HashSet(hashKey string, expireTime time.Duration, params HashParams) error {
	start := time.Now()
	opt := newOption()
	defer func() {
		if opt.Trace != nil {
			opt.Redis.Timestamp = timeutil.CSTLayoutString()
			opt.Redis.Handle = "HashSet"
			opt.Redis.Key = hashKey
			opt.Redis.Values = params.Values
			opt.Redis.TTL = expireTime.Minutes()
			opt.Redis.CostSeconds = time.Since(start).Seconds()
			opt.Trace.AppendRedis(opt.Redis)
		}
	}()

	for _, f := range params.Options {
		f(opt)
	}

	pipe := rc.client.TxPipeline()
	pipe.HSet(rc.ctx, hashKey, params.Values...)
	pipe.Expire(rc.ctx, hashKey, expireTime).Err()
	_, err := pipe.Exec(rc.ctx)
	if err != nil {
		return fmt.Errorf("redis set HashSet failed: %w", err)
	}

	return nil
}

// SetHashGetAll 获取哈希字段的所有值
func (rc *redisRepository) HashGetAll(hashKey string) (map[string]string, error) {
	fields, err := rc.client.HGetAll(rc.ctx, hashKey).Result()
	if err != nil {
		return nil, fmt.Errorf("redis set HashGetAll failed: %w", err)
	}
	return fields, nil
}

// SetHashGet 获取哈希字段的值
func (rc *redisRepository) HashGet(hashKey string, field string) (string, error) {
	value, err := rc.client.HGet(rc.ctx, hashKey, field).Result()
	if err != nil {
		return "", fmt.Errorf("redis set HashGet failed: %w", err)
	}
	return value, nil
}

// GetRedisContext 获取Redis上下文
func (rc *redisRepository) GetRedisContext() context.Context {
	return rc.ctx
}

// GetRedisClient 获取Redis客户端
func (rc *redisRepository) GetRedisClient() *redis.Client {
	return rc.client
}

// Close 关闭Redis连接
func (rc *redisRepository) Close() {
	if rc.client != nil {
		rc.client.Close()
	}
}

// WithTrace 设置trace信息
func WithTrace(t Trace) Option {
	return func(opt *option) {
		if t != nil {
			opt.Trace = t.(*trace.Trace)
			opt.Redis = new(trace.Redis)
		}
	}
}
