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
    "sort"
    "unsafe"

    "database/sql"
    "fmt"
    _ "github.com/mattn/go-sqlite3"
)

// #include <stdlib.h>
// #include <stdio.h>
// #include <stdbool.h>
// #include "vssparserutilities.h"
import "C"

var VSSTreeRoot C.long

type searchData_t struct { // searchData_t defined in vssparserutilities.h
	responsePath    [512]byte // vssparserutilities.h: #define MAXCHARSPATH 512; typedef char path_t[MAXCHARSPATH];
	foundNodeHandle int64     // defined as long in vssparserutilities.h
}

var db *sql.DB
var dbErr error

var errorResponseMap = map[string]interface{}{
	"action": "unknown",
	"error":  `{"number":AA, "reason": "BB", "message": "CC"}`,
}

func createStaticTables() int {
	stmt1, err := db.Prepare(`CREATE TABLE "VIN_TIV" ( "vin_id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "vin" TEXT NOT NULL )`)
	checkErr(err)

	_, err = stmt1.Exec()
	checkErr(err)

        stmt2, err2 := db.Prepare(`CREATE TABLE "TIV" ( "vin_id" INTEGER NOT NULL, "uuid" TEXT NOT NULL, "value" TEXT, UNIQUE("vin_id", "uuid") ON CONFLICT IGNORE, FOREIGN KEY("vin_id") REFERENCES "VIN_TIV"("vin_id") )`)
        checkErr(err2)

	_, err2 = stmt2.Exec()
	checkErr(err2)

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
	checkErr(dbErr)
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
	checkErr(err)
	if err != nil {
		return -1
	}

	_, err = stmt.Exec(vin)
	checkErr(err)
	if err != nil {
		return -1
	}
	return 0
}

func writeTvValue(vinId int, uuid string, value string, timestamp string) int {
	tableName := "TV_" + strconv.Itoa(vinId)
	sqlString := "INSERT INTO " + tableName + "(value, timestamp, uuid) values(?, ?, ?)"
	stmt, err := db.Prepare(sqlString)
	checkErr(err)
	if err != nil {
		return -1
	}

	_, err = stmt.Exec(value, timestamp, uuid)
	checkErr(err)
	if err != nil {
		return -1
	}
	return 0
}

func writeTivValue(vinId int, uuid string, value string) int {
        sqlString := "INSERT INTO TIV (vin_id, uuid) VALUES(?, ?)"
        stmt, err := db.Prepare(sqlString)
        checkErr(err)
        if (err != nil) {
fmt.Printf("writeTivValue:prepare-INSERT OR IGNORE error\n")
            return -1
        }

        _, err = stmt.Exec(vinId, uuid)
        checkErr(err)
        if (err != nil) {
fmt.Printf("writeTivValue:exec-INSERT OR IGNORE error\n")
            return -1
        }

        sqlString = "UPDATE TIV SET `value`=? WHERE `vin_id`=? AND `uuid`=?"
        stmt, err = db.Prepare(sqlString)
        checkErr(err)
        if (err != nil) {
fmt.Printf("writeTivValue:prepare-UPDATE error\n")
            return -1
        }

        _, err = stmt.Exec(value, vinId, uuid)
        checkErr(err)
        if (err != nil) {
fmt.Printf("writeTivValue:exec-UPDATE error\n")
            return -1
        }

        return 0
}

func readVinId(vin string) int {
	rows, err := db.Query("SELECT `vin_id` FROM VIN_TIV WHERE `vin`=?", vin)
	checkErr(err)
	if err != nil {
		return -1
	}
	var vinId int

	rows.Next()
	err = rows.Scan(&vinId)
	checkErr(err)
	if err != nil {
		return -1
	}
	rows.Close()
	return vinId
}

func readTivValue(vinId int, uuid string) string {
	rows, err := db.Query("SELECT `value` FROM TIV WHERE `vin_id`=? AND `uuid`=?", vinId, uuid)
	checkErr(err)
	if err != nil {
		return ""
	}
	var value string

	rows.Next()
	err = rows.Scan(&value)
	checkErr(err)
	if err != nil {
		return ""
	}
	rows.Close()
	return value
}

func readMax(tableName string, columnName string, uuid string) string {
	sqlString := "SELECT MAX(" + columnName + ") FROM " + tableName + " WHERE `uuid`=? "
	rows, err := db.Query(sqlString, uuid)

	var maxValue string
	rows.Next()
	err = rows.Scan(&maxValue)
	checkErr(err)
	if err != nil {
		return ""
	}
	rows.Close()
	return maxValue
}

