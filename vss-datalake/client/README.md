The CCS client is started as shown in the command line example below

$ ./ccs-client nodelist-filename gen2-server-url datalake-server-url vss-tree-filename vin sleeptime
An example could look like:<br>
$ ./client nodelist.txt gen2_server.w3.org 192.168.8.108 ../server/vss_rel_2.0.0-alpha+006.cnative GEO001 30

If the nodelist-filename points to a non-existent file, the CCS-client will create one from a local VSS tree file.
It will issue get requests on all leaf nodes in the tree, and for all nodes where a successful response is received, 
it will save the path in the nodelist file. 
An existing node list file can be edited manually, and read by the CCS client after restart. 
The get requests are issued in the order in the list, after the list is run through, 
the CCS client go into sleep the number of secons given by sleeptime, then it starts with going through the list again.

