package render

import (
	"codeberg.org/mjh/LibRate/cfg"
	"github.com/gofiber/storage/redis/v3"
)

func SetupCaching(conf *cfg.Config) *redis.Storage {
	return redis.New(redis.Config{
		Host:     conf.Redis.Host,
		Port:     conf.Redis.Port,
		Username: conf.Redis.Username,
		Password: conf.Redis.Password,
		Database: conf.Redis.PagesDB,
	})
}