func readTvValue(vinId int, uuid string, from string, to string, maxSamples int) string {
	var rows *sql.Rows
	var err error
	tableName := "TV_" + strconv.Itoa(vinId)
	sqlStringCommon := "SELECT `value`, `timestamp` FROM " + tableName + " WHERE `uuid`=? AND "
	if len(from) != 0 && len(to) != 0 {
		sqlString := sqlStringCommon + "`timestamp` > ? AND `timestamp` < ?"
		rows, err = db.Query(sqlString, uuid, from, to)
	} else if len(from) != 0 && len(to) == 0 {
		sqlString := sqlStringCommon + "`timestamp` > ?"
		rows, err = db.Query(sqlString, uuid, from)
	} else if len(from) == 0 && len(to) == 0 {
		maxTs := readMax(tableName, "timestamp", uuid)
		sqlString := sqlStringCommon + "`timestamp` = ?"
		rows, err = db.Query(sqlString, uuid, maxTs)
	} else {
		return ""
	}
	checkErr(err)
	if err != nil {
		return ""
	}
	var value string
	var timestamp string
	datapoints := "["
	numOfDatapoints := 0

	for rows.Next() {
		err = rows.Scan(&value, &timestamp)
		checkErr(err)
		if err != nil {
			return ""
		}
		datapoints += `{"value": "` + value + `", "timestamp": "` + timestamp + `"}, `
		numOfDatapoints++
		if (numOfDatapoints == maxSamples) {
		    break
		}
	}
	rows.Close()
	if (numOfDatapoints == 0) {
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
	sqlString := "CREATE TABLE " + tableName + " (`value` TEXT NOT NULL, `timestamp` TEXT NOT NULL, `uuid` TEXT, UNIQUE(`uuid`, `timestamp`) ON CONFLICT IGNORE)"
	stmt, err := db.Prepare(sqlString)
	checkErr(err)

	_, err = stmt.Exec()
	checkErr(err)
}

func makeOVDSServerHandler(serverChannel chan string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("OVDSServer:url=%s", req.URL.Path)
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
				fmt.Printf("OVDSserver:POST response=%s", response)
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
	fmt.Printf("initOVDSServer(): :8765/ovdsserver")
	agtServerHandler := makeOVDSServerHandler(serverChannel)
	muxServer.HandleFunc("/ovdsserver", agtServerHandler)
	fmt.Println(http.ListenAndServe(":8765", muxServer))
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

func initVssFile(filePath string) C.long {
	cfilePath := C.CString(filePath)
	root := C.VSSReadTree(cfilePath)
	C.free(unsafe.Pointer(cfilePath))

	return root
}

func translateNodeType(nodeType int) string {
	switch nodeType {
	case 9:
		return "STRING"
	case 10:
		return "SENSOR"
	case 11:
		return "ACTUATOR"
	case 12:
		return "STREAM"
	case 13:
		return "ATTRIBUTE"
	case 14:
		return "BRANCH"
	}
	return "unknown nodetype"
}

func translateDataType(dataType int) string {
	switch dataType {
	case 0:
		return "INT8"
	case 1:
		return "UINT8"
	case 2:
		return "INT16"
	case 3:
		return "UINT16"
	case 4:
		return "INT32"
	case 5:
		return "UINT32"
	case 6:
		return "DOUBLE"
	case 7:
		return "FLOAT"
	case 8:
		return "BOOLEAN"
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

func getVssDbMapping(path string) (string, int) {
	// call int VSSSearchNodes(char* searchPath, long rootNode, int maxFound, searchData_t* searchData, bool anyDepth, bool leafNodesOnly, int* validation);
	searchData := [1500]searchData_t{} // vssparserutilities.h: #define MAXFOUNDNODES 1500
	var anyDepth C.bool = false
	if path[len(path)-1] == '*' {
		anyDepth = true
	}
	var validation C.int = -1
	cpath := C.CString(UrlToPath(path))
	fmt.Printf("path=%s\n", path)
	var matches C.int = C.VSSSearchNodes(cpath, VSSTreeRoot, 1500, (*C.struct_searchData_t)(unsafe.Pointer(&searchData)), anyDepth, true, (*C.int)(unsafe.Pointer(&validation)))
	C.free(unsafe.Pointer(cpath))
	fmt.Printf("matches=%d\n", int(matches))
	dbMap := "["
	for i := 0; i < int(matches); i++ {
		uuid := C.GoString(C.VSSgetUUID((C.long)(searchData[i].foundNodeHandle)))
		var c_nodetype C.nodeTypes_t = C.VSSgetType((C.long)(searchData[i].foundNodeHandle))
		nodeType := translateNodeType(int(c_nodetype))
		var c_datatype C.nodeTypes_t = C.VSSgetDatatype((C.long)(searchData[i].foundNodeHandle))
		dataType := translateDataType(int(c_datatype))
		pathLen := getPathLen(string(searchData[i].responsePath[:]))
		dbMap += `{"path":"` + string(searchData[i].responsePath[:pathLen]) + `", "uuid":"` + uuid + `", "nodetype":"` + nodeType + `", "datatype":"` + dataType + `"}, `
	}
	if int(matches) > 0 {
		dbMap = dbMap[:len(dbMap)-2]
	}
	dbMap += "]"
	return dbMap, int(matches)
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
	output, matches := getVssDbMapping(path)
	if matches == 0 {
		return "", 3
	}
	elementStart := 0
	response := ""
	if matches > 1 {
		response = "["
	}
	for i := 0; i < matches; i++ {
		var treeMap = make(map[string]interface{})
		elementStop := strings.Index(output[elementStart:len(output)], "}")
		elementStart += strings.Index(output[elementStart+1:len(output)], "{") + 1
		jsonToMap(output[elementStart:elementStart+elementStop+1], &treeMap)
		nodetype := treeMap["nodetype"].(string)
		uuid := treeMap["uuid"].(string)
		var from string
		var maxSamples int
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
			maxSamples = 0
		} else {
		        var err error
			maxSamples, err = strconv.Atoi(reqMap["maxsamples"].(string))
			if (err != nil) {
			    fmt.Printf("Maxsamples invalid, err=%s\n", err)
 			    maxSamples = 0
			}
		}
		if nodetype == "ATTRIBUTE" {
			value := readTivValue(vinId, uuid)
			if (len(value) == 0) {
			    return "", 5
			}
			response += `{ "path":"` + path + `", "datapoints":[{"value":"` + value + `", "timestamp":""}]}, `
		} else {
			datapoints := readTvValue(vinId, uuid, from, to, maxSamples)
			if (len(datapoints) == 0) {
			    return "", 5
			}
			response += `{"path":"` + path + `", "datapoints":` + datapoints + `}, `
		}
	}
	response = response[:len(response)-2]
	if matches > 1 {
		response += "]"
	}
	return response, 0
}

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
	var timestamp string
	if reqMap["timestamp"] != nil {
		timestamp = reqMap["timestamp"].(string)
	}
	output, matches := getVssDbMapping(path)
	if matches != 1 {
		return "No matching path"
	}
	var dbMap = make(map[string]interface{})
	jsonToMap(output[1:len(output)-1], &dbMap)
	uuid := dbMap["uuid"].(string)
	nodetype := dbMap["nodetype"].(string)
	//fmt.Printf("nodetype=%s\n", nodetype)
	if nodetype != "ATTRIBUTE" && reqMap["timestamp"] == nil {
		return "Timestamp missing"
	}
	vinId := readVinId(vin)
	//fmt.Printf("First attempt to read vinId=%d\n", vinId)
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
	if nodetype == "ATTRIBUTE" {
		err = writeTivValue(vinId, uuid, value)
	} else {
		err = writeTvValue(vinId, uuid, value, timestamp)
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

type PathList struct {
	LeafPaths []string
}
var pathList PathList

func sortPathList(listFname string) {
	data, err := ioutil.ReadFile(listFname)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", listFname, err)
		return
	}
	err = json.Unmarshal([]byte(data), &pathList)
	if err != nil {
		fmt.Printf("Error unmarshal json=%s\n", err)
		return
	}
	sort.Strings(pathList.LeafPaths)
	file, _ := json.Marshal(pathList)
	_ = ioutil.WriteFile(listFname, file, 0644)
}

func createPathListFile(listFname string) {
	// call int VSSGetLeafNodesList(long rootNode, char* leafNodeList);
	clistFname := C.CString(listFname)
	C.VSSGetLeafNodesList(VSSTreeRoot, clistFname)
	C.free(unsafe.Pointer(clistFname))
	sortPathList(listFname)
}

func main() {

        if (len(os.Args) != 3) {
            fmt.Printf("./ovds_server db-file-name cnative-vss-tree-file-name\n")
            os.Exit(1)
        }

	serverChan := make(chan string)
	muxServer := http.NewServeMux()

	VSSTreeRoot = initVssFile(os.Args[2])
	if VSSTreeRoot == 0 {
		fmt.Println("VSS tree file not found")
		os.Exit(1)
	}

	createPathListFile("../vsspathlist.json")  // save in ovds directory, where ovds client will expect it to be

        InitDb(os.Args[1])
        defer db.Close()

        go initOVDSServer(serverChan, muxServer)

	for {
		select {
		case request := <-serverChan:
			fmt.Printf("main loop:request received")
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

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		//            panic(err)
	}
}
