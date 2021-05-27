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
    "net/http"
    "os"
    "strings"
    "strconv"
    "encoding/json"
    "time"
    "fmt"
    
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var stateStorageError bool = false


type PathList struct {
	LeafPaths []string
}

var pathList PathList

type DataPoint struct {
        Value string  `json:"value"`
	Timestamp string  `json:"ts"`
}

type DataPackage struct {
        Path string  `json:"path"`
	Datapoints []DataPoint `json:"dp"`
}

type SampleList struct {
	DataPack DataPackage `json:"datapackage"`
}

var sampleList SampleList

var vehicleVin string
var ovdsUrl string
var statestorageFname string

const RINGSIZE = 25
type RingElement struct {
	Value string
	Timestamp string
}

type RingBuffer struct {
    RingElem [RINGSIZE]RingElement
    Head int
    Tail int
}

var ringArray []RingBuffer
var latestTimestamp []string

func InitRingArray(numOfRings int) {
    ringArray = make([]RingBuffer, numOfRings)
    for i := 0 ; i < numOfRings ; i++ {
        ringArray[i].Head = 0
        ringArray[i].Tail = 0
    }
}

func writeRing(ringIndex int, value string, timestamp string) {
//fmt.Printf("writeRing(%d): value=%s, ts=%s\n", ringIndex, value, timestamp)
    ringArray[ringIndex].RingElem[ringArray[ringIndex].Tail].Value = value
    ringArray[ringIndex].RingElem[ringArray[ringIndex].Tail].Timestamp = timestamp
    ringArray[ringIndex].Tail++
    if (ringArray[ringIndex].Tail == RINGSIZE) {
        ringArray[ringIndex].Tail = 0
    }
}

func readRing(ringIndex int) (string, string) {
    currentHead := ringArray[ringIndex].Head
    return ringArray[ringIndex].RingElem[currentHead].Value, ringArray[ringIndex].RingElem[currentHead].Timestamp
}

func popReadRing(ringIndex int) {
    ringArray[ringIndex].Head++
    if (ringArray[ringIndex].Head == RINGSIZE) {
        ringArray[ringIndex].Head = 0
    }
}

func getNumOfUnreadRingElements(ringIndex int) int {
    head := ringArray[ringIndex].Head
    tail := ringArray[ringIndex].Tail
    if (head > tail) {
        tail += RINGSIZE
    }
    return tail - head
}

func jsonToStructList(jsonList string, list interface{}) {
	err := json.Unmarshal([]byte(jsonList), list)
	if err != nil {
		fmt.Printf("Error unmarshal json=%s, err=%s\n", jsonList, err)
		return
	}
}

func createPathList(fname string) int {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", fname, err)
		return 0
	}
	jsonToStructList(string(data), &pathList)
	return len(pathList.LeafPaths)
}

func initTimeStamps(numOfPaths int) {
    latestTimestamp = make([]string, numOfPaths)
    for i := 0 ; i < numOfPaths ; i++ {
        latestTimestamp[i] = "2000-01-01T00:00:00Z"
    }
}

func getOvdsSamples(path string, timestamp string, numOfSamples int) string {
	url := "http://" + ovdsUrl + ":8766/ovdsserver"
        data := `{"action":"get", "vin": "` + vehicleVin + `", "path":"` + path + `", "from":"` + timestamp + `", "maxsamples":"` + strconv.Itoa(numOfSamples) + `"}`
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		fmt.Printf("getOvdsSamples: Error creating request= %s \n", err)
		return ""
	}

	// Set headers
	req.Header.Set("Access-Control-Allow-Origin", "*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", ovdsUrl+":8766")

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("getOvdsSamples: Error in issuing request/response= %s ", err)
		return ""
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("getOvdsSamples: Error in reading response= %s ", err)
		return ""
	}

	return string(body)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func fillRings(ringArray []RingBuffer, numOfPaths int) {
    for i := 0 ; i < numOfPaths ; i++ {
        numOfFreeElements := RINGSIZE - getNumOfUnreadRingElements(i)
        response := getOvdsSamples(pathList.LeafPaths[i], latestTimestamp[i], numOfFreeElements)
        if (len(response) == 0 || strings.Contains(response, "error") == true) {
            continue
        }
        jsonToStructList(response, &sampleList)
        numOfFreeElements = len(sampleList.DataPack.Datapoints)
fmt.Printf("fillRings: i=%d, numOfFreeElements=%d\n", i, numOfFreeElements)
        for j := 0; j < numOfFreeElements ; j++ {
            writeRing(i, sampleList.DataPack.Datapoints[j].Value, sampleList.DataPack.Datapoints[j].Timestamp)
            latestTimestamp[i] = sampleList.DataPack.Datapoints[j].Timestamp
        }
    }
}

