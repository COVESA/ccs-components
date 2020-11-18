var mysql = require('mysql');
var vsspathlist = require("./vsspathlist.js")
var con = null

module.exports = {
    initDb: function(ip, user, pass, dbname) {
        con = mysql.createConnection({
            host: ip,
            user: user,
            password: pass,
            database: dbname
        });
        con.connect(function (err) {
            if (err) throw err;
            console.log("Connected!");
            con.query("CREATE DATABASE IF NOT EXISTS vssdatalake", function (err, result) {
                if (err) throw err;
                console.log("Database created");
                var sql = "CREATE TABLE  IF NOT EXISTS kvp (uuid VARCHAR(255) NOT NULL, value INTEGER NOT NULL, ts TIMESTAMP default CURRENT_TIMESTAMP)";
                con.query(sql, function (err, result) {
                    if (err) throw err;
                    console.log("Key Value table created");
                });
                sql = "CREATE TABLE  IF NOT EXISTS kap (uuid VARCHAR(255) NOT NULL, value INTEGER NOT NULL, ts TIMESTAMP default CURRENT_TIMESTAMP)";
                con.query(sql, function (err, result) {
                    if (err) throw err;
                    console.log("Key Array table created");
                });
            });
        });
    },
    addKVP: function(obj) {
        console.log("Testing " + obj.timestamp)
        var sql = "INSERT INTO kvp (uuid, value) VALUES ('"
                    + vsspathlist.uuidlist['LeafPaths'][obj.requestId] 
                    + "','" + parseInt(obj.value)  + "')"
                    // + "',UNIX_TIMESTAMP(STR_TO_DATE('" + obj.timestamp + "', '%Y-%m-%dT%H:%M:%sZ'))"
        console.log(sql)
        con.query(sql, function (err, result) {
            if (err) throw err;
            console.log("1 record inserted")
        });
    },
    addKAP: function(obj) {
        console.log("Testing " + obj.timestamp)
        var sql = "INSERT INTO kap (uuid, value) VALUES ('"
                    + vsspathlist.uuidlist['LeafPaths'][obj.requestId] 
                    + "','" + parseInt(obj.value)  + "')"
                    // + "',UNIX_TIMESTAMP(STR_TO_DATE('" + obj.timestamp + "', '%Y-%m-%dT%H:%M:%sZ'))"
        console.log(sql)
        con.query(sql, function (err, result) {
            if (err) throw err;
            console.log("1 record inserted")
        });
    },
    deinitDb: function() {
        conn.end();
    }
}

