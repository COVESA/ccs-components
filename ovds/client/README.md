The CCS client is started as shown in the command line example below

$ ./ccs_client gen2-server-url ovds-server-url vin sleeptime<br>
An example could look like:<br>
$ ./client gen2_server.w3.org 192.168.8.108 GEO001 30

The CCS client will at startup look for the file "vsspathlist.json" in its own directory, if not found there it will check its parent directory for it. 
The latter directory is where the OVDS server will store the file after it has created it from the VSS tree file that it reads at startup. 

It will issue get requests on all leaf nodes in the tree, and for all nodes where a successful response is received, 
it will save the path in the vsspathlist.json file in its own directory. 
An existing path list file can be edited manually, and read by the CCS client after restart. 
The get requests are issued in the order in the list.
After the list is run through, the CCS client go into sleep the number of seconds given by sleeptime, then it starts with going through the list again.

