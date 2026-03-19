package config

import "time"

// default const
const (
	DefaultConnMaxLifeTime time.Duration = 1 * time.Hour
	DefaultConnMaxIdleTime time.Duration = 15 * time.Minute
	DefaultDBPingInterval  time.Duration = 1 * time.Second
	DefaultDBRetryAttempts int           = 3

	// DefaultRedisDialTimeout default dial timeout
	DefaultRedisDialTimeout = time.Second * 5
	// DefaultRedisWriteTimeout default write timeout
	DefaultRedisWriteTimeout = time.Second * 2
	// DefaultRedisReadTimeout default read timeout
	DefaultRedisReadTimeout = time.Second * 2
	// DefaultRedisMaxIdleConn default max idle conn
	DefaultRedisMaxIdleConn = 20
	// DefaultRedisMaxActiveConn default max active conn
	DefaultRedisMaxActiveConn = 50
	// DefaultRedisIdleTimeout default idle timeout
	DefaultRedisIdleTimeout = 240 * time.Second
	// DefaultRedisMaxConnLifetime default max conn lifetime
	DefaultRedisMaxConnLifetime = time.Minute

	// DefaultCacheTTL default ttl for cache
	DefaultCacheTTL = 5 * time.Minute
)
