/**
* (C) 2020 Geotab Inc
*
* All files and artifacts in the repository at https://github.com/UlfBj/ccs-w3c-client
* are licensed under the provisions of the license provided by the LICENSE file in this repository.
*
**/

package main

import (
    "net/http"
    "encoding/json"
    "io/ioutil"
    "os"
    "strconv"
    "strings"

    "database/sql"
    "fmt"
    _ "github.com/mattn/go-sqlite3"
)

type searchData_t struct { // searchData_t defined in vssparserutilities.h
	responsePath    [512]byte // vssparserutilities.h: #define MAXCHARSPATH 512; typedef char path_t[MAXCHARSPATH];
	foundNodeHandle int64     // defined as long in vssparserutilities.h
}

var ovdsPort string

var db *sql.DB
var dbErr error

var errorResponseMap = map[string]interface{}{
	"action": "unknown",
	"error":  `{"number":AA, "reason": "BB", "message": "CC"}`,
}

func createStaticTables() int {
	stmt1, err := db.Prepare(`CREATE TABLE "VIN_TIV" ( "vin_id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "vin" TEXT NOT NULL )`)
	checkErr(err,1)

	_, err = stmt1.Exec()
	checkErr(err,2)

        stmt2, err2 := db.Prepare(`CREATE TABLE "TIV" ( "vin_id" INTEGER NOT NULL, "path" TEXT NOT NULL, "value" TEXT, UNIQUE("vin_id", "path") ON CONFLICT IGNORE, FOREIGN KEY("vin_id") REFERENCES "VIN_TIV"("vin_id") )`)
        checkErr(err2,3)

	_, err2 = stmt2.Exec()
	checkErr(err2,4)

	if err != nil || err2 != nil {
		return -1
	}
	return 0
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func InitDb(dbFile string) {
        doCreate := true
	if fileExists(dbFile) {
	    doCreate = false
	}
	db, dbErr = sql.Open("sqlite3", dbFile)
	checkErr(dbErr,5)
	if (doCreate) {
		err := createStaticTables()
		if err != 0 {
			fmt.Printf("\novdsServer: Unable to make static tables : %s\n", err)
			os.Exit(1)
		}
	}
}

func writeVIN(vin string) int {
	stmt, err := db.Prepare("INSERT INTO VIN_TIV(vin) values(?)")
	checkErr(err,6)
	if err != nil {
		return -1
	}

	_, err = stmt.Exec(vin)
	checkErr(err,7)
	if err != nil {
		return -1
	}
	return 0
}

func writeTvValue(vinId int, path string, value string, timestamp string) int {
	tableName := "TV_" + strconv.Itoa(vinId)
	sqlString := "INSERT INTO " + tableName + "(value, timestamp, path) values(?, ?, ?)"
	stmt, err := db.Prepare(sqlString)
	checkErr(err,8)
	if err != nil {
		return -1
	}

	_, err = stmt.Exec(value, timestamp, path)
	checkErr(err,9)
	if err != nil {
		return -1
	}
	return 0
}

func writeTivValue(vinId int, path string, value string) int {
        sqlString := "INSERT INTO TIV (vin_id, path) VALUES(?, ?)"
        stmt, err := db.Prepare(sqlString)
        checkErr(err,10)
        if (err != nil) {
fmt.Printf("writeTivValue:prepare-INSERT OR IGNORE error\n")
            return -1
        }

        _, err = stmt.Exec(vinId, path)
        checkErr(err,11)
        if (err != nil) {
fmt.Printf("writeTivValue:exec-INSERT OR IGNORE error\n")
            return -1
        }

        sqlString = "UPDATE TIV SET `value`=? WHERE `vin_id`=? AND `path`=?"
        stmt, err = db.Prepare(sqlString)
        checkErr(err,12)
        if (err != nil) {
fmt.Printf("writeTivValue:prepare-UPDATE error\n")
            return -1
        }

        _, err = stmt.Exec(value, vinId, path)
        checkErr(err,13)
        if (err != nil) {
fmt.Printf("writeTivValue:exec-UPDATE error\n")
            return -1
        }

        return 0
}

func readVinId(vin string) int {
	rows, err := db.Query("SELECT `vin_id` FROM VIN_TIV WHERE `vin`=?", vin)
	checkErr(err,14)
	if err != nil {
		return -1
	}
	defer rows.Close()
	var vinId int

	rows.Next()
	err = rows.Scan(&vinId)
	checkErr(err,15)
	if err != nil {
		return -1
	}
	return vinId
}

func readTivValue(vinId int, path string) string {
	rows, err := db.Query("SELECT `value` FROM TIV WHERE `vin_id`=? AND `path`=?", vinId, path)
	checkErr(err,16)
	if err != nil {
		return ""
	}
	defer rows.Close()
	var value string

	rows.Next()
	err = rows.Scan(&value)
	checkErr(err,17)
	if err != nil {
		return ""
	}
	return value
}

func readMax(tableName string, columnName string, path string) string {
	sqlString := "SELECT MAX(" + columnName + ") FROM " + tableName + " WHERE `path`=? "
	rows, err := db.Query(sqlString, path)
	defer rows.Close()

	var maxValue string
	rows.Next()
	err = rows.Scan(&maxValue)
	checkErr(err,18)
	if err != nil {
		return ""
	}
	return maxValue
}

func readTvValue(vinId int, path string, from string, to string, maxSamples string) string {
fmt.Printf("readTvValue:vinId=%d, path=%s, from=%s, to=%s, maxSamples=%s\n", vinId, path, from, to, maxSamples)
	var rows *sql.Rows
	var err error
	tableName := "TV_" + strconv.Itoa(vinId)
	sqlStringCommon := "SELECT `value`, `timestamp` FROM " + tableName + " WHERE `path`=? AND "
	if len(from) != 0 && len(to) != 0 {
 		    sqlString := sqlStringCommon + "`timestamp` > ? AND `timestamp` < ?"
		    rows, err = db.Query(sqlString, path, from, to)
	} else if len(from) != 0 && len(to) == 0 {
	        if (len(maxSamples) == 0) {
		    sqlString := sqlStringCommon + "`timestamp` > ?"
		    rows, err = db.Query(sqlString, path, from)
		} else {
		    sqlString := sqlStringCommon + "`timestamp` > ?  LIMIT ?"
		    rows, err = db.Query(sqlString, path, from, maxSamples)
		}
	} else if len(from) == 0 && len(to) == 0 {
		maxTs := readMax(tableName, "timestamp", path)
		sqlString := sqlStringCommon + "`timestamp` = ?"
		rows, err = db.Query(sqlString, path, maxTs)
	} else {
		fmt.Printf("readTvValue: DB not read.\n")
		return ""
	}
	defer rows.Close()
	checkErr(err,19)
	if err != nil {
		return ""
	}
	var value string
	var timestamp string
	datapoints := "["
	numOfDatapoints := 0

	for rows.Next() {
		err = rows.Scan(&value, &timestamp)
		checkErr(err,20)
		if err != nil {
			return ""
		}
		datapoints += `{"value":"` + value + `","ts":"` + timestamp + `"}, `
		numOfDatapoints++
	}
	if (numOfDatapoints == 0) {
 	    fmt.Printf("readTvValue: Data not found.\n")
	    return ""
	}
	datapoints = datapoints[:len(datapoints)-2]
//	if numOfDatapoints > 1 {
		datapoints += "]"  // livesim expects it always to be declared as an array
/*	} else {
		datapoints = datapoints[1:]
	}*/
	return datapoints
}

func createTvVin(vinId int) {
	tableName := "TV_" + strconv.Itoa(vinId)
	sqlString := "CREATE TABLE " + tableName + " (`value` TEXT NOT NULL, `timestamp` TEXT NOT NULL, `path` TEXT, UNIQUE(`path`, `timestamp`) ON CONFLICT IGNORE)"
	stmt, err := db.Prepare(sqlString)
	checkErr(err,21)

	_, err = stmt.Exec()
	checkErr(err,22)
}

func makeOVDSServerHandler(serverChannel chan string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
//		fmt.Printf("OVDSServer:url=%s\n", req.URL.Path)
		if req.URL.Path != "/ovdsserver" {
			http.Error(w, "404 url path not found.", 404)
		} else if req.Method != "POST" {
			http.Error(w, "400 bad request method.", 400)
		} else {
                        bodyBytes, err := ioutil.ReadAll(req.Body)
                        if err != nil {
                                http.Error(w, "400 request unreadable.", 400)
                        } else {
				fmt.Printf("OVDSserver:received POST request=%s\n", string(bodyBytes))
				serverChannel <- string(bodyBytes)
				response := <- serverChannel
				fmt.Printf("OVDSserver:POST response=%s\n", response)
                                if (len(response) == 0) {
                                    http.Error(w, "400 bad input.", 400)
                                } else {
	                            w.Header().Set("Access-Control-Allow-Origin", "*")
//				    w.Header().Set("Content-Type", "application/json")
				    w.Write([]byte(response))
                                }
                        }
		}
	}
}

