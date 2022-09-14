package store

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

func InitRedisDB(ctx context.Context) *redis.Client {
	dsn := fmt.Sprintf("%v:%v", viper.GetString("redis.host"), viper.GetString("redis.port"))

	rdb := redis.NewClient(&redis.Options{
		Addr:     dsn,
		Password: viper.GetString("redis.password"),
		DB:       0, // use default DB
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	return rdb
}
