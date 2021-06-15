# Live simulator
The live simulator reads the OVDS database that the OVDS server manages, and writes the data into the statestorage database in a temporal pace set by the timestamps of the data. 
The live simulator is built and started from the command lines as follows:

$ go build
$ ./livesim ULF url-to-ovds-server ../statestorage/statestorage.db 

where 
- ULF is the VIN associated with the data from OVDS
- url-to-ovds-server is the IP address of the OVDS server
- ../statestorage/statestorage.db is the path to the statestorage database.

The VISSv2 server retrieves data from a database file named statestorage.db if it finds it in its deployment directory,<br>
see the <a href="https://github.com/MEAE-GOT/W3C_VehicleSignalInterfaceImpl/tree/master/server/service_mgr">VISSv2 service manager</a>

So if livesim writes to this database, a VISSv2 client can access simulated "live" data coming from recorded data in the OVDS database that livesim reads from.
The OVDS server that livesim reads from must be started with the extra command parameter "livesim" to listen to listen for its requests.
The file sawtooth_trip.db contains Speed, Lat and Long signals, where the signal values create a sawtooth curve, which can be used for test purposes. 
The OVDS server is then started as below.<br>
$ ./server sawtooth_trip.db livesim