func initOVDSServer(serverChannel chan string, muxServer *http.ServeMux) {
//	fmt.Printf("initOVDSServer(): :8765/ovdsserver")
	fmt.Printf("initOVDSServer(): :" + ovdsPort + "/ovdsserver")
	agtServerHandler := makeOVDSServerHandler(serverChannel)
	muxServer.HandleFunc("/ovdsserver", agtServerHandler)
//	fmt.Println(http.ListenAndServe(":8765", muxServer))
	fmt.Println(http.ListenAndServe(":" + ovdsPort, muxServer))
}

func jsonToMap(request string, rMap *map[string]interface{}) {
	decoder := json.NewDecoder(strings.NewReader(request))
	err := decoder.Decode(rMap)
	if err != nil {
		fmt.Printf("jsonToMap: JSON decode failed for request:%s, err=%s\n", request, err)
	}
}

func finalizeMessage(responseMap map[string]interface{}) string {
	response, err := json.Marshal(responseMap)
	if err != nil {
		fmt.Print("Server core-FinalizeMessage: JSON encode failed. ", err)
		return `{"error":{"number":400,"reason":"JSON marshal error","message":""}}` //???
	}
	return string(response)
}

func setErrorResponse(reqMap map[string]interface{}, errRespMap map[string]interface{}, number string, reason string, message string) {
	if reqMap["action"] != nil {
		errRespMap["action"] = reqMap["action"]
	}
	errRespMap["error"] = `{"number":` + number + `,"reason":"` + reason + `","message":"` + message + `"}`
}

