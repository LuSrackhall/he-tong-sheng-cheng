package security

import (
	"sync"
	"time"
)

const (
	maxAttempts  = 5           // 最大失败次数
	windowPeriod = 5 * time.Minute // 时间窗口
	cleanupInterval = 10 * time.Minute // 清理间隔
)

type loginAttempt struct {
	count   int
	firstAt time.Time
}

// LoginRateLimiter 基于 IP 的登录暴力破解防护
type LoginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
}

// NewLoginRateLimiter 创建限流器并启动后台清理
func NewLoginRateLimiter() *LoginRateLimiter {
	l := &LoginRateLimiter{
		attempts: make(map[string]*loginAttempt),
	}
	go l.cleanup()
	return l
}

// Allow 检查该 IP 是否允许登录尝试
func (l *LoginRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.attempts[ip]
	if !ok {
		return true
	}

	// 窗口过期，重置
	if time.Since(a.firstAt) > windowPeriod {
		delete(l.attempts, ip)
		return true
	}

	// 未超限
	if a.count < maxAttempts {
		return true
	}

	// 已超限，拒绝
	return false
}

// RecordFailure 记录一次登录失败
func (l *LoginRateLimiter) RecordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.attempts[ip]
	if !ok {
		l.attempts[ip] = &loginAttempt{count: 1, firstAt: time.Now()}
		return
	}

	// 窗口过期，重置后重新计数
	if time.Since(a.firstAt) > windowPeriod {
		l.attempts[ip] = &loginAttempt{count: 1, firstAt: time.Now()}
		return
	}

	a.count++
}

// Reset 清除指定 IP 的失败记录（登录成功时调用）
func (l *LoginRateLimiter) Reset(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, ip)
}

// cleanup 定期清理过期条目，防止内存泄漏
func (l *LoginRateLimiter) cleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		for ip, a := range l.attempts {
			if time.Since(a.firstAt) > windowPeriod {
				delete(l.attempts, ip)
			}
		}
		l.mu.Unlock()
	}
}
