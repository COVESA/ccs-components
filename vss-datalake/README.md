The datalake server takes a database file name as command line input, as the examle shows below.

$ ./datalake_server path-to-file/vss-datalake.db

If the file does not exist, it creates an SQLite database with the provided name, and creates the tables:

CREATE TABLE "VIN_TIV" ( "vin_id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "vin" TEXT NOT NULL )
CREATE TABLE "TIV" ( "vin_id" INTEGER NOT NULL, "uuid" TEXT NOT NULL, "value" TEXT NOT NULL, FOREIGN KEY("vin_id") REFERENCES "VIN_TIV"("vin_id") )

The server supports the methods get/set/getmetadata. These methods are requested by the client via HTTP POST, with a JSON payload that specifies which method is requested, and the accompanying input parameters, see examples below (also found in datalake-request-examples.txt).


{"action":"get", "vin": "YV1DZ8256C2271234", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "from":"2020-01-01T02:59:43.492750Z", "to":"2020-03-31T02:59:43.492750Z"} // get specified period<br>
{"action":"get", "vin": "YV1DZ8256C2271234", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "from":"2020-01-09T02:59:43.492750Z"}  // get period from boundary up to latest value<br>
{"action":"get", "vin": "YV1DZ8256C2271234", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen"}  // get latest value


{"action":"set", "vin": "YV1DZ8256C2271234", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "value": "true", "timestamp":"2020-01-10T02:59:43.492Z"}


{"action":"getmetadata", "vin": "YV1DZ8256C2271234", "path":"Vehicle/Cabin/Door/", "depth": "2"}

When a request contains a VIN that has not been entered into the database before, a new table for this VIN is created:

CREATE TABLE TV_1 (`value` TEXT NOT NULL, `timestamp` TEXT NOT NULL, `uuid` TEXT)

where the index 1 in the table name is the value in the vin_id field of the entry for this VIN in the table VIN_TIV.

The server uses the cnative VSS tree manager from https://github.com/GENIVI/vss-tools, and the cnative instance of the VSS tree found there. 