func translateNodeType(nodeType int) string {
	switch nodeType {
	case 1:
		return "SENSOR"
	case 2:
		return "ACTUATOR"
	case 3:
		return "ATTRIBUTE"
	case 4:
		return "BRANCH"
	}
	return "unknown nodetype"
}

func translateDataType(dataType int) string {
	switch dataType {
	case 1:
		return "INT8"
	case 2:
		return "UINT8"
	case 3:
		return "INT16"
	case 4:
		return "UINT16"
	case 5:
		return "INT32"
	case 6:
		return "UINT32"
	case 7:
		return "DOUBLE"
	case 8:
		return "FLOAT"
	case 9:
		return "BOOLEAN"
	case 10:
		return "STRING"
	}
	return "unknown datatype"
}

func UrlToPath(url string) string {
	var path string = strings.TrimPrefix(strings.Replace(url, "/", ".", -1), ".")
	return path[:]
}

func getPathLen(path string) int {
	for i := 0; i < len(path); i++ {
		if path[i] == 0x00 { // the path buffer defined in searchData_t is initiated with all zeros
			return i
		}
	}
	return len(path)
}

func OVDSGetValue(reqMap map[string]interface{}) (string, int) {
	if reqMap["vin"] == nil {
		return "", 1
	}
	vin := reqMap["vin"].(string)
	vinId := readVinId(vin)
	if vinId == -1 {
		return "", 4
	}
	if reqMap["path"] == nil {
		return "", 2
	}
	path := reqMap["path"].(string)
	response := ""
	var from string
	var maxSamples string
	if reqMap["from"] == nil {
		from = ""
	} else {
		from = reqMap["from"].(string)
	}
	var to string
	if reqMap["to"] == nil {
		to = ""
	} else {
		to = reqMap["to"].(string)
	}
	if reqMap["maxsamples"] == nil {
		maxSamples = ""
	} else {
	    maxSamples = reqMap["maxsamples"].(string)
/*	        var err error
		maxSamples, err = strconv.Atoi(reqMap["maxsamples"].(string))
		if (err != nil) {
		    fmt.Printf("Maxsamples invalid, err=%s\n", err)
		    maxSamples = 0
		}*/
	}
	datapoints := readTvValue(vinId, path, from, to, maxSamples)
	if (len(datapoints) == 0) {
	    value := readTivValue(vinId, path)
	    if (len(value) == 0) {
	        return "", 5
	    }
	    response += `{"path":"` + path + `","dp":[{"value":"` + value + `","ts":""}]}, `
	} else {
	    response += `{"path":"` + path + `","dp":` + datapoints + `}, `
	}
	response = response[:len(response)-2]
	return response, 0
}

