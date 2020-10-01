#nodejs-ccs-driver

1. A quick and dirty implementation of a CCS client that fills up the VSS Data Lake (DB)

2. Reuses most of the JS code written in the following location
[JS Client](https://github.com/MEAE-GOT/W3C_VehicleSignalInterfaceImpl/tree/master/client/client-1.0/Javascript)

3. Pre-requisites

- Install mysql - standard installation with a user and password
    - This example uses "admin" and "admin123"
    - Port number 3306

- Get node packages
    `code`
    npm install
    `code`

- Run the **nodejs-ccs-driver**
    `code`
    node index.js
    `code`

4. The **nodejs-ccs-driver** does the following steps
    a. Tries to create a DB called *vssdatalake*
    b. Tries to create two tables called *kap* and *kvp* (Key Array Pair, Key Value Pair)
    c. Connects to Gen2 server instance.  (hardcoded to localhost:8080 for demo)
    d. Sends a request every second sequentially to Gen2 server.
    e. Decompress the response and write to DB.

5. Grafana instance reads off mysql DB and shows visualizations.