The Open Vehicle Data Set (OVDS) server takes a database file name and the name of a file containing a VSS tree as command line input, as the examle shows below.

$ ./ovds_server db-file-name vss-tree-filename

The VSS tree must have the binary format. 
If the database file does not exist, it creates an SQLite database with the provided name, and creates the tables:

CREATE TABLE "VIN_TIV" ( "vin_id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "vin" TEXT NOT NULL )<br>
CREATE TABLE "TIV" ( "vin_id" INTEGER NOT NULL, "path" TEXT NOT NULL, "value" TEXT NOT NULL, FOREIGN KEY("vin_id") REFERENCES "VIN_TIV"("vin_id") )

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

When a request contains a VIN that has not been entered into the database before, a new table for this VIN is created:

CREATE TABLE TV_1 (`value` TEXT NOT NULL, `timestamp` TEXT NOT NULL, `path` TEXT)

where the index 1 in the example table name above is the value in the vin_id field of the entry for this VIN in the table VIN_TIV.

