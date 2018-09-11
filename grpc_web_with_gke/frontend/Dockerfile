FROM grpcweb/common as build-env

RUN  git clone https://github.com/grpc/grpc-web.git && \
     cd grpc-web && \
     make install-plugin
ADD . /app
WORKDIR /app
ADD package_build.json /app/package.json
RUN mkdir www
RUN protoc -I=.  echo/echo.proto --js_out=import_style=commonjs:www --proto_path=. --grpc-web_out=import_style=commonjs,mode=grpcwebtext:www
RUN npm install --silent && \
    npx webpack client.js


FROM node:8 as build-env-node
COPY --from=build-env /app/dist /app/dist
WORKDIR /app
ADD package_web.json /app/package.json
RUN npm  install --silent


FROM node:8
COPY --from=build-env-node /app /app
ADD app.js /app/app.js
ADD CA_crt.pem /app/CA_crt.pem
ADD server_crt.pem /app/server_crt.pem
ADD server_key.pem /app/server_key.pem
WORKDIR /app
EXPOSE  8000
CMD ["node", "app.js"]