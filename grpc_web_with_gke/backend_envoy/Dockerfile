FROM envoyproxy/envoy:latest

COPY envoy.yaml /etc/envoy/envoy.yaml

EXPOSE 18080


ADD CA_crt.pem /etc/envoy/CA_crt.pem
ADD server_crt.pem /etc/envoy/server_crt.pem
ADD server_key.pem /etc/envoy/server_key.pem

WORKDIR /etc/envoy
CMD /usr/local/bin/envoy -c /etc/envoy/envoy.yaml