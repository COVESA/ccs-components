To build the Open Vehicle Data Set (OVDS) server:
$ go build

The OVDS server then takes a database file name as command line input, as the example shows below.

$ ./server db-file-name

There is one exemption, and that is when it is used together with the livesim vehicle data simulator, it should then be started as below instead.
$ ./ovds_server db-file-name livesim


If the database file does not exist, it creates an SQLite database with the provided name, and creates the tables:

CREATE TABLE "VIN_TIV" ( "vin_id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "vin" TEXT NOT NULL )<br>
CREATE TABLE "TIV" ( "vin_id" INTEGER NOT NULL, "path" TEXT NOT NULL, "value" TEXT NOT NULL, FOREIGN KEY("vin_id") REFERENCES "VIN_TIV"("vin_id") )

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
The server supports the methods get/set. These methods are requested by the client via HTTP POST, with a JSON payload that specifies which method is requested, and the accompanying input parameters, see examples below.


{"action":"get", "vin": "YV1DZ8256C2271234", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "from":"2020-01-01T02:59:43.492750Z", "to":"2020-03-31T02:59:43.492750Z"} // get specified period<br>
{"action":"get", "vin": "YV1DZ8256C2271234", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "from":"2020-01-09T02:59:43.492750Z"}  // get period from boundary up to latest value<br>
{"action":"get", "vin": "YV1DZ8256C2271234", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "from":"2020-01-09T02:59:43.492750Z", "maxsamples":"5"}  // get period from boundary up to latest value, not more than 5 samples<br>
{"action":"get", "vin": "YV1DZ8256C2271234", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen"}  // get latest value


For time invariant signals, i. e. signals of node type ATTRIBUTE, the request must not include a timestamp.
{"action":"set", "vin": "YV1DZ8256C2271234", "path":"Vehicle.VehicleIdentification.Year", "value": "1957"}<br>

For time variant signals, i. e. signals of node type SENSOR or ACTUATOR, the request must include a timestamp.
{"action":"set", "vin": "YV1DZ8256C2271234", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "value": "true", "timestamp":"2020-01-10T02:59:43.492Z"}<br>

Set requests to the same VIN and path are ignored if there already is an entry in the DB for the provided timestamp. 

When a set request contains a VIN that has not been entered into the database before, a new table for this VIN is created:

CREATE TABLE TV_1 (`value` TEXT NOT NULL, `timestamp` TEXT NOT NULL, `path` TEXT)

where the index 1 in the example table name above is the value in the vin_id field of the entry for this VIN in the table VIN_TIV.

