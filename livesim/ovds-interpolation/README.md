# OVDS interpolator
If the data in the OVDS has been processed with the curve logging algorithm, then there is not a stream of data with a contstant sample period. 
To get a stream of data with an equidistant time period between samples as output from the live simulator, 
the OVDS interpolator creates a new OVDS, filling in the missing samples through interpolation. 

The reason to get an equidistant stream is that it can be synched with an equidistant capture period by the VISSv2 server.
Another solution would be to update the VISSv2 server to be able to asynchronously capture new data points as they emerge, which is in the pipe, 
but will not be done now. 

The OVDS interpolator is started with the following command:

$ ./ovds_interpolator old-ovds-server-ipadr old-vin new-ovds-server-ipadr new-vin

where 
 - old-ovds-server-ipadr is the IP address of the OVDS server from which the data to be interpolated is read,
 - new-ovds-server-ipadr is the IP address of the OVDS server to which interpolated data is written,
 - old-vin is the VIN used in the existing OVDS,
 - new-vin is the VIN to be used in the new OVDS.
 
 The two OVDS servers must be started before the OVDS interpolator, and they cannot be started having the same IP address.
 