func convertFromIsoTime(isoTime string) (time.Time, error) {
        correctIsoTime := strings.ReplaceAll(isoTime, ".", ":")  // Adapter uses YYYY-MM-DDThh.mm.ssZ
	time, err := time.Parse(time.RFC3339, correctIsoTime)
	return time, err
}

func getCurrentUtcTime() time.Time {
    return time.Now().UTC()
}

func getOldestTimestamp(ringArray []RingBuffer, numOfPaths int) time.Time {
    oldestTime := getCurrentUtcTime()  // ts must be older then current time
    oldestTimeOriginal := oldestTime
    for i := 0 ; i < numOfPaths ; i++ {   // check the next to be sent in each ring, select the "oldest"
        _, timestamp := readRing(i)
        if (len(timestamp) == 0) {
            continue
        }
fmt.Printf("getOldestTimestamp: readRing(%d)=%s\n", i, timestamp)
        ts, err := convertFromIsoTime(timestamp)
        if (err == nil) {
            if (ts.Before(oldestTime)) {
                oldestTime = ts
            }
        } else {
            fmt.Printf("convertFromIsoTime: error for timestamp=%s\n", timestamp)
        }
    }
fmt.Printf("getOldestTimestamp: oldest-ts=%s\n", oldestTime)
    if (oldestTime.Equal(oldestTimeOriginal) == true) {
        fmt.Printf("Live simulator: Done reading the OVDS database.\nGoodbye.\n")
        os.Exit(0)
    }
    return oldestTime
}

func InitDb(dbFile string) *sql.DB {
        var db *sql.DB
        var err error
	if (fileExists(dbFile)) {
		db, err = sql.Open("sqlite3", dbFile)
               if (err != nil) {
		    fmt.Printf("\nDB %s failed to open.\n", dbFile)
		    os.Exit(1)
		    return nil
               }
	} else {
		fmt.Printf("\nDB %s must exist, or else the statestorage manager must be started also with a filename to a JSON pathlist.\nSee README\n", dbFile)
		os.Exit(1)
		return nil
	}
        return db
}

func writeToStatestorage(path string, value string, timestamp string) {
	stmt, err := db.Prepare("UPDATE VSS_MAP SET value=?, timestamp=? WHERE `path`=?")
	if (err != nil) {
		fmt.Printf("Db prepare update failed, err=%s", err)
		stateStorageError = true
		return
	}

	_, err = stmt.Exec(value, timestamp, path)
	if err != nil {
		fmt.Printf("Db exec update failed, err=%s", err)
		stateStorageError = true
		return
	}
        fmt.Printf("writeToStatestorage:  value=%s, ts=%s, path=%s\n", value, timestamp, path)
}

func pushRingSamples(ringArray []RingBuffer, numOfPaths int, currentTime time.Time) int {
    minNumOfUnread := RINGSIZE
    for i := 0 ; i < numOfPaths ; i++ {
        value, timestamp := readRing(i)
        ts, err := convertFromIsoTime(timestamp)
        if (err == nil) {
            if (ts.Before(currentTime)) {
                writeToStatestorage(pathList.LeafPaths[i], value, timestamp)
                popReadRing(i)
                if (getNumOfUnreadRingElements(i) < minNumOfUnread) {
                   minNumOfUnread = getNumOfUnreadRingElements(i)
                }
            }
        }
    }
    return minNumOfUnread
}

func main() {
    if len(os.Args) != 4 {
        fmt.Printf("livesim command line: ./livesim VIN OVDS-server-url statestorage-db-filename\n")
	os.Exit(1)
    }
    fmt.Printf("Remember to start the OVDS server to run with livesim.\n")
    vehicleVin = os.Args[1]
    ovdsUrl = os.Args[2]
    statestorageFname = os.Args[3]
    db = InitDb(statestorageFname)
    numOfPaths := createPathList("vsspathlist.json")
    initTimeStamps(numOfPaths)
    InitRingArray(numOfPaths)
    fillRings(ringArray, numOfPaths)
    timeDiff := getOldestTimestamp(ringArray, numOfPaths).Sub(getCurrentUtcTime())
    for {
        currentTime := getCurrentUtcTime().Add(timeDiff)
        minFill := pushRingSamples(ringArray, numOfPaths, currentTime)
        if (minFill == 0) { // could also be done if sleep time > x
            fillRings(ringArray, numOfPaths)
        }
        currentTime = getCurrentUtcTime().Add(timeDiff)
        wakeUp := getOldestTimestamp(ringArray, numOfPaths).Sub(currentTime)
        fmt.Printf("Sleep for: %s\n", wakeUp)
        if (wakeUp < 0) {
            if (stateStorageError == false) {
                fmt.Printf("Live simulator: Done reading the OVDS database.\nGoodbye.\n")
            } else {
                fmt.Printf("Live simulator: Error writing to the %s database.\nGoodbye.\n", statestorageFname)
            }
            break
        }
        time.Sleep(wakeUp)
    }
}

