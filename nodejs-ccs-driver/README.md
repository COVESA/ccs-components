# Setup Guide

## Pre-requisites

### Installing Grafana in Ubuntu 

`sudo apt-get install -y adduser libfontconfig1`\
`wget https://dl.grafana.com/oss/release/grafana_7.2.0_amd64.deb`\
`sudo dpkg -i grafana_7.2.0_amd64.deb`

### Setup grafana

Login to localhost:3000 in browser
setup admin passwords.

### Install MySQL 
`sudo apt install mysql-server`\
`sudo mysql_secure_installation`
  -- configure security options (password policy etc)

### Create admin user

`sudo mysql -u root\
CREATE USER 'admin'@'localhost' IDENTIFIED WITH mysql_native_password BY 'admin123';\
GRANT ALL PRIVILEGES ON *.* TO 'admin'@'localhost';\
FLUSH PRIVILEGES\
exit`

### Login using new admin user
`mysql -u admin -p`
`password: admin123`

### Create new database 'vssdatalake'
`CREATE DATABASE vssdatalake`

### Get node packages
`cd node-ccs-driver`
`npm install`

### Run the **nodejs-ccs-driver**
`npm start`

# nodejs-ccs-driver
## *A quick and dirty implementation of a CCS client that fills up the VSS Data Lake (MySQL DB)*

## Reuses most of the JS code written in the following location
[JS Client](https://github.com/MEAE-GOT/W3C_VehicleSignalInterfaceImpl/tree/master/client/client-1.0/Javascript)


## The **nodejs-ccs-driver** does the following steps
    1. Tries to create a DB called *vssdatalake*
    2. Tries to create two tables called *kap* and *kvp* (Key Array Pair, Key Value Pair)
    3. Connects to Gen2 server instance.  (hardcoded to localhost:8080 for demo)
    4. Sends a request every second sequentially to Gen2 server.
    5. Decompress the response and write to DB.

## Grafana instance reads off mysql DB and shows visualizations.
