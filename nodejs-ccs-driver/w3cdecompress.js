var uuiddata = require("./uuiddata.js")
var vsspathlist = require("./vsspathlist.js")

module.exports = {
    decompressMessage: function(message) {
        index = 0
        var finalMsg = ""
        var uuidmap = vsspathlist.uuidlist["LeafPaths"];
        while (index < message.length) {
            charmsg = message[index]
            // console.log("message[" + index + "] " + charmsg);
            if (charmsg > 127) {
                var testmsg = uuiddata.keywordlist["keywords"][charmsg - 128]
                index = index + 1
                //keywords
                if (charmsg - 128 == 3) { //timestamp                
                    const todayYr = new Date()
                    timestamp = parseInt(Math.floor(todayYr.getFullYear() / 10) * 10)
                    charmsg = message[index]
                    var byte1 = charmsg
                    charmsg = message[index + 1]
                    var byte2 = charmsg
                    charmsg = message[index + 2]
                    var byte3 = charmsg
                    charmsg = message[index + 3]
                    var byte4 = charmsg
                    
                    var yy = parseInt((byte1 & 0b00111100) >>> 2)
                    timestamp += parseInt(yy)
                    timestamp += '-'
                    
                    var mm = ((byte1 & 0b00000011) << 2) | ((byte2 & 0b11000000) >>> 6)
                    timestamp += mm.toString().padStart(2, '0')
                    timestamp += '-'
                    
                    var dd = parseInt((byte2 & 0b00111110) >>> 1)
                    timestamp += parseInt(dd).toString().padStart(2, '0')
                    timestamp += 'T'
                    var hh = ((byte2 & 0b00000001) << 4) | ((byte3 & 0b11110000) >>> 4)
                    timestamp += parseInt(hh).toString().padStart(2, '0')
                    timestamp += ':'
                    
                    var MM = ((byte3 & 0b00001111) << 2) | ((byte4 & 0b11000000) >>> 6)
                    timestamp += parseInt(MM).toString().padStart(2, '0')
                    timestamp += ':'
                    
                    var ss = parseInt(byte4 & 0b00111111)
                    timestamp += parseInt(ss).toString().padStart(2, '0')
                    timestamp += 'Z'
                    finalMsg = finalMsg + '"' + testmsg + '":"' + timestamp + '"'
                    index = index + 4
                    
                    // console.log("byte1 " + byte1 )
                    // console.log("byte2 " + byte2 )
                    // console.log("byte3 " + byte3 )
                    // console.log("byte4 " + byte4 )
                    // console.log("Timestamp assigned " + timestamp )
                    // console.log("Timestamp assigned " + timestamp )
                    // console.log("Timestamp assigned " + timestamp )
                    // console.log("Hour  " + hh )
                    // console.log("Timestamp assigned " + timestamp )
                    // console.log("bit1  " + ((byte3 & 0b00001111) <<  2) )
                    // console.log("bit2  " + ((byte4 & 0b11000000) >>> 6) )
                    // console.log("Timestamp assigned " + timestamp )
                    // console.log("Timestamp assigned " + timestamp )
                    // console.log("Timestamp assigned " + timestamp)
                    continue
                } else if (charmsg - 128 == 4) { //path
                    var uuidindex = 0
                    uuidindex += (uuidindex << 8) + message[index + 1]
                    uuidindex += message[index + 1]
                    // console.log("UUID INDEX = " + uuidindex + "value = " + uuidmap[uuidindex])
                    finalMsg = finalMsg + '"' + testmsg + '":"' + uuidmap[uuidindex] + '"'
                    index = index + 2
                    continue
                } else if (charmsg - 128 > 7 && charmsg - 128 < 13) { //req types
                    finalMsg = finalMsg + '"' + testmsg + '"'
                    continue
                } else if (charmsg - 128 > 12 && charmsg - 128 < 23) { //values
                    var numvals = 0
                    finalMsg += '"'
                    if (testmsg.startsWith("n")) {
                        finalMsg += '-'
                    }
                    if (testmsg.endsWith("int8")) {
                        numvals = 1
                    }
                    if (testmsg.endsWith("int16")) {
                        numvals = 2
                    }
                    if (testmsg.endsWith("int24")) {
                        numvals = 3
                    }
                    if (testmsg.endsWith("int32")) {
                        numvals = 4
                    }
                    if (testmsg.endsWith("bool")) {
                        numvals = 1
                    }
                    if (testmsg.endsWith("float")) {
                        numvals = 4
                    }
                    if (numvals != 0) {
                        if (testmsg.endsWith("float")) {
                            var buf = new ArrayBuffer(numvals);
                            var view = new DataView(buf);
                            for (var i = 0; i < numvals; i++) {
                                view.setUint8(i, message[index + i])
                                // console.log("Byte Float value = " + message[index + i])
                            }
                            index = index + numvals
                            var num = view.getFloat32(0, true);
                            // console.log("Float value = " + Number(num))
                            finalMsg += num
                        }else{
                            var realval = 0;
                            for (var i = 0; i < numvals; i++) {
                                realval = (realval << 8) + message[index + i]
                            }
                            index = index + numvals
                            if (testmsg.endsWith("bool")) {
                                if (realval == 0) {
                                    finalMsg += "false"
                                } else {
                                    finalMsg += "true"
                                }
    
                            }else {
                                finalMsg += realval.toString()
                            }
                        }
                        numvals = 0
                    } else {
                        var i = 0
                        while (message[index + i]) {
                            console.log(message[index + i])
                            i = i + 1
                        }
                        index = index + i
                    }
                    finalMsg += '"'
                    continue;
                } else {
                    if (charmsg - 128 != 23) {
                        finalMsg = finalMsg + '"' + testmsg + '":'
                        //console.log("Adding " + testmsg)
                    }
                }
            } else {
                if (message[index] == 91){
                    finalMsg += "["    
                }
                else if (message[index] == 93){
                    finalMsg += "]"    
                }else{
                    finalMsg += message[index]
                }
                //console.log("Adding " + message[index])
                index = index + 1
            }
        }
        finalMsg = finalMsg.replace(/\"\"/g, '\",\"')
                            .replace(/\}\{/g, '\},\{')
        // console.log(finalMsg)
        return '{' + finalMsg + '}'
    }
}
