/**
* (C) 2021 Geotab
*
* All files and artifacts in the repository at https://github.com/MEAE-GOT/W3C_VehicleSignalInterfaceImpl
* are licensed under the provisions of the license provided by the LICENSE file in this repository.
*
**/
package main

import (
	"os"
	"fmt"
	"time"
	"strings"
	"strconv"
	"net/http"
	"encoding/json"
	"io/ioutil"

  MQTT  "github.com/eclipse/paho.mqtt.golang"
)

var uniqueTopicName string
var ovdsChan chan string

var vissv2Url string
var ovdsUrl string
var thisVin string
var subscriptionInterval string

type PathList struct {
	LeafPaths []string
}

var pathList PathList

func getBrokerSocket(isSecure bool) string {
//	FVTAddr := os.Getenv("MQTT_BROKER_ADDR")
        FVTAddr := "test.mosquitto.org"   // does it work for testing?
	if FVTAddr == "" {
		FVTAddr = "127.0.0.1"
	}
	if (isSecure == true) {
	    return "ssl://" + FVTAddr + ":8883"
        } 
	return "tcp://" + FVTAddr + ":1883"
}

var publishHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
//    fmt.Printf("Topic=%s\n", msg.Topic())
//    fmt.Printf("Payload=%s\n", string(msg.Payload()))
    ovdsChan <- string(msg.Payload())
}

func mqttSubscribe(brokerSocket string, topic string) {
    fmt.Printf("mqttSubscribe:Topic=%s\n", topic)
    opts := MQTT.NewClientOptions().AddBroker(brokerSocket)
    opts.SetDefaultPublishHandler(publishHandler)

    c := MQTT.NewClient(opts)
    if token := c.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }
    if token := c.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
        fmt.Println(token.Error())
        os.Exit(1)
    }
}

func publishMessage(brokerSocket string , topic string, payload string) {   
    fmt.Printf("publishMessage:Topic=%s, Payload=%s\n", topic, payload)
    opts := MQTT.NewClientOptions().AddBroker(brokerSocket)

    c := MQTT.NewClient(opts)
    if token := c.Connect(); token.Wait() && token.Error() != nil {
        fmt.Println(token.Error())
        os.Exit(1)
    }
    token := c.Publish(topic, 0, false, payload)
    token.Wait()
    c.Disconnect(250)
}

func subscribeVissV2Response(brokerSocket string) {
    mqttSubscribe(brokerSocket, uniqueTopicName)
}

func publishVissV2Request(brokerSocket string, path string, reqId int) {
    request := `{"action":"subscribe","path":"` + path + `","filter":{"op-type":"capture","op-value":"time-based","op-extra":{"period":"` + 
                subscriptionInterval + `"}},"requestId":"` + strconv.Itoa(reqId) + `"}`
    payload := `{"topic":"` + uniqueTopicName + `", "request":` + request + `}`
    publishMessage(brokerSocket, "/" + thisVin + "/Vehicle", payload)
}

func jsonToMap(request string, rMap *map[string]interface{}) {
	decoder := json.NewDecoder(strings.NewReader(request))
	err := decoder.Decode(rMap)
	if err != nil {
		fmt.Printf("jsonToMap: JSON decode failed for request:%s, err=%s\n", request, err)
	}
}

func extractMessage(message string) (string, string, string) { // assuming format: {"action": "subscription", "subscriptionId": "A", “data”:{“path”:”B”, “dp”:{“value”:”C”, “ts”:”D”}}}
    var msgMap = make(map[string]interface{})
    jsonToMap(message, &msgMap)
    data, _ := json.Marshal(msgMap["data"])

    jsonToMap(string(data), &msgMap)
    path := msgMap["path"].(string)
    dp, _ := json.Marshal(msgMap["dp"])

    jsonToMap(string(dp), &msgMap)
    value := msgMap["value"].(string)
    ts := msgMap["ts"].(string)
fmt.Printf("path=%s, value=%s, ts=%s\n", path, value, ts)
    return path, value, ts
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

func writeToOvds(ovdsChan chan string) {
	for {
	    message := <- ovdsChan
	    fmt.Printf("writeToOVDS: message = %s\n", message)
	    if (strings.Contains(message, "data") == false) {
	        continue
	    }
	    path, value, timeStamp := extractMessage(message)

	    payload := `{"action":"set", "vin":"` + thisVin + `", "path":"` +  path + `", "value":"` + value + `", "timestamp":"` + timeStamp + `"}`
	    fmt.Printf("writeToOVDS: request payload= %s \n", payload)

	    url := "http://" + ovdsUrl + ":8765/ovdsserver"
	    req, err := http.NewRequest("POST", url, strings.NewReader(payload))
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
	}
}


func main() {
	if len(os.Args) != 6 {
		fmt.Printf("CCS MQTT client command line: ./mqtt_client gen2-server-url OVDS-server-url vin interval-time topic-name\n")
		os.Exit(1)
	}
	vissv2Url = os.Args[1]
	ovdsUrl = os.Args[2]
	thisVin = os.Args[3]
	subscriptionInterval = os.Args[4]
	uniqueTopicName = os.Args[5]
	_, err := strconv.Atoi(subscriptionInterval)
	if (err != nil) {
		fmt.Printf("Interval time must be an integer.\n")
		os.Exit(1)
	}

	if createListFromFile("vsspathlist.json") == 0 {
	    if createListFromFile("../vsspathlist.json") == 0 {
		fmt.Printf("Failed in creating list from vsspathlist.json\n")
		os.Exit(1)
	    }
	}

	ovdsChan = make(chan string)
	go writeToOvds(ovdsChan)

	brokerSocket := getBrokerSocket(false)
	subscribeVissV2Response(brokerSocket)

	elements := len(pathList.LeafPaths)
	for i := 0; i < elements; i++ {
		publishVissV2Request(brokerSocket, pathList.LeafPaths[i], i)
	        time.Sleep(27 * time.Millisecond)
	}
	for {
	        time.Sleep(1000 * time.Millisecond)  // to keep alive...
	}
}
