### Openssl

The CA Certificate and server certificate in this repo is the same and shared across all the components.  If you wanted to test this repo out with your own self-signed certs, you can use the following procedure to gerneate your own CA and server certificates.

- Setup the serial and index file for openssl
```
cd certs
mkdir new_certs
touch index.txt
echo 00 > serial
```

- Generate the CA certificate and key
```
openssl genrsa -out CA_key.pem 2048
openssl req -x509 -days 600 -new -nodes -key CA_key.pem -out CA_crt.pem -extensions v3_ca -config openssl.cnf    -subj "/C=US/ST=California/L=Mountain View/O=Google/OU=Enterprise/CN=MyCA"
```

- Edit `openssl.cnf` and set the  SNI values as needed

```
[alt_names]
DNS.1 = server.domain.com
DNS.2 = grpc.domain.com
DNS.3 = grpcweb.domain.com
DNS.4 = localhost
```

- Generate the server certificates
```
openssl genrsa -out server_key.pem 2048
openssl req -config openssl.cnf -days 400 -out server_csr.pem -key server_key.pem -new -sha256  -extensions v3_req  -subj "/C=US/ST=California/L=Mountain View/O=Google/OU=Enterprise/CN=server.domain.com"
openssl ca -config openssl.cnf -days 400 -notext  -in server_csr.pem   -out server_crt.pem
```

- Copy the `.pem` files to each folder (`frontend/`, `backend_grpc/`, `backend_envoy`).
- Edit `gke_config/fe-secret.yaml` and place the base64 encoded version of the server cert/key file as the tls key and cert.