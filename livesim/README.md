# Live simulator
The live simulator reads the OVDS database that the OVDS server manages, and writes the data into the statestorage database in a temporal pace set by the timestamps of the data. 
The live simulator is started from the command line as follows:

$ ./livesim ULF001 url-to-ovds-server ../statestorage/statestorage.db 

where 
- ULF001 is the VIN associated with the data from OVDS
- url-to-ovds-server is the IP address of the OVDS server
- ../statestorage/statestorage.db is the path to the statestorage database.

The service manager component of the Gen server retrieves data from the statestorage database that it finds in its deployment directory. 
This enables a Gen2 client to read simulated "live" data coming from recorded data in an OVDS database.

