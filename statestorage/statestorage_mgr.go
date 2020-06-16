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
    "unsafe"

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
	stmt1, err := db.Prepare(`CREATE TABLE "VSS_MAP" ( "signal_id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "path" TEXT NOT NULL, "value" TEXT )`)
	checkErr(err)

	_, err = stmt1.Exec()
	checkErr(err)

        stmt2, err2 := db.Prepare(`CREATE TABLE "NATIVE_VALUE" ( "signal_id" INTEGER NOT NULL, "int_value" INTEGER, "float_value" FLOAT, "boolean_value" BOOLEAN, FOREIGN KEY("signal_id") REFERENCES "VSS_MAP"("signal_id") )`)
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

func InitDb(dbFile string, isNewDB bool) {
	if (isNewDB == true && fileExists(dbFile)) {
		fmt.Printf("\ndataStorageMgr: DB %s already exist, cannot create a new with same name.\n", dbFile)
		os.Exit(1)
	} else {
		db, dbErr = sql.Open("sqlite3", dbFile)
		checkErr(dbErr)
		if (isNewDB == true) {
 		    err := createStaticTables()
		    if err != 0 {
			    fmt.Printf("\novdsServer: Unable to make static tables : %s\n", err)
			    os.Exit(1)
		    }
		}
	}

}


func createVssList(fname string) int {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", fname, err)
		return -1
	}
	elements := strings.Count(string(data), "{")
	fmt.Printf("Before jsonToStructList. elements=%d\n", elements)
	return jsonToStructList(string(data), elements)
}

func checkErr(err error) {
	if err != nil {
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


func runVssList(trimList bool) int {
	elements := len(pathList.LeafPaths)
	for i := 0; i < elements; i++ {
	    if (writeVssEntry(pathList.LeafPaths[i]) == -1) {
	        return -1)
	    }
	}
	return 0
}

func populateVSS(vssPathFile string) int {
    if createVssList(vssPathFile) != 0 {
	return -1
    }
    return runVssList()
}

func createDomainTable(domainName string) {
	sqlString := "CREATE TABLE " + domainName + "_MAP (`signal_id` INTEGER NOT NULL, `handle` TEXT NOT NULL, FOREIGN KEY(`signal_id`) REFERENCES `VSS_MAP`(`signal_id`) )"
	stmt, err := db.Prepare(sqlString)
	checkErr(err)

	_, err = stmt.Exec()
	checkErr(err)
}

func getNonMappedPaths(domainName string) *sql.Rows {
	var rows *sql.Rows
	var err error
	domainTableName := domainName + "_MAP"
	sqlString := "SELECT `path` FROM VSS_MAP WHERE VSS_MAP(`signal_id`) <> " + domainTableName + "(`signal_id`)"
	rows, err = db.Query(sqlString)
	checkErr(err)
	if err != nil {
		return nil
	}
	return rows
}

func getNextPath(rows *sql.Rows, domainName string) string {
    var path string
    rows.Next()
    err = rows.Scan(&path)
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

func populateProprietary() {
    var domainName, command, vssPath string
    fmt.Printf("Name of non-VSS domain:")
    fmt.Scanf("%s", &domainName)
    if (!domainTableExists(domainName)) {
        createDomainTable(domainName)
    }
    var rows *sql.Rows
    defer rows.Close()
    rows = getNonMappedPaths(domainName)
    command[0] = 'a'  //anything but s
    for {
        if (command[0] != 's') {
            vssPath = getNextPath(rows, domainName)
            if (len(vssPath) == 0) {
                break
            }
        }
        fmt.Printf("VSS path to be mapped: %s\n", vssPath)
        fmt.Printf("\nSelect command - m(ap)/n(ext)/s(earch)/q(uit):")
        fmt.Scanf("%s", &command)
        switch command {
          case "m": fallthrough
          case "map":
              fmt.Printf("%s handle to be mapped to %s:")
              fmt.Scanf("%s", &handle)
              createMap(domainName, vssPath, handle)
          case "n": fallthrough
          case "next": continue
          case "s": fallthrough
          case "search":
              fmt.Printf("VSS path to search for:")
              fmt.Scanf("%s", &vssPath)
              rows.First()
              var path string
              for {
                      rows.Next()
                      err = rows.Scan(&path)
                      checkErr(err)
                      if err != nil {
	                    break
                      }
                      if (path == vssPath) {
                          vssPath = path
                          continue
                      }
              }
          case "q": fallthrough
          case "quit": break
        }
    }
}

func main() {

        if (len(os.Args) == 3) {
            InitDb(os.Args[1], true)
            defer db.Close()
            if (populateVSS(os.Args[2]) == -1) {
                fmt.Printf("Failed to populate DB with VSS paths\n")
                os.Exit(1)
            }
            os.Exit(0)
        } else if (len(os.Args) == 2 {
            InitDb(os.Args[1], false)
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

