# State storage concept

As shown in the figure below the state storage is located in between the vehicle server (on its northbound side) and one or more feeders (on its southbound side). 
![State storage architecture](state_storage_swa.jpg?raw=true)<br>
*Fig 1. State storage architecture overview<br>
For reference, see also the <a href="https://at.projects.genivi.org/wiki/display/MIG/CCS+Proof-Of-Concept+-+Work+Breakdown+Structure">COVESA CVII proof-of-concept project architecture</a>:

The state storage provides the following features:
### 1. Holds a copy of the latest data point for all the vehicle's VSS defined signals
The statestorage provides storage of the "state of the vehicle", i. e. it holds a copy of the latest value and timestamp (=data point), of all vehicle signals that are defined in the VSS tree for the vehicle.<br>

### 2. Allows for extension of new feeders
In an after-market scenarion where the vehicle gets new functionality installed that results in new signals to be communicated, 
a new feeder handling these signals can be added. The VSS tree must then also be extended with the signals, e. g. added to the private branch.

### 3. Handles the access synchronization of the state storage memory
The vehicle server and the feeders are in relation to each other asynchronous agents, which means they try to access the state storage in an asychronous manner. 
It is the responsibility of the state storage to handle this in a way so that data written to it does not become corrupted. 

### 4. Supports the "state transition" of vehicle actuators
When a vehicle actuator is set to a new value it might take some time for it to change from its current state to the new state represented by the new value. 
An example could be a window that is currently closed, but is set to become fully open.<br>
To enable the client (to the vehicle server) that requested this state change to follow up on its execution, the state storage shall provide two data points for every actuator signal, 
one "desired datapoint", and one "current datapoint". A client request for a new value is then written to the desired datapoint by the server, 
and the subsequent state changes over a period of time is written to the current datapoint by the feeder. The client can hence follow the state change by reading the actuator value, 
which is then obtained by the server from the current datapoint.<br>

### State storage implementations
The state storage is currently available in two different implementations, found in respective subdirectory:<br>
1. Based on an SQLite database
2. Based on a Redis database

The following method signatures are meant to "inspire" the implementations to allow for simple exchange between them(here expressed in Go syntax):<br>

func GetDataPoint(path string) (val string, ts string)<br>
func SetDataPoint(path string, val string, ts string) string<br>

The GetDataPoint method takes a VSS path as input, and returns the value and timestamp of that datapoint.<br>
The SetDataPoint method takes a VSS path, and the associated value and timestamp as input, and returns a status value.<br>

The fact that those methods shall be able to access either the desired or current datapoint must be resolved by the respective implementations.<br>

The state storage implementations typically also requires an initialisation request at system startup, 
a method signature for this is likely to differ significantly between implementations, so there is no common proposal shown here, 
but it is expected that an implementation example is shown in the respective implementation subdirectories. 
The initialisation might require root privileges. <br>

The respective implementation directories shall contain sufficient information so that a developer of a system including a vehicle server, a state storage, 
and one or more feeders shall be able to implement the interactions between these components.<br>
An example of a state storage integration into such a subsystem can be found at<br>
<a href="https://github.com/w3c/automotive-viss2">W3C CVII VISSv2 server implementation</a>

## Extending the state storage concept to storage of time series data
An initiative in this direction is looked upon in this project:
<a href="https://github.com/slawr/vss-otaku">VSS OTAKU</a>

