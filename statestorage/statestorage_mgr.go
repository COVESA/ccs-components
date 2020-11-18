/**
* (C) 2020 Geotab Inc
*
* All files and artifacts in the repository at https://github.com/UlfBj/ccs-w3c-client
* are licensed under the provisions of the license provided by the LICENSE file in this repository.
*
**/

package main

import (
    "io/ioutil"
    "os"
    "strings"
    "strconv"
    "encoding/json"
    "time"

    "database/sql"
    "fmt"
    _ "github.com/mattn/go-sqlite3"
)
type PathList struct {
	LeafPaths []string
}

var pathList PathList

var db *sql.DB
var dbErr error

func createStaticTables() int {
	stmt1, err := db.Prepare(`CREATE TABLE "VSS_MAP" ( "signal_id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "path" TEXT NOT NULL, "value" TEXT, "timestamp" TEXT )`)
	checkErr(err)

	_, err = stmt1.Exec()
	checkErr(err)

/*        stmt2, err2 := db.Prepare(`CREATE TABLE "NATIVE_VALUE" ( "signal_id" INTEGER NOT NULL, "int_value" INTEGER, "float_value" FLOAT, "boolean_value" BOOLEAN, FOREIGN KEY("signal_id") REFERENCES "VSS_MAP"("signal_id") )`)
        checkErr(err2)

	_, err2 = stmt2.Exec()
	checkErr(err2)*/

	if err != nil /*|| err2 != nil*/ {
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

func InitNewDb(dbFile string) {
	if (fileExists(dbFile)) {
		fmt.Printf("\ndataStorageMgr: DB %s already exist, cannot create a new with same name.\n", dbFile)
		os.Exit(1)
	}
	db, dbErr = sql.Open("sqlite3", dbFile) // Creates a new dbFile(?)
	checkErr(dbErr)
	err := createStaticTables()
	if err != 0 {
		fmt.Printf("\novdsServer: Unable to make static tables : %s\n", err)
		os.Exit(1)
	}
}

func InitExistingDb(dbFile string) {
	if (fileExists(dbFile)) {
		db, dbErr = sql.Open("sqlite3", dbFile)
		checkErr(dbErr)
	} else {
		fmt.Printf("\nDB %s must exist, or else the statestorage manager must be started also with a filename to a JSON pathlist.\nSee README\n", dbFile)
		os.Exit(1)
	}
}

func jsonToStructList(jsonList string, elements int) int {
	err := json.Unmarshal([]byte(jsonList), &pathList) //exclude curly braces when only one key-value pair
	if err != nil {
		fmt.Printf("Error unmarshal json=%s\n", err)
		return -1
	}
	return 0
}

func createVssList(fname string) int {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", fname, err)
		return -1
	}
	elements := strings.Count(string(data), "{")
	return jsonToStructList(string(data), elements)
}

func checkErr(err error) {
	if err != nil {
		fmt.Printf("Checkerr(): ")
		fmt.Println(err)
	}
}

func writeVssEntry(path string) int {
	stmt, err := db.Prepare("INSERT INTO VSS_MAP (path) values(?)")
	checkErr(err)
	if err != nil {
		return -1
	}

	_, err = stmt.Exec(path)
	checkErr(err)
	if err != nil {
		return -1
	}
	return 0
}


func runVssList() int {
	elements := len(pathList.LeafPaths)
	for i := 0; i < elements; i++ {
	    if (writeVssEntry(pathList.LeafPaths[i]) == -1) {
	        return -1
	    }
	}
	return 0
}

func populateVSS(vssPathFile string) int {
    fmt.Printf("Creating the database and populatng it with VSS paths...")
    if createVssList(vssPathFile) != 0 {
	return -1
    }
    return runVssList()
}

func createDomainTable(domainName string) {
	sqlString := "CREATE TABLE " + domainName + "_MAP (`signal_id` INTEGER NOT NULL, `handle` TEXT NOT NULL, `scale` FLOAT DEFAULT 1.0, `offset` FLOAT DEFAULT 0.0, `data_type` TEXT DEFAULT 'unknown', FOREIGN KEY(`signal_id`) REFERENCES `VSS_MAP`(`signal_id`) )"
	stmt, err := db.Prepare(sqlString)
	checkErr(err)

	_, err = stmt.Exec()
	checkErr(err)
}

func getNonMappedPaths(domainName string, rows **sql.Rows) int {
	var err error
	domainTableName := domainName + "_MAP"
	sqlString := "SELECT signal_id, path FROM VSS_MAP WHERE signal_id NOT IN (SELECT signal_id FROM " + domainTableName + ")"
	*rows, err = db.Query(sqlString)
	checkErr(err)
	if err != nil {
		return -1
	}
	return 0
}

func getNextPath(rows **sql.Rows) string {
    var signalId int
    var path string
    (*rows).Next()
    err := (*rows).Scan(&signalId, &path)
    checkErr(err)
    if err != nil {
	return ""
    }
    return path
}

func getSignalIdForPath(vssPath string) int {
	rows, err := db.Query("SELECT `signal_id` FROM VSS_MAP WHERE `path`=?", vssPath)
	checkErr(err)
	if err != nil {
		return -1
	}
	var signalId int

	rows.Next()
	err = rows.Scan(&signalId)
	checkErr(err)
	if err != nil {
		return -1
	}
	rows.Close()
	return signalId
}

func createMap(domainName string, vssPath string, handle string) int {
	domainTableName := domainName + "_MAP"
	signalId := getSignalIdForPath(vssPath)
	sqlString := "INSERT INTO " + domainTableName + "(signal_id, handle) values(?,?)"
	stmt, err := db.Prepare(sqlString)
	checkErr(err)
	if err != nil {
		return -1
	}

	_, err = stmt.Exec(signalId, handle)
	checkErr(err)
	if err != nil {
		return -1
	}
	return 0
}

func domainTableExists(domainName string) bool {
	domainTableName := domainName + "_MAP"
	sqlString := "SELECT `signal_id` FROM `" + domainTableName + "`"
	rows, err := db.Query(sqlString)
	if err != nil {
		return false
	}
        rows.Close()
	return true
}

func isMapped(domainName string, path string) bool {
	rows, err := db.Query("SELECT `signal_id` FROM VSS_MAP WHERE `path`=?", path)
	checkErr(err)
	if err != nil {
		return false
	}
	var signalId int

	rows.Next()
	err = rows.Scan(&signalId)
	checkErr(err)
	if err != nil {
		return false
	}
	defer rows.Close()
	rows, err = db.Query("SELECT `handle` FROM " + domainName + "_MAP WHERE `signal_id`=?", signalId)
	checkErr(err)
	if err != nil {
		return false
	}
	if (rows.Next() == false) {
	    return false
	}
	return true
}

func populateProprietary() {
    var domainName, command, vssPath string
    fmt.Printf("Name of non-VSS domain:")
    fmt.Scanf("%s", &domainName)
    if (!domainTableExists(domainName)) {
        createDomainTable(domainName)
    }
    var rows *sql.Rows
    if (getNonMappedPaths(domainName, &rows) == -1) {
        fmt.Printf("All VSS paths already mapped.\n")
        os.Exit(0)
    }
    command = "a"  //anything but s
    for {
        if (command[0] != 's') {
            vssPath = getNextPath(&rows)
            if (len(vssPath) == 0) {
                break
            }
        }
        fmt.Printf("VSS path to be mapped: %s\n", vssPath)
        fmt.Printf("\nSelect command - m(ap)/t(ransform)/d(atatype)/n(ext)/s(earch)/q(uit):")
        fmt.Scanf("%s", &command)
        switch command {
          case "m": fallthrough
          case "map":
              var handle string
              fmt.Printf("%s handle to be mapped to %s:", domainName, vssPath)
              fmt.Scanf("%s", &handle)
              rows.Close() // unlock DB for createMap()
              createMap(domainName, vssPath, handle)
              doTransformation(domainName, vssPath)
              doDatatype(domainName, vssPath)
              getNonMappedPaths(domainName, &rows)  // restart
          case "t": fallthrough
          case "transform":
              if (isMapped(domainName, vssPath) == true) {
                  rows.Close() // unlock DB for doTransformation()
                  doTransformation(domainName, vssPath)
                  getNonMappedPaths(domainName, &rows)  // restart
              } else {
                  fmt.Printf("The path must be mapped first.\n")
                  command = "s"
              }
          case "d": fallthrough
          case "datatype":
              if (isMapped(domainName, vssPath) == true) {
                  rows.Close() // unlock DB for doDatatype()
                  doDatatype(domainName, vssPath)
                  getNonMappedPaths(domainName, &rows)  // restart
              } else {
                  fmt.Printf("The path must be mapped first.\n")
                  command = "s"
              }
          case "n": fallthrough
          case "next": continue
          case "s": fallthrough
          case "search":
              fmt.Printf("VSS path to search for:")
              fmt.Scanf("%s", &vssPath)
              rows.Close()
              getNonMappedPaths(domainName, &rows)  //start from beginning
              for {
                      path := getNextPath(&rows)
  fmt.Printf("path= %s\n", path)
                      if len(path) == 0 {
	                    return
                      }
                      if (path == vssPath) {
                          vssPath = path
                          break
                      }
              }
          case "q": fallthrough
          case "quit": 
              db.Close()
              return
          case "w" :
              var handle string
              fmt.Printf("non-VSS handle:")
              fmt.Scanf("%s", &handle)
              var value string
              fmt.Printf("Value:")
              fmt.Scanf("%s", &value)
              rows.Close() // unlock DB for writeData()
              writeData(domainName, handle, value)
              getNonMappedPaths(domainName, &rows)  // restart
          default: 
              fmt.Printf("Invalid command.\n")
              command = "s"
        }
    }
}

func main() {
        if (len(os.Args) == 3) {
            InitNewDb(os.Args[1])
            defer db.Close()
            if (populateVSS(os.Args[2]) == -1) {
                fmt.Printf("Failed to populate DB with VSS paths\n")
                os.Exit(1)
            }
            fmt.Printf("\nDone.\n")
            os.Exit(0)
        } else if (len(os.Args) == 2) {
            InitExistingDb(os.Args[1])
            defer db.Close()
            populateProprietary()
            os.Exit(0)
        } else {
            fmt.Printf("Vehicle state storage manager command line must either be started:\n" +
             "- with a file name to a database, and a file name to a file containing VSS-paths, \n" +
             "when a new database with VSS mapping is to be created.\n" +
             "- with a file name to a database only, when proprietary mapping is to be done.\n")
            os.Exit(1)
        }
}

func getRfcTime() string {
    withTimeZone := time.Now().Format(time.RFC3339)   // 2020-05-01T15:34:35+02:00
    if (withTimeZone[len(withTimeZone)-6] == '+') {
        return withTimeZone[:len(withTimeZone)-6] + "Z"
    } else {
        return withTimeZone
    }
}

func writeData(domainName string, handle string, value string) {
	rows, err := db.Query("SELECT `signal_id` FROM " + domainName + "_MAP WHERE `handle`=?", handle)
	checkErr(err)
	if err != nil {
		return
	}
	var signalId int

	rows.Next()
	err = rows.Scan(&signalId)
	checkErr(err)
	if err != nil {
		return
	}
	rows.Close()
	fmt.Printf("signalId=%d\n",signalId)
	stmt, err2 := db.Prepare("UPDATE VSS_MAP SET value=?, timestamp=? WHERE `signal_id`=?")
	checkErr(err2)
	if err2 != nil {
		return
	}

	_, err2 = stmt.Exec(value, getRfcTime(), signalId)
	checkErr(err2)
	if err2 != nil {
		return
	}
	return
}

func doTransformation(domainName string, vssPath string) {
              var yesno string
              var scale string
              var offset string
              fmt.Printf("Do you want to enter transformation (y/n):")
              fmt.Scanf("%s", &yesno)
              if (yesno == "y") {
                  fmt.Printf("Scale:")
                  fmt.Scanf("%s", &scale)
                  fmt.Printf("Offset:")
                  fmt.Scanf("%s", &offset)
                  updateTransformation(domainName, vssPath, scale, offset)
              }
}

func updateTransformation(domainName string, path string, scaleStr string, offsetStr string) {
	scale, err := strconv.ParseFloat(scaleStr, 8)
	if err != nil {
		fmt.Printf("Error converting %s to float.\n",scaleStr)
		return
	}
	offset, err := strconv.ParseFloat(offsetStr, 8)
	if err != nil {
		fmt.Printf("Error converting %s to float.\n",offsetStr)
		return
	}
	rows, err := db.Query("SELECT `signal_id` FROM VSS_MAP WHERE `path`=?", path)
	checkErr(err)
	if err != nil {
		return
	}
	var signalId int

	rows.Next()
	err = rows.Scan(&signalId)
	checkErr(err)
	if err != nil {
		return
	}
	rows.Close()
	fmt.Printf("signalId=%d\n",signalId)
	stmt, err2 := db.Prepare("UPDATE " + domainName + "_MAP SET scale=?, offset=? WHERE `signal_id`=?")
	checkErr(err2)
	if err2 != nil {
		return
	}

	_, err2 = stmt.Exec(scale, offset, signalId)
	checkErr(err2)
	if err2 != nil {
		return
	}
}

func doDatatype(domainName string, vssPath string) {
              var yesno string
              var datatype string
              fmt.Printf("Do you want to enter datatype (y/n):")
              fmt.Scanf("%s", &yesno)
              if (yesno == "y") {
                  fmt.Printf("Datatype:")
                  fmt.Scanf("%s", &datatype)
                  updateDatatype(domainName, vssPath, datatype)
              }
}

func updateDatatype(domainName string, path string, datatype string) {
	rows, err := db.Query("SELECT `signal_id` FROM VSS_MAP WHERE `path`=?", path)
	checkErr(err)
	if err != nil {
		return
	}
	var signalId int

	rows.Next()
	err = rows.Scan(&signalId)
	checkErr(err)
	if err != nil {
		return
	}
	rows.Close()
	fmt.Printf("signalId=%d\n",signalId)
	stmt, err2 := db.Prepare("UPDATE " + domainName + "_MAP SET data_type=? WHERE `signal_id`=?")
	checkErr(err2)
	if err2 != nil {
		return
	}

	_, err2 = stmt.Exec(datatype, signalId)
	checkErr(err2)
	if err2 != nil {
		return
	}
}

