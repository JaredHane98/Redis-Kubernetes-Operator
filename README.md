This repository showcases a highly available and durable Redis database with Replicants and Sentinels spread across multiple availability zones. It features leader promotion, follower recovery, sentinel recovery, TLS authentication with certificate manager, and password protection. The accompanying example is designed to run on AWS, but the standalone operator should run anywhere including minikube or kind clusters. 

# Prerequisites

[kubectl](https://kubernetes.io/docs/reference/kubectl/)
[docker](https://docs.docker.com/engine/install/)
[eksctl](https://eksctl.io/installation/)
[jsonnet](https://github.com/google/jsonnet)


# Getting started

Clone the repository

```bash
git clone https://github.com/JaredHane98/Redis-Kubernetes-Operator.git
cd Redis-Kubernetes-Operator
```

# Creating cluster resources

First we need to create a highly available VPC with subnets spread across multiple AZs.
```bash
chmod +x ./vpc-script.sh
./vpc-script.sh
```

# Setting up system node

Once the VPC script has complete we can begin creating the cluster nodes. 

```bash
eksctl create cluster -f ./cluster-launch.yaml
```

# Install kubernetes dashboard(Optional)

```bash
helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/
helm upgrade --install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --create-namespace --namespace kubernetes-dashboard
kubectl apply -f ./dashboard-adminuser.yml 
kubectl -n kubernetes-dashboard create token admin-user --duration=48h 
kubectl -n kubernetes-dashboard port-forward svc/kubernetes-dashboard-kong-proxy 8443:443
```

Access the dashboard at https://127.0.0.1:8443


# Install certificate manager
- kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.15.3/cert-manager.yaml


# Install Prometheus and Grafana
```bash
cd kube-prometheus
jb init 
jb install github.com/prometheus-operator/kube-prometheus/jsonnet/kube-prometheus@main
wget https://raw.githubusercontent.com/prometheus-operator/kube-prometheus/main/build.sh -O build.sh
chmod +x build.sh
kubectl create namespace redis-database
jb update
./build.sh redis-database.jsonnet
kubectl apply --server-side -f manifests/setup
kubectl wait \
	--for condition=Established \
	--all CustomResourceDefinition \
	--namespace=monitoring
kubectl apply -f manifests/

kubectl port-forward svc/grafana -n monitoring 3000:3000
kubectl port-forward svc/prometheus-k8s -n monitoring 9090:9090
```

Access the Prometheus and Grafana Dashboard at localhost:3000 and localhost:9090


# Create an AWS IAM OIDC provider for the cluster

```bash
cluster_name=example-cluster
oidc_id=$(aws eks describe-cluster --name $cluster_name --query "cluster.identity.oidc.issuer" --output text | cut -d '/' -f 5)
echo $oidc_id
aws iam list-open-id-connect-providers | grep $oidc_id | cut -d "/" -f4
```

If an output was returned skip this step

```bash
eksctl utils associate-iam-oidc-provider --cluster $cluster_name --approve
```


# Create AWS Load Balancer

Skip this step if you already have a AWS Load Balancer policy.

```bash
aws iam create-policy \
  --policy-name AWSLoadBalancerControllerIAMPolicy \
  --policy-document file://iam_policy.json
```

Replace role attach-policy-arn with the output of the previous step

```bash
eksctl create iamserviceaccount \
  --cluster=example-cluster \
  --namespace=kube-system \
  --name=aws-load-balancer-controller \
  --role-name AmazonEKSLoadBalancerControllerRole \
  --attach-policy-arn=arn:aws:iam::123456789123:policy/AWSLoadBalancerControllerIAMPolicy \
  --approve
helm repo add eks https://aws.github.io/eks-charts
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=example-cluster \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller
```

Make sure the load balancer is installed correctly

```bash
kubectl get deployments -n kube-system
```

You should see

```bash
kubectl get deployments -n kube-system
NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
aws-load-balancer-controller   2/2     2            2           22s
```

# Create the Redis Database Node

```bash
eksctl create nodegroup --config-file=./database-node-launch.yaml
```

# Launch the Redis Database Operator

```bash
kubectl apply -f ./redis_operator_resources.yaml
```

# Install the service monitor for Redis

```bash
kubectl apply -f ./service-monitor.yaml
```

# Launch The Redis Database

```bash
kubectl apply -f ./redisreplication-launch.yaml
```

Check the status of the Redis database

```bash
kubectl get pods -n redis-database
NAME                 READY   STATUS    RESTARTS   AGE
redisreplication-0   2/2     Running   0          55s
redisreplication-1   1/2     Running   0          55s
redisreplication-2   1/2     Running   0          55s
```

Note the the master is the only instance considered READY. This is purposefully done to prevent traffic from being directed towards the slaves.

You can now upload redis-dashboard.json to Grafana to view the statistics of your Redis Database.

# Launch the sentinel nodes

```bash
- eksctl create nodegroup --config-file=./gen-purpose-amd64-node-launch.yaml
```

# Launch the sentinels 

```bash
kubectl apply -f './redissentinel-launch.yaml'
```

Check the status of the Sentinel Instances

```bash
kubectl get pods -n redis-database
NAME                 READY   STATUS    RESTARTS   AGE
redisreplication-0   2/2     Running   0          16m
redisreplication-1   1/2     Running   0          16m
redisreplication-2   1/2     Running   0          16m
redissentinel-0      1/1     Running   0          2m8s
redissentinel-1      1/1     Running   0          107s
redissentinel-2      1/1     Running   0          87s
```

# Launch the worker nodes

```bash
eksctl create nodegroup --config-file=./gen-purpose-arm64-node-launch.yaml
```

# Launch the workers

```bash
kubectl apply -f './redis-worker-deployment.yaml'
```

Get the address.

```bash
kubectl describe ingress/redis-ingress -n redis-database
```

Make an enviromental variable of the URL/

```bash
export TARGET_URL=URL
```

Wait for the address to become available.

```bash
curl --request GET http://{$TARGET_URL}/readiness
OK
```

# Running K6 test

Run the test locally(REQUIRES 50MBPS+ UPLOAD AND POWERFUL MACHINE)

```bash
cd k6
k6 run api-test.js
```

Run the test in the cloud.

```bash
cd k6
k6 cloud api-test.js
```

Run K6 test in Kubernetes. 

```bash
eksctl create nodegroup --config-file=./k6-node-launch.yaml
```

Create the deployment

```bash
cat > k6-job.yaml << EOF
---
apiVersion: batch/v1
kind: Job
metadata:
  name: k6-job
  labels:
    app: k6
spec:
  backoffLimit: 3
  template:
    metadata:
      labels:
        app: k6
    spec:
      containers:
        - name: k6
          image: grafana/k6:latest
          args:
            - run
            - /scripts/script.js
          volumeMounts:
            - name: k6-test-script
              mountPath: /scripts
          env:
          - name: TARGET_URL
            value: $TARGET_URL
      nodeSelector:
        cluster.io/instance-type: redis-k6-node
      tolerations:
        - key: "redis-k6-key"
          operator: "Equal"
          value: "true"
          effect: "NoSchedule"
      volumes:
        - name: k6-test-script
          configMap:
            name: k6-test-script
      restartPolicy: Never
EOF
```

Launch the K6 test

```bash
kubectl apply -f ./k6-configmap.yaml
kubectl apply -f ./k6-job.yaml
```


# Cleaning up the resources

```bash
eksctl delete nodegroup --config-file='./gen-purpose-arm64-node-launch.yaml' --approve
eksctl delete nodegroup --config-file='./gen-purpose-amd64-node-launch.yaml' --approve
eksctl delete nodegroup --config-file='./database-node-launch.yaml' --approve
eksctl delete nodegroup --config-file='./k6-node-launch.yaml' --approve
eksctl delete cluster -f '/home/jhane/workspace/custom-resource-operators/cluster-launch.yaml'
```

Wait for the cluster resources to be completed deleted.

```bash
export VPC_ID=YOUR_VPC_ID
chmod +x vpc-script.sh
./vpc-script.sh
```