/*func extractData(dataMap map[string]interface{}) (string, string, string) {
//    var dataMap = make(map[string]interface{})
//    jsonToMap(data, &dataMap)
    if dataMap["path"] == nil {
	return "", "", ""
    }
    path := dataMap["path"].(string)
    value, ts := extractDp(dataMap["dp"].(map[string]interface{}))
    return path, value, ts
}

func extractDp(dpMap map[string]interface{}) (string, string) {
//    var dpMap = make(map[string]interface{})
//    jsonToMap(dataPoint, &dpMap)
    if dpMap["value"] == nil {
	return "", ""
    }
    value := dpMap["value"].(string)
    if dpMap["ts"] == nil {
	return value, ""
    }
    return value, dpMap["ts"].(string)
}*/

func OVDSSetValue(reqMap map[string]interface{}) string {
	if reqMap["vin"] == nil {
		return "VIN missing"
	}
	vin := reqMap["vin"].(string)
	if reqMap["path"] == nil {
		return "Path missing"
	}
	path := reqMap["path"].(string)
	if reqMap["value"] == nil {
		return "Value missing"
	}
	value := reqMap["value"].(string)
	timeInvariantNode := true
	timestamp := ""
	if reqMap["timestamp"] != nil {
	    timeInvariantNode = false
	    timestamp = reqMap["timestamp"].(string)
	    timestamp = strings.Replace(timestamp, ".", ":", -1)  // Adapter  has incorrect format yyyy-mm-ddThh.mm.ssZ
	}
	if len(path) == 0 {
		return "Data invalid"
	}
	vinId := readVinId(vin)
	if vinId == -1 {
		err := writeVIN(vin)
		if err != 0 {
			return "Failed to write VIN"
		}
		vinId = readVinId(vin)
		//fmt.Printf("Second attempt to read vinId=%d\n", vinId)
		if vinId == -1 {
			return "Failed to create VIN entry"
		}
		createTvVin(vinId)
	}
	var err int
	if timeInvariantNode == true {
		err = writeTivValue(vinId, path, value)
	} else {
		err = writeTvValue(vinId, path, value, timestamp)
	}
	if err != 0 {
		return "Failed to store sample"
	}
	return "200 OK"
}

func cleanupResponse(resp string) string {
    resp = fixEscapeChars(resp)
    resp = strings.Replace(resp, "\"{", "{", -1)
    resp = strings.Replace(resp, "}\"", "}", -1)
    resp = strings.Replace(resp, "\"[{", "[{", -1)
    resp = strings.Replace(resp, "}]\"", "}]", -1)
    return resp
}

func nextQuoteMark(message string) int {
    for i := 0 ; i < len(message) ; i++ {
        if (message[i] == '"') {
            return i
        }
    }
    return -1
}

