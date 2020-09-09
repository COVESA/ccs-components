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
	"unsafe"

	"fmt"
	"time"
)

// #include <stdlib.h>
// #include <stdio.h>
// #include <stdbool.h>
// #include "vssparserutilities.h"
import "C"

var VSSTreeRoot C.long

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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func jsonToStructList(jsonList string, elements int) int {
	err := json.Unmarshal([]byte(jsonList), &pathList) //exclude curly braces when only one key-value pair
	if err != nil {
		fmt.Printf("Error unmarshal json=%s\n", err)
		return -1
	}
	/*    var listElement ListElement
	frontBoundary := -1
	for i := 0 ; i < elements ; i++ {
		frontBoundary = strings.Index(jsonList[frontBoundary+1:len(jsonList)], "{")
		fmt.Printf("Inside jsonToStructList. frontBoundary=%d\n", frontBoundary)
		if (frontBoundary == -1) {
			return -1
		}
		endBoundary := strings.Index(jsonList[frontBoundary:len(jsonList)], "}")
		err := json.Unmarshal([]byte(jsonList[frontBoundary+1:endBoundary+1]), &listElement)  //exclude curly braces when only one key-value pair
		if err != nil {
			fmt.Printf("Error unmarshal json %s ; %s\n", jsonList[frontBoundary+1:endBoundary+1], err)
			return -1
		}
		nodeList = append(nodeList, listElement)
	}*/
	return 0
}

func createListFromFile(fname string) int {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", fname, err)
		return -1
	}
	elements := strings.Count(string(data), "{")
	fmt.Printf("Before jsonToStructList. elements=%d\n", elements)
	return jsonToStructList(string(data), elements)
}

func createListFromTree(treeFname string, listFname string) int {
	// call int VSSGetLeafNodesList(long rootNode, char* leafNodeList);
	ctreeFname := C.CString(treeFname)
	vssRoot := C.VSSReadTree(ctreeFname)
	C.free(unsafe.Pointer(ctreeFname))
	clistFname := C.CString(listFname)
	//    var matches C.int =
	C.VSSGetLeafNodesList(vssRoot, clistFname)
	C.free(unsafe.Pointer(clistFname))
	return createListFromFile(listFname)
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
		if strings.Contains(response, "error") {
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

	if len(os.Args) != 7 {
		fmt.Printf("CCS client command line: ./client pathlist-filename gen2-server-url OVDS-server-url vss-tree-filename vin sleeptime\n")
		os.Exit(1)
	}
	sleep, _ := strconv.Atoi(os.Args[6])
	gen2Url = os.Args[2]
	ovdsUrl = os.Args[3]
	thisVin = os.Args[5]
	if fileExists(os.Args[1]) {
		if createListFromFile(os.Args[1]) != 0 {
			fmt.Printf("Failed in createListFromFile\n")
			os.Exit(1)
		}
	} else {
		if createListFromTree(os.Args[4], os.Args[1]) != 0 {
			fmt.Printf("Failed in createListFromTree\n")
			os.Exit(1)
		}
		fmt.Printf("After createListFromTree\n")
		runList(true)
		saveListAsFile(os.Args[1])
	}
	for {
		runList(false)
		time.Sleep(time.Duration(sleep) * time.Second)
	}
}
