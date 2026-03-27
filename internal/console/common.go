package console

import (
	"context"

	"github.com/Back-Seat-Boy/kumpul-be/internal/config"
	"github.com/kumparan/cacher"
	"github.com/kumparan/go-connect"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func initializeCockroachConn() {
	connect.InitializeCockroachConn(config.DBDSN(), &connect.CockroachDBConnectionOptions{
		PingInterval:    config.DBPingInterval(),
		RetryAttempts:   5,
		MaxIdleConns:    config.MaxIdleConns(),
		MaxOpenConns:    config.MaxOpenConns(),
		ConnMaxLifetime: config.ConnMaxLifeTime(),
		LogLevel:        config.LogLevel(),
	})
}

func newCacheKeeper() cacher.Keeper {
	cacheKeeper := cacher.NewKeeper()

	if !config.DisableCaching() {
		redisOpts := &connect.RedisConnectionPoolOptions{
			DialTimeout:     config.RedisDialTimeout(),
			ReadTimeout:     config.RedisReadTimeout(),
			WriteTimeout:    config.RedisWriteTimeout(),
			IdleCount:       config.RedisMaxIdleConn(),
			PoolSize:        config.RedisMaxActiveConn(),
			IdleTimeout:     config.RedisIdleTimeout(),
			MaxConnLifetime: config.RedisMaxConnLifetime(),
		}

		redisConn, err := connect.NewRedigoRedisConnectionPool("redis://"+config.RedisCacheHost(), redisOpts)
		continueOrFatal(err)

		redisLockConn, err := connect.NewRedigoRedisConnectionPool("redis://"+config.RedisLockHost(), redisOpts)
		continueOrFatal(err)

		cacheKeeper.SetConnectionPool(redisConn)
		cacheKeeper.SetLockConnectionPool(redisLockConn)
		cacheKeeper.SetDefaultTTL(config.CacheTTL())
	}

	cacheKeeper.SetDisableCaching(config.DisableCaching())

	return cacheKeeper
}

func gracefulShutdownHTTPServer(e *echo.Echo) {
	if e == nil {
		return
	}

	if err := e.Shutdown(context.Background()); err != nil {
		log.Errorf("error shutting down http server: %v", err)
	}
}

func continueOrFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
