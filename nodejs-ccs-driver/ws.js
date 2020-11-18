var w3cdecompress = require("./w3cdecompress.js")
var vsspathlist = require("./vsspathlist.js")
var WebSocketClient = require('websocket').client;
const { uuidlist } = require("./vsspathlist.js");
var db = require("./db.js");
var w3cclient = new WebSocketClient();
// w3cclient.binaryType = "blob";

var wsconnection = null

module.exports = {

    get: function(index) {
        if (wsconnection && wsconnection.connected) {
            wsconnection.send('{"action":"get", "path": "' + vsspathlist.uuidlist['LeafPaths'][index] + '", "requestId": "' + index + '"}')
        }
    },
    w3cInit: function(ip, port) {
        connstr = "ws://" + ip + ":" + port + "/"
        console.log('Connection String : ' + connstr);
        w3cclient.connect(connstr, "gen2c");
    }
}

w3cclient.on('connectFailed', function (error) {
    console.log('Connect Error: ' + error.toString());
});

w3cclient.on('connect', function (connection) {
    console.log('WebSocket Client Connected');
    wsconnection = connection
    connection.on('error', function (error) {
        console.log("Connection Error: " + error.toString());
        wsconnection = null
    });
    connection.on('close', function () {
        console.log(connection.protocol + ' Connection Closed');
        wsconnection = null
    });
    connection.on('message', function (msg) {
        console.log("Processing response :" + Object.keys(msg));
        if (msg.type == "binary") {
            // reader = new FileReader();
            // reader.onload = () => {
                var localdata = Buffer.from(msg.binaryData).toJSON().data
                console.log(localdata)
                var decStr = w3cdecompress.decompressMessage(localdata)
                // var decStr = w3cdecompress.decompressMessage(reader.result)
                console.log(decStr)
                decObj = JSON.parse(decStr)
                console.log(typeof decObj.value)
                //write to DB
                if (typeof decObj.value == "string"){
                    db.addKVP(decObj)
                } else {
                    db.addKAP(decObj)
                }
                var prettyObj = JSON.stringify(decObj, null, 2)
                ratio = Math.round((decStr.length * 100) / localdata.length);
                console.log("Decompressed : " + decStr.length + "\n" + JSON.stringify(decObj) + "\n");
                console.log("Compressed   : " + localdata.length + "\n");
                console.log("Compression rate : " + ratio + "%\n\n");
            // };
            // fs.write("./temp.txt", Buffer.from(msg.binaryData))
            // reader.readAsBinaryString("./temp.txt");
        }
    });
});
