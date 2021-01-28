/**
* (C) 2020 Geotab Inc
*
* All files and artifacts in the repository at https://github.com/UlfBj/ccs-w3c-client
* are licensed under the provisions of the license provided by the LICENSE file in this repository.
*
**/

package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"fmt"
	"time"
)

var gen2Url string
var ovdsUrl string
var thisVin string

type PathList struct {
	LeafPaths []string
}

var pathList PathList

func pathToUrl(path string) string {
	var url string = strings.Replace(path, ".", "/", -1)
	return "/" + url
}

func jsonToStructList(jsonList string) int {
	err := json.Unmarshal([]byte(jsonList), &pathList)
	if err != nil {
		fmt.Printf("Error unmarshal json=%s\n", err)
		pathList.LeafPaths = nil
		return 0
	}
	return len(pathList.LeafPaths)
}

func createListFromFile(fname string) int {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return 0
	}
	return jsonToStructList(string(data))
}

func saveListAsFile(fname string) {
	buf, err := json.Marshal(pathList)
	if err != nil {
		fmt.Printf("Error marshalling from file %s: %s\n", fname, err)
		return
	}

	err = ioutil.WriteFile(fname, buf, 0644)
	if err != nil {
		fmt.Printf("Error writing file %s: %s\n", fname, err)
		return
	}
}

func getGen2Response(path string) string {
	url := "http://" + gen2Url + ":8888" + pathToUrl(path)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("getGen2Response: Error creating request=%s.", err)
		return ""
	}

	// Set headers
	req.Header.Set("Access-Control-Allow-Origin", "*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", gen2Url+":8888")

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("getGen2Response: Error in issuing request/response= %s ", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("getGen2Response: Error in reading response= %s ", err)
		return ""
	}

	return string(body)
}

func writeToOVDS(data string, path string) {
	/*        type DataPoint struct {
		Value string
		Timestamp string
	}
	jsonizedResponse := `{"datapoint":` + data + "}"
	fmt.Printf("writeToOVDS: Data= %s \n", jsonizedResponse)
	var dataPoint DataPoint
	err := json.Unmarshal([]byte(jsonizedResponse), &dataPoint)
	if err != nil {
		fmt.Printf("writeToOVDS: Error JSON decoding of data= %s \n", err)
		return
	}*/
	url := "http://" + ovdsUrl + ":8765/ovdsserver"
	fmt.Printf("writeToOVDS: data = %s\n", data)

	payload := `{"action":"set", "vin":"` + thisVin + `" , ` +  data[1:]
	fmt.Printf("writeToOVDS: request payload= %s \n", payload)

	req, err := http.NewRequest("POST", url, strings.NewReader(payload)) //bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("writeToOVDS: Error creating request= %s \n", err)
		return
	}

	// Set headers
	req.Header.Set("Access-Control-Allow-Origin", "*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", ovdsUrl+":8765")

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	_, err = client.Do(req)
	if err != nil {
		fmt.Printf("writeToOVDS: Error in issuing request/response= %s ", err)
		return
	}
	//	defer resp.Body.Close()

	/*	body, err := ioutil.ReadAll(resp.Body)   // TODO Handle error response
		if err != nil {
			fmt.Printf("writeToOVDS: Error in reading response= %s ", err)
			return
		}*/
}

func runList(elements int, sleepTime int) {
	for i := 0; i < elements; i++ {
		response := getGen2Response(pathList.LeafPaths[i])
		if (len(response) == 0) {
		    fmt.Printf("\nrunList: Cannot connect to server.\n")
		    os.Exit(-1)
		}
		writeToOVDS(response, pathList.LeafPaths[i])
	        time.Sleep((time.Duration)(sleepTime) * time.Millisecond)
	}
	fmt.Printf("\n****************** Iteration done ************************************\n")
}

func main() {

	if len(os.Args) != 5 {
		fmt.Printf("CCS client command line: ./client gen2-server-url OVDS-server-url vin list-iteration-time\n")
		os.Exit(1)
	}
	gen2Url = os.Args[1]
	ovdsUrl = os.Args[2]
	thisVin = os.Args[3]
	iterationPeriod, _ := strconv.Atoi(os.Args[4])

	if createListFromFile("vsspathlist.json") == 0 {
	    if createListFromFile("../vsspathlist.json") == 0 {
		fmt.Printf("Failed in creating list from vsspathlist.json\n")
		os.Exit(1)
	    }
	}

	fmt.Printf("Client starts to read from VISSv2 server, and write to OVDS server..\n")
	elements := len(pathList.LeafPaths)
	sleepTime := (iterationPeriod*1000-elements*30)/elements  // 30 = estimated time in msec for one roundtrip - get data from VISSv2 server, write data to OVDS
	if (sleepTime < 1) {
	    sleepTime = 1
	}
	for {
		runList(elements, sleepTime)
	}
}
