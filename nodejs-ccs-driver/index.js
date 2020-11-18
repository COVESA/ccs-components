var db = require("./db.js")
var ws = require("./ws.js")

var index = 0

function getLoop(arg){
    ws.get(index)
    index = index + 1
}

function main() {
    console.log("Hello world");
    ws.w3cInit('localhost', 8080);
    db.initDb('localhost', 'admin', 'admin123', 'vssdatalake')
    setInterval(getLoop, 1000, 'Test');
}

if (require.main === module) {
    main();
}