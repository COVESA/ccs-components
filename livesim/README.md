# Live simulator
The live simulator reads the OVDS database that the OVDS server manages, and writes the data into the statestorage database in a temporal place set by the timestamps of the data. 
The live simulator is built and started from the command lines as follows:

To build the executable:<br>
$ go build<br>
To learn about the input parameters to the executable, start it first without parameters, to show the following help text:<br>
$ ./livesim<br>
[-o|--ovdsUrl] is required<br>
usage: print [-h|--help] -o|--ovdsUrl "<value>" -v|--vin "<value>" -p|--dbPath<br>
             "<value>" [-i|--dbImpl (sqlite|redis)]<br>

             Prints provided string to stdout<br>

Arguments:<br>

  -h  --help     Print help information<br>
  -o  --ovdsUrl  IP/url to OVDS server<br>
  -v  --vin      VIN from OVDS DB<br>
  -p  --dbPath   Path and name of state storage SQLite DB file<br>
  -i  --dbImpl   Database impl must be either sqlite or redis. Default: sqlite<br>

The OVDS server URL/IP address, and the VIN, are mandatory.<br>
If the SQLite state storage implementation is selected (default), then the path and name of the SQLite DB file must be provided.<br>
If the Redis state storage implementation is selected, then the Redis DB must be initiated, see the 
<a href="https://github.com/w3c/automotive-viss2/tree/master/server/service_mgr">VISSv2 service manager README</a>.

If the live simulator is used together with the VISSv2 server, and that server is configured to run using an SQLite state storage implementation, 
then the VISSv2 server retrieves data from a database file named statestorage.db if it finds it in its deployment directory,<br>
see the <a href="https://github.com/w3c/automotive-viss2/tree/master/server/service_mgr">VISSv2 service manager README</a>.

So if livesim writes to the state storage used by the VISSv2 server, a VISSv2 client can access simulated "live" data coming from recorded data in the OVDS database that livesim reads from.
The OVDS server that livesim reads from must be started with the extra command parameter "livesim" to listen to listen for its requests.
The file sawtooth_trip.db contains Speed, Lat and Long signals, where the signal values create a sawtooth curve, which can be used for test purposes. 
The OVDS server is then started as below.<br>
$ ./server sawtooth_trip.db livesim

