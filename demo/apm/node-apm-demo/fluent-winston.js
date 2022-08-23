var winston = require('winston');
var config = {
  host: process.env.MELT_LOGS_TARGET,
  port: 8006,
  timeout: 3.0,
  requireAckResponse: true // Add this option to wait response from Fluentd certainly
};
var fluentTransport = require('fluent-logger').support.winstonTransport();
var fluent = new fluentTransport('test-tag', config);
var logger = winston.createLogger({
  transports: [fluent, new (winston.transports.Console)()]
});
 
logger.on('flush', () => {
  console.log("flush");
})
 
logger.on('finish', () => {
  console.log("finish");
  fluent.sender.end("end", {}, () => {})
});

function yourFunction(){
    // do whatever you like here
    // send an event record with 'tag.label'
    console.log("adding winston logs .... @ ", Date.now());
    logger.info('Info Log Sample');
    logger.warn('Warning log sample');
    logger.error('Error log sample');
    logger.debug('Debug log sample');
    setTimeout(yourFunction, 5000);
}

yourFunction();


