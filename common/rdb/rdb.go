package rdb

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var RDB = redis.NewClient(&redis.Options{
	Addr:     "127.0.0.1:6378",
	Password: "", // no password set
	DB:       0,  // use default DB
})

var Ctx = context.Background()
