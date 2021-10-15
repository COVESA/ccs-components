To build the CCS client
$ go build

## ccs-w3c-client usage
```
usage: print [-h|--help] -w|--vissv2Url "<value>" [-o|--ovdsUrl "<value>"]
             [-v|--vin "<value>"] [-p|--iterationPeriod <integer>]
             [-a|--accessMode (get|subscribe)]

             Prints provided string to stdout

Arguments:

  -h  --help             Print help information
  -w  --vissv2Url        IP/URL to W3C VISS v2 server (REQUIRED)
  -o  --ovdsUrl          IP/url to OVDS server
  -v  --vin              VIN number. Default: ULF001
  -p  --iterationPeriod  estimated time in msec for one roundtrip. Default: 15
  -a  --accessMode       CCS client access-mode must be either get or
                         subscribe.. Default: get
```
An example could look like:<br>
$ ./client gen2_server.w3.org 192.168.8.108 GEO001 30 subscribe

The CCS client will at startup look for the file "vsspathlist.json" in its own directory, if not found there it will check its parent directory for it. 
The latter directory is where the OVDS server will store the file after it has created it from the VSS tree file that it reads at startup. 

Depending on the access-mode, which must be either "get" or "subscribe", it will access all leaf nodes in the vsspathlist.json file from the VISSv2 server, and via the OVDS server write it to the OVDS database.
For the get access mode it tries to spread out the requests of one complete run through of all paths to be close to the iteration-period in seconds. 
For the subscribe access mode it spreads out the subscribe requests over the iteration period. The subscriptions are time-based with a period of 3 seconds.

For the get access mode, if the iteration period is to small, and the number of paths high, it may not be possible to issue all the requests within that period, which leads to that the period time is extended. One cycle of read/write took in the WiFI home network where it was tested about 30 msec, so iteration-periods less than that times the number of paths may not be achieved. 

The JSON path list file can be edited manually, if the complete path list in the vsspathlist.json that the server creates is unnecessarily big. 
The get requests are issued in the order as shown in the path list file.
After the list is run through, the CCS client immediately starts a new iteration.

The HTML test client can be used for manual testing of the OVDS server. When started the URL/IP-address to the OVDS server must be entered, 
"ovdsserver" must be entered in Path field (without quote signs), and then the payload is entered in the Value field, see payload examples in the README in the ovds directory. 
For a write to the node Vehicle/Acceleration/Longitudinal, the payload could look like:
{"action":"set", "vin": "YV1DZ8256C2271234", "path":"Vehicle/Acceleration/Longitudinal", "value": "0.123", "timestamp":"2020-01-10T02:59:43.492Z"}
