
const express = require('express');
const https = require('https');
const app = express();
const morgan = require('morgan');
var fs = require('fs');

const port = 8000;

app.use(
  morgan('combined')
);

var winston = require('winston');
var logger = winston.createLogger({
  transports: [
    new (winston.transports.Console)({ level: 'info' })
  ]
 });

app.use('/dist', express.static('dist'))

app.get('/', (request, response) => {
  logger.info('Called /');
  response.send(`<html>

  <head>
      <title>gRPC-web</title>
      <script src="dist/main.js"></script>
  <script>
  </script>
  </head>
  <body>
      <p>
              <form id="rpcSubmit" name="rpcSubmit">
                  gRPC Message to send:<br>
                  <input type="text" name="rpcMessage" id="rpcMessage" value="Hi Sal"><br>
                  <input type="submit" value="Submit">
              </form>
      </p>
  <br/>
  <hr>
  
  <h3>Unary gRPC Response</h3>
  <div id="unary"></div>
  <br/>
  
  <h3>ServerStreaming gRPC Response</h3>
  <div id="streaming"></div>
  </body>
  
  </html>
  `);
})

app.get('/_ah/health', (request, response) => {
  response.send('ok');
})


var privateKey = fs.readFileSync( 'server_key.pem' );
var certificate = fs.readFileSync( 'server_crt.pem' );

https.createServer({
    key: privateKey,
    cert: certificate
}, app).listen(port);