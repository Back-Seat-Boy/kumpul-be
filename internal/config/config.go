package config

import (
	"fmt"
	"time"

	"github.com/kumparan/go-utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func InitConfig() {
	viper.AddConfigPath(".")
	viper.AddConfigPath("./..")
	viper.SetConfigName("config")

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Warningf("%v", err)
	}
	log.Info("Using config file: ", viper.ConfigFileUsed())
}

func Port() string {
	if !viper.IsSet("port") {
		return "8080"
	}
	return viper.GetString("port")
}

func Environment() string {
	if !viper.IsSet("env") {
		return "development"
	}
	return viper.GetString("env")
}

func GrpcPort() string {
	if !viper.IsSet("grpc.port") {
		return "9090"
	}
	return viper.GetString("grpc.port")
}

func DBHost() string {
	return viper.GetString("database.host")
}

func DBDatabase() string {
	return viper.GetString("database.database")
}

func DBUser() string {
	return viper.GetString("database.username")
}

func DBPassword() string {
	return viper.GetString("database.password")
}

func DBDSN() string {
	ssl := "require"
	if Environment() != "production" {
		ssl = "disable"
	}
	return fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=%s", DBUser(), DBPassword(), DBHost(), DBDatabase(), ssl)
}

func MaxIdleConns() int {
	if !viper.IsSet("database.maxIdleConns") {
		return 3
	}
	return viper.GetInt("database.maxIdleConns")
}

func MaxOpenConns() int {
	if !viper.IsSet("database.maxOpenConns") {
		return 15
	}
	return viper.GetInt("database.maxOpenConns")
}

func ConnMaxLifeTime() time.Duration {
	return utils.ValueOrDefault[time.Duration](viper.GetDuration("database.connMaxLifeTime"), DefaultConnMaxLifeTime)

}

func ConnMaxIdleTime() time.Duration {
	return utils.ValueOrDefault[time.Duration](viper.GetDuration("database.connMaxIdleTime"), DefaultConnMaxIdleTime)

}

func DBPingInterval() time.Duration {
	return utils.ValueOrDefault[time.Duration](viper.GetDuration("database.pingInterval"), DefaultDBPingInterval)
}

func DBRetryAttempts() int {
	if viper.GetInt("database.retryAttempts") > 0 {
		return viper.GetInt("database.retryAttempts")
	}
	return DefaultDBRetryAttempts
}

func RedisDialTimeout() time.Duration {
	if viper.GetDuration("redis.dial_timeout") < time.Second {
		return DefaultRedisDialTimeout
	}
	return viper.GetDuration("redis.dial_timeout")
}

func RedisWriteTimeout() time.Duration {
	if viper.GetDuration("redis.write_timeout") < time.Second {
		return DefaultRedisWriteTimeout
	}
	return viper.GetDuration("redis.write_timeout")
}

func RedisReadTimeout() time.Duration {
	if viper.GetDuration("redis.read_timeout") < time.Second {
		return DefaultRedisReadTimeout
	}
	return viper.GetDuration("redis.read_timeout")
}

func RedisMaxIdleConn() int {
	if viper.GetInt("redis.max_idle_conn") <= 0 {
		return DefaultRedisMaxIdleConn
	}
	return viper.GetInt("redis.max_idle_conn")
}

func RedisMaxActiveConn() int {
	if viper.GetInt("redis.max_active_conn") <= 0 {
		return DefaultRedisMaxActiveConn
	}
	return viper.GetInt("redis.max_active_conn")
}

func RedisCacheHost() string {
	return viper.GetString("redis.cache_host")
}

func RedisLockHost() string {
	return viper.GetString("redis.lock_host")
}

func RedisIdleTimeout() time.Duration {
	return utils.ValueOrDefault[time.Duration](viper.GetDuration("redis.idle_timeout"), DefaultRedisIdleTimeout)
}

func RedisMaxConnLifetime() time.Duration {
	return utils.ValueOrDefault[time.Duration](viper.GetDuration("redis.max_conn_lifetime"), DefaultRedisMaxConnLifetime)
}

func DisableCaching() bool {
	return viper.GetBool("disable_caching")
}

func CacheTTL() time.Duration {
	if !viper.IsSet("cache_ttl") {
		return DefaultCacheTTL
	}

	return viper.GetDuration("cache_ttl")
}

func LogLevel() string {
	return viper.GetString("log_level")
}

func GoogleClientID() string {
	return viper.GetString("google.client_id")
}

func GoogleClientSecret() string {
	return viper.GetString("google.client_secret")
}

func GoogleRedirectURL() string {
	return viper.GetString("google.redirect_url")
}

func CORSAllowedOrigins() []string {
	return viper.GetStringSlice("cors.allowed_origins")
}

func CORSAllowCredentials() bool {
	return viper.GetBool("cors.allow_credentials")
}

func FrontendURL() string {
	return viper.GetString("frontend_url")
}

func CloudinaryCloudName() string {
	return viper.GetString("cloudinary.cloud_name")
}

func CloudinaryAPIKey() string {
	return viper.GetString("cloudinary.api_key")
}

func CloudinaryAPISecret() string {
	return viper.GetString("cloudinary.api_secret")
}

func CloudinaryUploadFolder() string {
	return viper.GetString("cloudinary.upload_folder")
}
