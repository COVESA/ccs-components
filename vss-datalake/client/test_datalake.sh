#!/bin/bash
# datalake server command examples
# HTTP POST on server port no: 8765, path: datalakeserver

serverurl="http://localhost:8765/datalakeserver"
ct='Content-Type:application/json' 

# **** Request SET examples ****
set_example()
{

	echo '{"action":"set", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "value": "true", "timestamp":"2020-01-10T02:59:43.492Z"}'
	curl -d \
	'{"action":"set", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "value": "true", "timestamp":"2020-01-10T02:59:43.492Z"}' \
	-H "$ct" "$serverurl"
	sleep 1
	echo ""

	echo '{"action":"set", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "value": "false", "timestamp":"2020-01-11T02:59:43.492Z"}'
	curl -d \
	'{"action":"set", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "value": "false", "timestamp":"2020-01-11T02:59:43.492Z"}' \
	-H $ct $serverurl
	sleep 1
	echo ""

	echo '{"action":"set", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "value": "true", "timestamp":"2020-01-12T02:59:43.492Z"}'
	curl -d \
	'{"action":"set", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "value": "true", "timestamp":"2020-01-12T02:59:43.492Z"}' \
	-H $ct $serverurl
	sleep 1
	echo ""

	echo '{"action":"set", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "value": "false", "timestamp":"2020-04-01T02:59:43.492Z"}'
	curl -d \
	'{"action":"set", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "value": "false", "timestamp":"2020-04-01T02:59:43.492Z"}' \
	-H $ct $serverurl
	sleep 1
	echo ""
}

# **** Request GET examples ****
get_example()
{
	echo '{"action":"get", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "from":"2020-01-01T02:59:43.492750Z", "to":"2020-03-31T02:59:43.492750Z"}'
	curl -d \
	'{"action":"get", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "from":"2020-01-01T02:59:43.492750Z", "to":"2020-03-31T02:59:43.492750Z"}' \
	-H $ct $serverurl
	sleep 1
	echo ""

	echo '{"action":"get", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "from":"2020-01-09T02:59:43.492750Z"}'
	curl -d \
	'{"action":"get", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen", "from":"2020-01-09T02:59:43.492750Z"}' \
	-H $ct $serverurl
	sleep 1
	echo ""

	echo '{"action":"get", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen"}'
	curl -d \
	'{"action":"get", "vin": "GEO001", "path":"Vehicle/Cabin/Door/Row1/Left/IsOpen"}' \
	-H $ct $serverurl
	sleep 1
	echo ""

	echo '{action:get, vin: GEO001, path:Vehicle/Cabin/Door/Row1/Left/IsOpen}'
	curl -d '{action:get, vin: GEO001, path:Vehicle/Cabin/Door/Row1/Left/IsOpen}' \
	-H $ct $serverurl
	echo ""
}

set_example
#get_example
#get_example
#get_example
#get_example
#get_example


#**** GET response examples ****

#curl -d \
#'{"action":"getmetadata", "vin": "GEO001", "path":"Vehicle/Cabin/Door/", "depth": "2"}'\
#-H 'Content-Type: application/json' http://localhost:8765/datalakeserver
#{"datapackage":"{ \"path\":\"Vehicle/Cabin/Door/Row1/Left/IsOpen, \"datapoints\":\"{\"value\": \"false\", \"timestamp\": \"2020-04-01T02:59:43.492Z\"}}"}

#{"datapackage":"{ \"path\":\"Vehicle/Cabin/Door/Row1/Left/IsOpen, \"datapoints\":\"[{\"value\": \"true\", \"timestamp\": \"2020-01-10T02:59:43.492Z\"}, {\"value\": \"false\", \"timestamp\": \"2020-01-11T02:59:43.492Z\"}, {\"value\": \"true\", \"timestamp\": \"2020-01-12T02:59:43.492Z\"}, {\"value\": \"false\", \"timestamp\": \"2020-04-01T02:59:43.492Z\"}]}"}
