// !!!!! redisInit must be executed before running feederclient !!!!

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

var feederClient *redis.Client


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
        fmt.Printf("Job done.\nDatapoint=%s\n", dp)
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

    feederClient = redis.NewClient(&redis.Options{
        Network:  "unix",
        Addr:     "/var/tmp/vissv2/redisDB.sock",
        Password: "",
        DB:       1,
    })

    cPath := "Vehicle.Chassis.Test101"
    fmt.Printf("Path to current datapoint=%s\n", cPath)

    Cvalue := "value1"
    Cts := "2022-02-21T13:37:00Z"
    fmt.Printf("Current value=%s, current timestamp=%s\n", Cvalue, Cts)

    status := redisSet(feederClient, cPath, Cvalue, Cts)
    if status != 0 {
        fmt.Printf("Feeder-redisSet() call failed.\n")
    } else {
        fmt.Printf("Feeder-redisSet() call succeeded.\n") 
    }   


    dPath := cPath + ".D"
    fmt.Printf("\n\nPath to desired datapoint=%s\n", dPath)
    dVal, dTs := redisGet(feederClient, dPath)
    fmt.Printf("Feeder-redisGet() call succeeded.\nDesired value=%s, Desired timestamp=%s\n", dVal, dTs)
}
