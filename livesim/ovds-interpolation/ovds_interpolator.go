/**
* (C) 2021 Geotab Inc
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
)

var oldOvdsIpAdr string  //IP address to OVDS server from which data is read
var newOvdsIpAdr string  //IP address to OVDS server to which data is written
var oldVin string //VIN used in existing OVDS
var newVin string //VIN used in new OVDS

var createSawTooth bool = false
var adapterFix bool = false

type PathList struct {
	LeafPaths []string
}

var pathList PathList

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

func fixAdapter(index int) {
    value := ""
    ts := "2000-01-01T21:00:00Z"  // should be older than anything that will be read
    for {
        value, ts = readDataPoint(pathList.LeafPaths[index], ts)
        if (value == "") {
            break
        }
        ts = strings.Replace(ts, ".", ":",-1)  // fix for the timestamp format from the Adapter 
        writeDataPoint(pathList.LeafPaths[index], value, ts)
    }
}

func doInterpolate(index int) {
    oldValue, oldTs := readDataPoint(pathList.LeafPaths[index], "2000-01-01T21:00:00Z") // should be older than anything that will be found in the DB
    compressedDps := 0
    totalExpandedDps := 0
    for {
        expandedDps := 0
        fmt.Printf("Old value=%s, old ts=%s\n", oldValue, oldTs)
        newValue, newTs := readDataPoint(pathList.LeafPaths[index], oldTs)
        if (newValue == "") {
            break
        }
        fmt.Printf("New value=%s, new ts=%s\n", newValue, newTs)
//        if (compressedDps > 0) { // do not interpolate from the "older than anything"
            interpolateVal, interpolateTs := calculateInterpolations(newValue, newTs, oldValue, oldTs)
            for i := 0 ; i < len(interpolateVal) ; i++ {
                writeDataPoint(pathList.LeafPaths[index], interpolateVal[i], interpolateTs[i])
                expandedDps++
            }
//        }
        oldValue = newValue
        oldTs = newTs
        compressedDps++
        totalExpandedDps += expandedDps
    }
    fmt.Printf("For %s: Compressed datapoints=%d, Expanded data points=%d\n", pathList.LeafPaths[index], compressedDps, totalExpandedDps)
}

func calculateInterpolations(newValue string, newTs string, oldValue string, oldTs string) ([]string, []string) {
    var interpolateTs, interpolateVal []string
    firstTs := RFC3339ToUnix(oldTs)
    lastTs := RFC3339ToUnix(newTs)
    timeDiff := lastTs - firstTs // diff in seconds
    var i int64
    for i = 0 ; i < timeDiff ; i++ {
        interpolateTs = append(interpolateTs, UnixToRFC3339(firstTs + i))
        if (createSawTooth == true) {
            interpolateVal = append(interpolateVal, strconv.Itoa(int(i)))
        } else {
            interpolateVal = append(interpolateVal, interpolateValue(oldValue, newValue, i, timeDiff))
        }
        
    }
    return interpolateVal, interpolateTs
}

func interpolateValue(oldValue string, newValue string, index int64, period int64) string {
    oldValue = strings.Replace(oldValue, ",", ".", -1)
    newValue = strings.Replace(newValue, ",", ".", -1)
    floatVal := interpolateFloatValue(oldValue, newValue, index, period)
    switch (AnalyzeValueType(oldValue)) {
      case 3:
          return  strconv.FormatFloat(floatVal, 'f', -1, 64)
      case 1:
          delta := 0.5
          if (floatVal < 0) {
              delta = -0.5
          }
          intVal := (int)(floatVal + delta)
          return strconv.Itoa(intVal)
      default:
	  fmt.Printf("Type of value is not supported, value=%s\n", oldValue)
          return ""
    }
}

func interpolateFloatValue(oldValue string, newValue string, index int64, period int64) float64 {
    oldVal, err := strconv.ParseFloat(oldValue, 64)
    if err != nil {
	fmt.Printf("Failed to convert oldValue=%s to float err=%s\n", oldValue, err)
	return 0
    }
    newVal, err := strconv.ParseFloat(newValue, 64)
    if err != nil {
	fmt.Printf("Failed to convert newValue=%s to float err=%s\n", newValue, err)
	return 0
    }
    interpolVal := oldVal + (newVal - oldVal)/(float64)(period) * (float64)(index)
    return  interpolVal
}

func AnalyzeValueType(value string) int {
    _, err := strconv.Atoi(value)
    if (err == nil) {
        return 1  //int type
    }
    if (value == "true" || value == "false") {
        return 2 // bool type
    }
    if (isFloatType(value) == true) {
        return 3 // float type
    }
    return 0
}

func isFloatType(value string) bool {
    _, err := strconv.ParseFloat(value, 64)
    if err != nil {
        return false
    }
    return true
}

func RFC3339ToUnix(ts string) int64 {
        t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		fmt.Printf("Failed to convert RFC3339 time to Unix time err=%s", err)
		return -1
	}
        return t.Unix()
}

func UnixToRFC3339(t int64) string {
        ts := time.Unix(t, 0).Format(time.RFC3339)
        ts = ts[:strings.Index(ts, "+")] + "Z"
        return ts
}

func readDataPoint(path string, fromTs string) (string, string) {
	url := "http://" + oldOvdsIpAdr + ":8765/ovdsserver"
        request := `{"action":"get", "vin": "` + oldVin + `", "path":"`+ path + `", "from":"` + fromTs + `", "maxsamples":"1"}`
	fmt.Printf("readDataPoint: Request= %s \n", request)
	req, err := http.NewRequest("POST", url, strings.NewReader(request))
	if err != nil {
		fmt.Printf("readDataPoint: Error creating request= %s \n", err)
		return "", ""
	}

	// Set headers
	req.Header.Set("Access-Control-Allow-Origin", "*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", oldOvdsIpAdr+":8765")

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("readDataPoint: Error in issuing request/response= %s ", err)
		return "", ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("readDataPoint: Error in reading response= %s ", err)
		return "", ""
	}
	fmt.Printf("readDataPoint: Response= %s \n", string(body))
	// response from OVDS server: {"datapackage":{“path”:”X”, “dp”: [{“value”:”Y1”, “ts”:”Z1”}, …, {“value”:”Yn”, “ts”:”Zn”}]}}, or error response
	if (strings.Contains(string(body), "error")) {
		return "", ""
	}
	value, ts := extractFromDp(body)
	ts  = strings.Replace(ts, ".", ":", -1)// convert dot to colon, due bug in Adapter
	return value, ts
}

func extractFromDp(body []byte) (string, string) {
	var responseMap map[string]interface{}
	err := json.Unmarshal(body, &responseMap)
	if err != nil {
		fmt.Printf("Error unmarshal response=%s, err=%s\n", string(body), err)
		return "", ""
	}
       dataPackage, err := json.Marshal(responseMap["datapackage"])
	if err != nil {
		fmt.Printf("Error marshal datapackage, err=%s\n", err)
		return "", ""
	}
	err = json.Unmarshal(dataPackage, &responseMap)
	if err != nil {
		fmt.Printf("Error unmarshal dataPackage=%s, err=%s\n", string(dataPackage), err)
		return "", ""
	}
	value := ""
	ts := ""
        switch vv := responseMap["dp"].(type) {
          case []interface{}:
            fmt.Println(vv, "is an array:, len=",strconv.Itoa(len(vv)))
	    index := len(vv) - 1   // length is 1 or 2. If 2 select the last one.
	    dpObject := vv[index]
            switch vvv := dpObject.(type) {
              case map[string]interface{}:
                value = vvv["value"].(string)
                ts = vvv["ts"].(string)
            }
          default:
            fmt.Println(vv, "is of an unknown type")
        }
        return value, ts
}

func writeDataPoint(path string, value string, ts string) {
	url := "http://" + newOvdsIpAdr + ":8765/ovdsserver"
        request := `{"action":"set", "vin": "` + newVin + `", "path":"`+ path + `", "value":"` + value + `", "timestamp":"` + ts + `"}`
	req, err := http.NewRequest("POST", url, strings.NewReader(request))
	if err != nil {
		fmt.Printf("writeDataPoint: Error creating request= %s \n", err)
		return
	}

	// Set headers
	req.Header.Set("Access-Control-Allow-Origin", "*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", newOvdsIpAdr+":8765")

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("writeDataPoint: Error in issuing request/response= %s ", err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("writeDataPoint: Error in reading response= %s ", err)
		return
	}
	if (strings.Contains(string(body), "error") == true) {
	    fmt.Printf("writeDataPoint: Response for %s= %s \n", string(body), path)
	}
}

func main() {
    if (len(os.Args) != 5 && len(os.Args) != 6) {
        fmt.Printf("ovds-interpolator command line: ./ovds_interpolator old-ovds-server-ipadr old-vin new-ovds-server-ipadr new-vin\n")
	os.Exit(1)
    }
    oldOvdsIpAdr = os.Args[1]
    oldVin = os.Args[2]
    newOvdsIpAdr = os.Args[3]
    newVin = os.Args[4]
    if (len(os.Args) == 6 && os.Args[5] == "sawtooth") { // these command line params are undocumented...
        createSawTooth = true
    } else if (len(os.Args) == 6 && os.Args[5] == "adapterfix") {
        adapterFix = true
    }
    numOfPaths := createPathList("vsspathlist.json")
    for i := 0 ; i < numOfPaths ; i++ {
      if (adapterFix == true) {
        fixAdapter(i)
      } else {
        doInterpolate(i)
      }
    }
}


