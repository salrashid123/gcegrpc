# GKE gRPC with TrafficDirector

Sample showing gRPC using Traffic Director Proxyless Loadbalancing

- [Traffic Director setup with Google Kubernetes Engine and proxyless gRPC services](https://cloud.google.com/traffic-director/docs/set-up-proxyless-gke)


This type of loadbalancing does not utilize an intermediate LB but rather each gRPC Client acquires a list of IP:port pairs for each GKE Pod where the service runs.
From there, each client is in charge of connecting _directly_ to the server's pod without a proxy.

In this mode

`client (on VM)` --> [acquire list of service NEG pods via xds] --> `gke pod with gRPC Service`


A couple of important notes:

* gRPC Server *must* have [gRPC HealthCheck](https://github.com/grpc/grpc/blob/master/doc/health-checking.md) enabled on the serving port
* gRPC Server *must not* use TLS (as of 3/20/24), [gRPC HealthChecks no not support TLS](https://cloud.google.com/load-balancing/docs/health-check-concepts#criteria-protocol-grpc) which means TrafficDirector does not work with TLS

**HOWEVER**, if you run the gRPC service over TLS, you can configure  can use an https healthcheck proxy instead.  For more information, see the `With TLS` section below

* [https://github.com/salrashid123/grpc_health_proxy](https://github.com/salrashid123/grpc_health_proxy)


### Setup

First configure a service account 

```bash
export PROJECT_ID=`gcloud config get-value core/project`
export PROJECT_NUMBER=`gcloud projects describe $PROJECT_ID --format="value(projectNumber)"`
export SERVICE_ACCOUNT_EMAIL=xds-svc-client@$PROJECT_ID.iam.gserviceaccount.com

gcloud iam service-accounts create xds-svc-client --display-name "XDS Client Service Account"


gcloud projects add-iam-policy-binding ${PROJECT_ID} \
   --member serviceAccount:${SERVICE_ACCOUNT_EMAIL} \
   --role roles/compute.networkViewer

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
   --member serviceAccount:${SERVICE_ACCOUNT_EMAIL} \
   --role roles/trafficdirector.client
```

### Without TLS

```bash
gcloud services enable trafficdirector.googleapis.com

gcloud container  clusters create cluster-grpc \
 --zone us-central1-a  --num-nodes 3 --enable-ip-alias \
 --tags=allow-health-checks

cd no_tls/
kubectl apply -f .  
kubectl get svcneg 
kubectl get svcneg fe-srv-td  -o yaml

# NEG_NAME is declared in fe-srv-td.yaml
gcloud compute network-endpoint-groups list 
export NEG_NAME="fe-srv-td"

gcloud compute health-checks create grpc grpc-gke-td-hc  --use-serving-port

gcloud compute firewall-rules create grpc-gke-allow-health-checks \
  --network default --action allow --direction INGRESS \
  --source-ranges 35.191.0.0/16,130.211.0.0/22 \
  --target-tags allow-health-checks \
  --rules tcp:50051

gcloud compute backend-services create grpc-gke-td-service \
   --global \
   --load-balancing-scheme=INTERNAL_SELF_MANAGED \
   --protocol=GRPC \
   --health-checks grpc-gke-td-hc

gcloud compute backend-services add-backend grpc-gke-td-service \
   --global \
   --network-endpoint-group $NEG_NAME \
   --network-endpoint-group-zone us-central1-a \
   --balancing-mode RATE \
   --max-rate-per-endpoint 5

gcloud compute url-maps create grpc-gke-url-map \
--default-service grpc-gke-td-service

gcloud compute url-maps add-path-matcher grpc-gke-url-map \
--default-service grpc-gke-td-service \
--path-matcher-name grpc-gke-path-matcher \
--new-hosts fe-srv-td:50051

gcloud compute target-grpc-proxies create grpc-gke-proxy \
--url-map grpc-gke-url-map \
--validate-for-proxyless

gcloud compute forwarding-rules create grpc-gke-forwarding-rule \
--global \
--load-balancing-scheme=INTERNAL_SELF_MANAGED \
--address=0.0.0.0 \
--target-grpc-proxy=grpc-gke-proxy \
--ports 50051 \
--network default
```

Wait a couple of mins and you should see is traffic directory detecting the backend services you deployed

Now create a VM 

```bash
gcloud  compute  instances create xds-client-vm  \
 --service-account=$SERVICE_ACCOUNT_EMAIL \
 --scopes=https://www.googleapis.com/auth/cloud-platform \
 --zone us-central1-a --image-family debian-11  --image-project=debian-cloud

gcloud compute ssh xds-client-vm  --zone  us-central1-a

apt-get update && apt-get install wget zip git -y

# Install golang
wget https://golang.org/dl/go1.20.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.20.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

On the VM the xds grpc Client:

```bash
git clone https://github.com/GoogleCloudPlatform/traffic-director-grpc-bootstrap.git
cd traffic-director-grpc-bootstrap/
go build .
export PROJECT_NUMBER=`curl -s "http://metadata.google.internal/computeMetadata/v1/project/numeric-project-id" -H "Metadata-Flavor: Google"`
./td-grpc-bootstrap --gcp-project-number=$PROJECT_NUMBER --gke-cluster-name=cluster-grpc --output=xds_bootstrap.json

export GRPC_XDS_BOOTSTRAP=`pwd`/xds_bootstrap.json

git clone https://github.com/salrashid123/gcegrpc.git
cd gcegrpc/gke_td_xds/client
go run src/grpc_client.go --host xds:///fe-srv-td:50051
```


What you should see is the list of backend

```bash
$ go run src/grpc_client.go --host xds:///fe-srv-td:50051 --useTLS=false
2024/03/30 13:23:39 RPC Response: 0 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-2hxsr"
2024/03/30 13:23:40 RPC Response: 1 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-2hxsr"
2024/03/30 13:23:41 RPC Response: 2 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-pw4wd"
2024/03/30 13:23:42 RPC Response: 3 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-2hxsr"
2024/03/30 13:23:43 RPC Response: 4 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-pw4wd"
2024/03/30 13:23:44 RPC Response: 5 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-2hxsr"
2024/03/30 13:23:45 RPC Response: 6 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-pw4wd"
2024/03/30 13:23:46 RPC Response: 7 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-2hxsr"
2024/03/30 13:23:47 RPC Response: 8 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-pw4wd"
2024/03/30 13:23:48 RPC Response: 9 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-2hxsr"
2024/03/30 13:23:49 RPC Response: 10 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-pw4wd"
2024/03/30 13:23:50 RPC Response: 11 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-2hxsr"
2024/03/30 13:23:51 RPC Response: 12 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-pw4wd"
2024/03/30 13:23:52 RPC Response: 13 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-2hxsr"
2024/03/30 13:23:53 RPC Response: 14 message:"Hello unary RPC msg   from hostname fe-deployment-68b49b7677-pw4wd"
```

### Delete

To delete this config:

```bash
gcloud compute forwarding-rules delete grpc-gke-forwarding-rule --global -q
gcloud compute target-grpc-proxies delete grpc-gke-proxy  -q 
gcloud compute url-maps delete grpc-gke-url-map -q
gcloud compute backend-services delete grpc-gke-td-service --global -q
gcloud compute firewall-rules delete grpc-gke-allow-health-checks -q
gcloud compute health-checks delete grpc-gke-td-hc -q
kubectl delete -f .
kubectl delete svcneg/fe-srv-td
gcloud compute network-endpoint-groups delete fe-srv-td --zone=us-central1-a -q
```


---


### With TLS

```bash
gcloud services enable trafficdirector.googleapis.com
gcloud container  clusters create cluster-grpc \
 --zone us-central1-a  --num-nodes 3 --enable-ip-alias \
 --tags=allow-health-checks

cd tls/
kubectl apply -f .  
kubectl get svcneg 
kubectl get svcneg fe-srv-td  -o yaml

# NEG_NAME is declared in fe-srv-td.yaml
gcloud compute network-endpoint-groups list 
export NEG_NAME="fe-srv-td"


## using https proxy
gcloud compute health-checks create https https-gke-td-hc  --port=8080 --request-path=/healthz --enable-logging

gcloud compute firewall-rules create https-gke-allow-health-checks \
  --network default --action allow --direction INGRESS \
  --source-ranges 35.191.0.0/16,130.211.0.0/22 \
  --target-tags allow-health-checks \
  --rules tcp:8080
    
gcloud compute backend-services create grpc-gke-td-service \
   --global \
   --load-balancing-scheme=INTERNAL_SELF_MANAGED \
   --protocol=GRPC \
   --health-checks https-gke-td-hc

gcloud compute backend-services add-backend grpc-gke-td-service \
   --global \
   --network-endpoint-group $NEG_NAME \
   --network-endpoint-group-zone us-central1-a \
   --balancing-mode RATE \
   --max-rate-per-endpoint 5

gcloud compute url-maps create grpc-gke-url-map --default-service grpc-gke-td-service

gcloud compute url-maps add-path-matcher grpc-gke-url-map \
--default-service grpc-gke-td-service \
--path-matcher-name grpc-gke-path-matcher \
--new-hosts fe-srv-td:50051

gcloud compute target-grpc-proxies create grpc-gke-proxy \
--url-map grpc-gke-url-map \
--validate-for-proxyless

gcloud compute forwarding-rules create grpc-gke-forwarding-rule \
--global \
--load-balancing-scheme=INTERNAL_SELF_MANAGED \
--address=0.0.0.0 \
--target-grpc-proxy=grpc-gke-proxy \
--ports 50051 \
--network default
```

Wait a couple of mins and you should see is traffic directory detecting the backend services you deployed

Now create a VM 

```bash
gcloud  compute  instances create xds-client-vm  \
 --service-account=$SERVICE_ACCOUNT_EMAIL \
 --scopes=https://www.googleapis.com/auth/cloud-platform \
 --zone us-central1-a --image-family debian-11  --image-project=debian-cloud

gcloud compute ssh xds-client-vm  --zone  us-central1-a

apt-get update && apt-get install wget zip git -y

# Install golang
wget https://golang.org/dl/go1.20.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.20.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

Acquire the xds grpc Client:

```bash
git clone https://github.com/GoogleCloudPlatform/traffic-director-grpc-bootstrap.git
cd traffic-director-grpc-bootstrap/

go build .
export PROJECT_NUMBER=`curl -s "http://metadata.google.internal/computeMetadata/v1/project/numeric-project-id" -H "Metadata-Flavor: Google"`

./td-grpc-bootstrap --gcp-project-number=$PROJECT_NUMBER --gke-cluster-name=cluster-grpc --output=xds_bootstrap.json

export GRPC_XDS_BOOTSTRAP=`pwd`/xds_bootstrap.json

git clone https://github.com/salrashid123/gcegrpc.git
cd gcegrpc/gke_td_xds/client

$ go run src/grpc_client.go --useTLS --host xds:///fe-srv-td:50051 --tlsCert ../../certs/CA_crt.pem

2024/03/30 13:42:13 RPC Response: 0 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-jrg66"
2024/03/30 13:42:14 RPC Response: 1 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-rvb2s"
2024/03/30 13:42:15 RPC Response: 2 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-jrg66"
2024/03/30 13:42:16 RPC Response: 3 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-rvb2s"
2024/03/30 13:42:17 RPC Response: 4 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-jrg66"
2024/03/30 13:42:18 RPC Response: 5 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-rvb2s"
2024/03/30 13:42:19 RPC Response: 6 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-jrg66"
2024/03/30 13:42:20 RPC Response: 7 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-rvb2s"
2024/03/30 13:42:21 RPC Response: 8 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-jrg66"
2024/03/30 13:42:22 RPC Response: 9 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-rvb2s"
2024/03/30 13:42:23 RPC Response: 10 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-jrg66"
2024/03/30 13:42:24 RPC Response: 11 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-rvb2s"
2024/03/30 13:42:25 RPC Response: 12 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-jrg66"
2024/03/30 13:42:26 RPC Response: 13 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-rvb2s"
2024/03/30 13:42:27 RPC Response: 14 message:"Hello unary RPC msg   from hostname fe-deployment-6f569bb8c8-jrg66"
```

### Delete

To delete this config:

```bash
gcloud compute forwarding-rules delete grpc-gke-forwarding-rule --global -q
gcloud compute target-grpc-proxies delete grpc-gke-proxy  -q 
gcloud compute url-maps delete grpc-gke-url-map -q
gcloud compute backend-services delete grpc-gke-td-service --global -q
gcloud compute firewall-rules delete https-gke-allow-health-checks -q
gcloud compute health-checks delete https-gke-td-hc -q
kubectl delete -f .
kubectl delete svcneg/fe-srv-td
gcloud compute network-endpoint-groups delete fe-srv-td --zone=us-central1-a -q
```

---

#### References

- [gRPC xDS Loadbalancing](https://github.com/salrashid123/grpc_xds)
- [gRPC HealthCheck Proxy](https://github.com/salrashid123/grpc_health_proxy)
- [Proxyless gRPC with Google Traffic Director](https://github.com/salrashid123/grpc_xds_traffic_director)
- [Kubernetes xDS service for gRPC loadbalancing](https://github.com/salrashid123/k8s_grpc_xds)

---

To enable verbose logging, first set

```bash
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info
```