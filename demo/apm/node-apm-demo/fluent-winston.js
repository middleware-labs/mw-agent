var winston = require('winston');
var config = {
  host: process.env.MW_LOGS_TARGET,
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
    
    let random = Math.random();

    if (random < 0.9) {
      logger.info('Info Log Sample');
      logger.info('Info Account created');
     
      logger.debug('Debug memory usage is 300 MB');
      logger.info('Info telemetry data is being collected');
      logger.info('Info package.json is already cached');
    }
    
    if (random < 0.7) {
      logger.warn('Warning log sample');
      logger.warn('Warning record does not look like TLS handshake');
      logger.warn('Warning Could not read all the files');
     
      logger.info('Info New version released');
      logger.info('Info adding new data');
      logger.debug('Debug auth handshake successful');
      logger.error('Error address not found');
      logger.warn('Warning gitignore file missing');
      logger.warn('Warning Memory usage is over 500MB');
      logger.warn('Warning server timeout');
    }

    if (random < 0.1) {
      logger.debug('Debug running collector');
    }

    if (random < 0.6) {
      logger.debug('Debug printing all values');
      logger.debug('Debug memory usage is 400 MB');
    }

    if (random < 0.5) {
      logger.error('Error log sample');
      logger.error('Error authentication failed');
      logger.warn('Warning CPU usage is over 80%');
      logger.error('Error file not found');
      logger.info('Info Agent Installed');
      logger.error('Error docker build failed, dependency missing');
      logger.error('Error application crashed due to unhandled exception');
      logger.error('Error an error with long text to verify the design for long text, should be handled by ellipsis in design');
    }

    if (random < 0.3) {
      logger.debug('Debug log sample');
      
      logger.debug('Debug memory usage is 500 MB');
      logger.debug('Debug memory usage is 600 MB');
    }
    
    setTimeout(yourFunction, 5000);
}

yourFunction();


