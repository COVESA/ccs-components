## Redis state storage implementation

### Install Redis on Ubuntu
$ sudo apt-get install redis-server

### Configure Redis
Redis clients can connect over either a TCP socket or a Unix Domain socket. TCP is default. 
For performance and security reasons, this implementation will use the UDS alternative, which means that the redis.conf file must be updated.
On Ubuntu this file is found at /etc/redis/redis.conf.
To update the file, open it in an editor, and add the following rows. Preferrably close to the lines where it is described in the file, search for "unixsocket".<br>
unixsocket /var/tmp/vissv2/redisDB.sock<br>
unixsocketperm 777<br>

Then comment the line<br>
port 6379<br>
and add the line<br>
port 0<br>
for Redis to ignore the port number as it is not used in UDS.

Save the file. 

### Current vs Desired datapoints in the state storage
The implementation of the Current and Desired datapoints has the following design.
Datapoints are stored in Redis as a JSON string containing both the value and the timestamp, with the path being the key. 
The key-value pair stored in one entry only contains one datapoint. 
To distinguish between Current and Desired datapoints, the key is modified. 
For Current datapoints the VSS path is used as is, but for Desired datapoints the path is appended with ".D".
For example, the path "Vehicle.Abc" will access the Current datapoint of this path, while the path "Vehicle.Abc.D" will access the Desired datapoint of the path "Vehicle.Abc".
It is the caller of the GetDataPoint(), SetDataPoint() methods to append this for access to Desired datapoints. 

The Go files
- redisInit.go
- serverclient.go
- feederclient.go

are included to allow some initial testing before implementing Redis as state storage in a "technology stack" project. 
After installing and configuring Redis as described above, building and running the redisInit.go will start the Redis daemon. 
The three Go files need to be stored in separate directories before the<br>
$ go build<br>
command is applied, or else the Go compiler will complain.<br>
Then the serverclient and feederclient can be built, and when running them they will read and write datapoints in Redis, transferring the datapoints from one to the other. 

## Implementation of the getDataPoint() method
```
import (
	"encoding/json"
	"github.com/go-redis/redis"
)

type RedisDp struct {
	Val string
	Ts string
}

func GetDataPoint(path string) (val string, ts string) {
    dp, err := client.Get(path).Result()
    if err != nil {
        if err.Error() == "redis: nil" {
            return "", "" // Data not found.
        } else {
            return "", ""  // Other error.
        }
    } else {
	var currentDp RedisDp
	err := json.Unmarshal([]byte(dp), &currentDp)
	if err != nil {
		return "", ""  // Unmarshal failed.
	} else {
	    return currentDp.Val, currentDp.Ts
	}
    }
}
```
## Implementation of the method setDataPoint() method
```
import (
	"encoding/json"
	"time"
	"github.com/go-redis/redis"
)

func SetDataPoint(path string, val string, ts string) string {
    dp := `{"val":"` + val + `", "ts":"` + ts + `"}`
    err := client.Set(path, dp, time.Duration(0)).Err()  // duration = 0 => keep indefinitively
    if err != nil {
        return -1 // Failure.
    } else {
        return 0 // Success.
    }
}
```

## Implementation of the initStateStorage() method
```
import (
	"os/exec"
	"github.com/go-redis/redis"
)

func initStateStorage(udsPath string) bool {
    client := redis.NewClient(&redis.Options{
        Network:  "unix",
        Addr:     udsPath,
        Password: "",
        DB:       1,
    })
    err := client.Ping().Err()
    if err != nil {
        out, err := exec.Command("redis-server", "/etc/redis/redis.conf").Output()
        if err != nil {
            false
        } else {
            true // Redis server is started.
        }
    } else {
            true // Redis server was already running.
    }
}
```

## Extending the state storage concept to storage of time series data
An initiative in this direction is looked upon in this project:
<a href="https://github.com/slawr/vss-otaku">VSS OTAKU</a>
