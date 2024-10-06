# Generate the cluster configuration files
- chmod +x ./vpc-script.sh
- ./vpc-script.sh

# Launch the cluster using eksctl 
- eksctl create cluster -f ./cluster-launch.yaml


# Install Kubernetes Dashboard(Optional)
- helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/
- helm upgrade --install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --create-namespace --namespace kubernetes-dashboard
- kubectl apply -f ./dashboard-adminuser.yml 
- kubectl -n kubernetes-dashboard create token admin-user --duration=48h 
- kubectl -n kubernetes-dashboard port-forward svc/kubernetes-dashboard-kong-proxy 8443:443
- https://127.0.0.1:8443


# Install Certificate Manager
- kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.15.3/cert-manager.yaml




# Install Prometheus and Grafana
- make sure you have jsonnet installed
- cd kube-prometheus
- jb init 
- jb install github.com/prometheus-operator/kube-prometheus/jsonnet/kube-prometheus@main
- wget https://raw.githubusercontent.com/prometheus-operator/kube-prometheus/main/build.sh -O build.sh
- chmod +x build.sh
- kubectl create namespace redis-database
- jb update
- ./build.sh redis-database.jsonnet
- kubectl apply --server-side -f manifests/setup
- kubectl wait \
	--for condition=Established \
	--all CustomResourceDefinition \
	--namespace=monitoring
- kubectl apply -f manifests/

- kubectl port-forward svc/grafana -n monitoring 3000:3000
- kubectl port-forward svc/prometheus-k8s -n monitoring 9090:9090

- login to localhost:3000 & localhost:9090



# Create an IAM OIDC provider for your cluster
- cluster_name=example-cluster
- oidc_id=$(aws eks describe-cluster --name $cluster_name --query "cluster.identity.oidc.issuer" --output text | cut -d '/' -f 5)
- echo $oidc_id
- aws iam list-open-id-connect-providers | grep $oidc_id | cut -d "/" -f4
- if an output was returned then skip the next step
- eksctl utils associate-iam-oidc-provider --cluster $cluster_name --approve



# Create AWS Load Balancer
- skip this step if you already have a load balancer policy
- curl -O https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.7.2/docs/install/iam_policy.json
- aws iam create-policy \
    --policy-name AWSLoadBalancerControllerIAMPolicy \
    --policy-document file://iam_policy.json

# Replace the arn with the one outputted in the previous step
- eksctl create iamserviceaccount \
  --cluster=example-cluster \
  --namespace=kube-system \
  --name=aws-load-balancer-controller \
  --role-name AmazonEKSLoadBalancerControllerRole \
  --attach-policy-arn=arn:aws:iam::123456789123:policy/AWSLoadBalancerControllerIAMPolicy \
  --approve




- helm repo add eks https://aws.github.io/eks-charts

- helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=example-cluster \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller 

# Make sure the load balancer is installed
- kubectl get deployments -n kube-system

# Should see
kubectl get deployments -n kube-system
NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
aws-load-balancer-controller   2/2     2            2           22s

# Create the worker database nodes
- eksctl create nodegroup --config-file=./database-node-launch.yaml

# Launch the redis database operator
- kubectl apply -f ./redis_operator_resources.yaml

# Install the service monitor for Redis
- kubectl apply -f ./service-monitor.yaml

# Launch the redis database instances
- kubectl apply -f ./redisreplication-launch.yaml

# Check the launch of the Redis Instances through the dashboard or kubectl. Should see
kubectl get pods -n redis-database
NAME                 READY   STATUS    RESTARTS   AGE
redisreplication-0   2/2     Running   0          55s
redisreplication-1   1/2     Running   0          55s
redisreplication-2   1/2     Running   0          55s

# Check the instances through Prometheus with configuration file 763_rev6.json

# Launch the sentinel instance
- eksctl create nodegroup --config-file=./gen-purpose-amd64-node-launch.yaml

# Launch the sentinels 
- kubectl apply -f './redissentinel-launch.yaml'

# Check the launch of the Sentinel Instances through the dashboard or kubectl. Should see
kubectl get pods -n redis-database
NAME                 READY   STATUS    RESTARTS   AGE
redisreplication-0   2/2     Running   0          16m
redisreplication-1   1/2     Running   0          16m
redisreplication-2   1/2     Running   0          16m
redissentinel-0      1/1     Running   0          2m8s
redissentinel-1      1/1     Running   0          107s
redissentinel-2      1/1     Running   0          87s

# Launch the worker instances
- eksctl create nodegroup --config-file=./gen-purpose-arm64-node-launch.yaml

# Launch the workers
- kubectl apply -f './redis-worker-deployment.yaml'


# Get the address
- kubectl describe ingress/redis-ingress -n redis-database

# If you have a sufficient upload 50Mbs+ you can run the test locally
k6 run api-test.js

# Alternatively you can run the tests within the cloud
k6 cloud api-test.js

# Or you can spend a bit less by launching k6 into Kubernetes. Note the instance is much larger. Run the test and remove it.
- eksctl create nodegroup --config-file=./k6-node-launch.yaml


# Run K6 with address 
- k6 run api-test.js


# Removing the resources
- eksctl delete nodegroup --config-file='./gen-purpose-arm64-node-launch.yaml' --approve
- eksctl delete nodegroup --config-file='./gen-purpose-amd64-node-launch.yaml' --approve
- eksctl delete nodegroup --config-file='./database-node-launch.yaml' --approve
- eksctl delete nodegroup --config-file='./k6-node-launch.yaml' --approve
-  eksctl delete cluster -f '/home/jhane/workspace/custom-resource-operators/cluster-launch.yaml'


# Wait for the resources to be completely deleted. 
- export VPC_ID=YOUR_VPC_ID
- chmod +x vpc-script.sh
- ./vpc-script.sh



