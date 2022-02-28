// !!!!! redisInit must be executed before running redisFlush !!!!

package main

import (
	"fmt"
	"github.com/go-redis/redis"
//	"github.com/go-redis/redis/v8"
)

func main() {
/*	serverClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})*/  // If TCP were to be used. Requires reconfig of redis.conf


    client := redis.NewClient(&redis.Options{
        Network:  "unix",
        Addr:     "/var/tmp/vissv2/redisDB.sock",
        Password: "",
        DB:       1,
    })

    status := client.FlushDB()
    fmt.Printf("Redis flush status=%s\n", status)
}
