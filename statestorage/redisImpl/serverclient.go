// !!!!! redisInit must be executed before running serverclient !!!!

package main

import (
	"fmt"
	"encoding/json"
	"time"
	"github.com/go-redis/redis"
//	"github.com/go-redis/redis/v8"
)

type RedisDp struct {
	Val string
	Ts string
}

var serverClient *redis.Client

func redisGet(client *redis.Client, path string) (string, string) {
    dp, err := client.Get(path).Result()
    if err != nil {
        if err.Error() != "redis: nil" {
            fmt.Printf("Job failed. Error()=%s\n", err.Error())
            return "", ""
        } else {
            fmt.Printf("Data not found.\n")
            return "", ""
        }
    } else {
        fmt.Printf("Datapoint=%s\n", dp)
	var currentDp RedisDp
	err := json.Unmarshal([]byte(dp), &currentDp)
	if err != nil {
		fmt.Printf("Unmarshal failed for signal entry=%s, error=%s", string(dp), err)
		return "", ""
	} else {
	    fmt.Printf("Data: val=%s, ts=%s\n", currentDp.Val, currentDp.Ts)
	    return currentDp.Val, currentDp.Ts
	}
    }
}

func redisSet(client *redis.Client, path string, val string, ts string) int {
    dp := `{"val":"` + val + `", "ts":"` + ts + `"}`
    err := client.Set(path, dp, time.Duration(0)).Err()
    if err != nil {
        fmt.Printf("Job failed. Err=%s\n",err)
        return -1
    } else {
        fmt.Println("Datapoint=%s\n", dp)
        return 0
    }
}

func main() {
/*	serverClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})*/  // If TCP were to be used. Requires reconfig of redis.conf


    serverClient = redis.NewClient(&redis.Options{
        Network:  "unix",
        Addr:     "/var/tmp/vissv2/redisDB.sock",
        Password: "",
        DB:       1,
    })

    cPath := "Vehicle.Chassis.Test101"
    fmt.Printf("Path to current datapoint=%s\n", cPath)
    cVal, cTs := redisGet(serverClient, cPath)
    fmt.Printf("Server-redisGet(): Current value=%s, Current timestamp=%s\n", cVal, cTs)


    dPath := cPath + ".D"
    fmt.Printf("\n\nPath to desired datapoint=%s\n", dPath)

    dVal := "value2"
    dTs := "2022-02-22T13:37:59Z"
    fmt.Printf("Desired value=%s, desired timestamp=%s\n", dVal, dTs)

    status := redisSet(serverClient, dPath, dVal, dTs)
    if status != 0 {
        fmt.Printf("Server-redisSet() call failed.\n")
    } else {
        fmt.Printf("Server-redisSet() call succeeded.\n")
    }  
}
