// redisInit must be started with root permission (sudo ./redisInit)
// !!!!! redisInit must be executed before running serverclient or feederclient !!!!

package main

import (
	"fmt"
	"os/exec"
	"github.com/go-redis/redis"
//	"github.com/go-redis/redis/v8"
)


func main() {
    client := redis.NewClient(&redis.Options{
        Network:  "unix",
        Addr:     "/var/tmp/vissv2/redisDB.sock",
        Password: "",
        DB:       1,
    })
    err := client.Ping().Err()
    if err != nil {
        out, err := exec.Command("redis-server", "/etc/redis/redis.conf").Output()
        if err != nil {
            fmt.Printf("Starting redis server failed. Err=%s\n", err)
        } else {
            fmt.Printf("Redis server started.%s\n", out)
        }
    } else {
            fmt.Printf("Redis server is running.\n")
    }

}
