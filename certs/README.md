### Openssl

The CA Certificate and server certificate in this repo is the same and shared across all the components.  If you wanted to test this repo out with your own self-signed certs, you can use the following procedure to gerneate your own CA and server certificates.



- Edit `openssl.cnf` and set the  SNI values as needed

```bash
[ server_ext ]
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = server.domain.com
DNS.2 = grpc.domain.com
DNS.3 = be-srv
DNS.4 = be-srv.default.svc.cluster.local
DNS.5 = be-srv-lb
DNS.6 = be-srv-lb.default.svc.cluster.local
DNS.7 = grpc.domain.com
DNS.8 = grpcweb.domain.com
IP.1 = 127.0.0.1
```

- Generate the server certificates

```bash
openssl genrsa -out server_key.pem 2048
openssl req -config openssl.cnf  -out server_csr.pem -key server_key.pem -new -sha256  -extensions server_ext  -subj "/C=US/ST=California/L=Mountain View/O=Google/OU=Enterprise/CN=server.domain.com"
openssl ca -config openssl.cnf  -notext  -in server_csr.pem   -out server_crt.pem
```
