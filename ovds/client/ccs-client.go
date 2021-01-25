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
		return -1
	}
	return 0
}

func createListFromFile(fname string) int {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return -1
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

func writeToOVDS(response string, path string) {
	/*        type DataPoint struct {
		Value string
		Timestamp string
	}
	jsonizedResponse := `{"datapoint":` + response + "}"
	fmt.Printf("writeToOVDS: Response= %s \n", jsonizedResponse)
	var dataPoint DataPoint
	err := json.Unmarshal([]byte(jsonizedResponse), &dataPoint)
	if err != nil {
		fmt.Printf("writeToOVDS: Error JSON decoding of response= %s \n", err)
		return
	}*/
	url := "http://" + ovdsUrl + ":8765/ovdsserver"
	fmt.Printf("writeToOVDS: response = %s\n", response)

	data := `{"action":"set", "vin":"` + thisVin + `" ,"path":"` + path + `", ` + response[1:]
	fmt.Printf("writeToOVDS: request payload= %s \n", data)

	req, err := http.NewRequest("POST", url, strings.NewReader(data)) //bytes.NewBuffer(data))
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

func runList(trimList bool) {
	elements := len(pathList.LeafPaths)
	for i := 0; i < elements; i++ {
		response := getGen2Response(pathList.LeafPaths[i])
		if (len(response) == 0) {
		    fmt.Printf("\nrunList: Cannot connect to server.\n")
		    os.Exit(-1)
		}
		if (strings.Contains(response, "error") == true) {
			fmt.Printf("runList: Error in response= %s ", response)
			if trimList {
				copy(pathList.LeafPaths[i:], pathList.LeafPaths[i+1:])
				pathList.LeafPaths = pathList.LeafPaths[:len(pathList.LeafPaths)-1]
			}
		} else {
			writeToOVDS(response, pathList.LeafPaths[i])
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func main() {

	if len(os.Args) != 5 {  // 7-> 5;  2->1 3->2 5->3 6->4  (1 och 4 deletas) 
//		fmt.Printf("CCS client command line: ./client pathlist-filename gen2-server-url OVDS-server-url vss-tree-filename vin sleeptime\n")
		fmt.Printf("CCS client command line: ./client gen2-server-url OVDS-server-url vin sleeptime\n")
		os.Exit(1)
	}
	gen2Url = os.Args[1]
	ovdsUrl = os.Args[2]
	thisVin = os.Args[3]
	sleep, _ := strconv.Atoi(os.Args[4])

	if createListFromFile("vsspathlist.json") != 0 {
	    if createListFromFile("../vsspathlist.json") != 0 {
		fmt.Printf("Failed in creating list from vsspathlist.json\n")
		os.Exit(1)
	    }
	}

	fmt.Printf("Client starts to read from VISSv2 server, and write to OVDS server..\n")
	runList(true)
	saveListAsFile("vsspathlist.json")
	fmt.Printf("Client saved in vsspathlist.json the paths that responded without error.\nClient will continue after %d secs of sleep..\n", sleep)

	for {
		time.Sleep(time.Duration(sleep) * time.Second)
		runList(false)
	}
}