func fixEscapeChars(resp string) string {  // keep in arrrays of scalars, else remove
    arrayEnd := 0
    arrayFront := strings.Index(resp, "[")
    for arrayFront > arrayEnd {
fmt.Printf("\n\nArrayFront=%d\n", arrayFront)
fmt.Printf("nextQuoteMark=%d\n\n", nextQuoteMark(resp[arrayFront:]))
        if (arrayFront >= 0 && nextQuoteMark(resp[arrayFront:]) == 2) {
            resp = resp[:arrayEnd] + strings.Replace(resp[arrayEnd:arrayFront+1], "\\", "", -1) + resp[arrayFront+1:]
            arrayFront = arrayEnd + strings.Index(resp[arrayEnd:], "[")  // recalibrate arrayFront
            arrayEnd = arrayFront + strings.Index(resp[arrayFront:], "]")
            arrayFront = arrayEnd + strings.Index(resp[arrayEnd:], "[")
//fmt.Printf("\n\nArray=%s\n\n\n", resp[arrayFront:arrayEnd])
        } else {
            bracketIndex := strings.Index(resp[arrayFront+1:], "[")
            if (bracketIndex >= 0) {
                arrayFront += 1 + bracketIndex
            } else {
                arrayFront = arrayEnd -1
            }
        }
    }
    resp = resp[:arrayEnd] + strings.Replace(resp[arrayEnd:], "\\", "", -1)
    return resp
}

func main() {

        if (len(os.Args) < 2 || len(os.Args) > 3) {
            fmt.Printf("The command to run the OVDS server must have input parameter as shown:\n./ovds_server db-file-name\nor\n./ovds_server db-file-name livesim\n")
            os.Exit(1)
        }
        ovdsPort = "8765"
        if (len(os.Args) == 3) {
            if (os.Args[2] == "livesim") {
                ovdsPort = "8766"  // OVDS to be used together with livesim
	        fmt.Printf("OVDS server to be used with the livesim vehicle data simulator.\n")
            } else {
                fmt.Printf("The command to run the OVDS server together with livesim must have input parameter as shown:\n./ovds_server db-file-name livesim\n")
                os.Exit(1)
            }
        }

	serverChan := make(chan string)
	muxServer := http.NewServeMux()

        InitDb(os.Args[1])
        defer db.Close()

        go initOVDSServer(serverChan, muxServer)

	for {
		select {
		case request := <-serverChan:
			fmt.Printf("main loop:request received\n")
			var requestMap = make(map[string]interface{})
			var responseMap = make(map[string]interface{})
			var err int
			jsonToMap(request, &requestMap)
			//			responseMap["action"] = requestMap["action"]
			switch requestMap["action"] {
			case "get":
				responseMap["datapackage"], err = OVDSGetValue(requestMap)
                                if (err != 0) {
                                       switch err {
                                       case 1:
		                           setErrorResponse(requestMap, errorResponseMap, "400", "Missing vin.", "")
                                       case 2:
		                           setErrorResponse(requestMap, errorResponseMap, "400", "Missing path.", "")
                                       case 3:
		                           setErrorResponse(requestMap, errorResponseMap, "400", "No matching path.", "")
                                       case 4:
		                           setErrorResponse(requestMap, errorResponseMap, "400", "No matching VIN.", "")
                                       case 5:
		                           setErrorResponse(requestMap, errorResponseMap, "400", "No data points found.", "")
                                       }
			                serverChan <- finalizeMessage(errorResponseMap)
                                       break
                                }
                               resp := finalizeMessage(responseMap)
                               resp = cleanupResponse(resp)   // due to simplistic map handling...
			        serverChan <- resp
			case "set":
				responseMap["status"] = OVDSSetValue(requestMap)
			        serverChan <- finalizeMessage(responseMap)
			case "getmetadata":
//				responseMap["metadata"] = OVDSGetMetadata(requestMap)
//			        serverChan <- finalizeMessage(responseMap)
                                fallthrough   //until OVDSGetMetadata implemented
			default:
				setErrorResponse(requestMap, errorResponseMap, "400", "Unknown action.", "Supported actions: set/set/getmetadata")
				serverChan <- finalizeMessage(errorResponseMap)
			} // switch
		}
	}
}

func checkErr(err error, id int) {
	if err != nil {
		fmt.Printf("checkErr: err=%s, id=%d\n", err, id)
	}
}
