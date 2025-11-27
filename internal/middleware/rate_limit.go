package middleware

import (
	"gin-app-start/pkg/response"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	rate       int                  // 每秒允许的请求数
	lastAccess map[string]time.Time // 记录每个客户端的最后访问时间
	tokens     map[string]int       // 记录每个客户端当前可用的令牌数
	mu         sync.Mutex           // 互斥锁，用于保护对 lastAccess 和 tokens 的并发访问
}

// newRateLimiter 创建一个新的速率限制器
func newRateLimiter(rate int) *rateLimiter {
	limiter := &rateLimiter{
		rate:       rate,
		lastAccess: make(map[string]time.Time),
		tokens:     make(map[string]int),
	}

	go limiter.cleanup()

	return limiter
}

// allow 检查是否允许当前请求
// 基于令牌桶算法的速率限制
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	lastTime, exists := rl.lastAccess[key]

	// 如果是第一次访问，初始化令牌数为 rate - 1
	if !exists {
		rl.lastAccess[key] = now
		rl.tokens[key] = rl.rate - 1
		return true
	}

	// 计算距离上次访问经过了多少秒
	elapsed := now.Sub(lastTime).Seconds()

	// 计算距离上次访问经过了多少秒，将其转换为令牌数
	// 若rate=100（每秒最多100次请求），elapsed < 0.01s，tokensToAdd = 0
	tokensToAdd := int(elapsed * float64(rl.rate))

	if tokensToAdd > 0 {
		rl.tokens[key] += tokensToAdd
		if rl.tokens[key] > rl.rate {
			rl.tokens[key] = rl.rate
		}
		rl.lastAccess[key] = now
	}

	if rl.tokens[key] > 0 {
		rl.tokens[key]--
		return true
	}

	return false
}

// cleanup 定期清理过期的访问记录
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute) // 每分钟执行一次清理
	defer ticker.Stop()

	// 清理过期的访问记录，保留最近 5 分钟内的记录
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, lastTime := range rl.lastAccess {
			if now.Sub(lastTime) > 5*time.Minute {
				delete(rl.lastAccess, key)
				delete(rl.tokens, key)
			}
		}
		rl.mu.Unlock()
	}
}

var globalLimiter *rateLimiter

func RateLimit(rate int) gin.HandlerFunc {
	if globalLimiter == nil {
		globalLimiter = newRateLimiter(rate)
	}

	return func(c *gin.Context) {
		key := c.ClientIP()

		if !globalLimiter.allow(key) {
			response.Error(c, 42900, "Too many requests, please try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
