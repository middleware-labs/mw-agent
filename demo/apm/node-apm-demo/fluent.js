const FluentClient = require("@fluent-org/logger").FluentClient;
const logger = new FluentClient("tag_prefix", {
    socket: {
        host: "localhost",
        port: 8006,
        timeout: 3000, // 3 seconds
    }
});

function yourFunction(){
    // do whatever you like here
    // send an event record with 'tag.label'
    logger.emit('label', {record: 'this is a log @ ' + Date.now()});
    setTimeout(yourFunction, 5000);
}

yourFunction();