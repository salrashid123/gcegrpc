const {EchoRequest} = require('./src/echo/echo_pb.js');
const {EchoServerClient} = require('./src/echo/echo_grpc_web_pb.js');

const grpc = {};
grpc.web = require('grpc-web');

var echoService = new EchoServerClient('https://server.domain.com:8080', null, null);

  
 function onLoad() {
    var self = this;
    document.getElementById('rpcSubmit').addEventListener('submit', function(evt){
      evt.preventDefault();

    var val = document.getElementById('rpcMessage').value;

    var unary_request = new EchoRequest();
    unary_request.setName('Unary Request for Message: '+ val);
  
    echoService.sayHello(unary_request, {"custom-header-1": "value1"},  function(err, response) {
        if (err) {
          alert('Error calling gRPC sayHello: '+err.code+' "'+  err.message+'"');
        } else {
          setTimeout(function () {
              console.log(response.getMessage());
              var x = document.getElementById("unary");
              x.innerHTML = response.getMessage();
        }, 500);
    }
    });
  
    var streamRequest = new EchoRequest();
    streamRequest.setName('Streaming Request for Message: ' + val);
  
    var stream = echoService.sayHelloStream(streamRequest, {"custom-header-1": "value1"});
  
  
      stream.on('data', function(response) {
        console.log(response.getMessage());
        var x = document.getElementById("streaming");
        x.innerHTML = x.innerHTML + "<br />" + response.getMessage();
  
      });
      stream.on('status', function(status) {
        if (status.metadata) {
          console.log("Received Streaming metadata");
          console.log(status.metadata);
        }
      });
      stream.on('error', function(err) {
        alert('Error codeStreaming : '+err.code+' "'+  err.message+'"');
      });
      stream.on('end', function() {
        console.log("stream end signal received");
      });

  });

  };

  
  
  window.addEventListener('load', function () {
    onLoad();
  }, false);


