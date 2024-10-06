#!/bin/bash

# Check if required environmental variables are set
if [ -z "$AZ_1" ] || [ -z "$AZ_2" ] || [ -z "$AZ_3" ]; then
    echo "Please set the following environment variables:"
    echo "AZ_1, AZ_2, AZ_3"
    exit 1
fi

# Default CIDRs
VPC_CIDR="10.0.0.0/16"
PUBLIC_SUBNET_CIDR_1="10.0.0.0/19"
PUBLIC_SUBNET_CIDR_2="10.0.32.0/19"
PUBLIC_SUBNET_CIDR_3="10.0.64.0/19"
PRIVATE_SUBNET_CIDR_1="10.0.96.0/19"
PRIVATE_SUBNET_CIDR_2="10.0.128.0/19"
PRIVATE_SUBNET_CIDR_3="10.0.160.0/19"

# Create the VPC
VPC_ID=$(aws ec2 create-vpc --cidr-block "$VPC_CIDR" --output text --query 'Vpc.VpcId')
echo "Created VPC: $VPC_ID"

# Create Public Subnets
PUBLIC_SUBNET_ID_1=$(aws ec2 create-subnet --vpc-id "$VPC_ID" --cidr-block "$PUBLIC_SUBNET_CIDR_1" --availability-zone "$AZ_1" --output text --query 'Subnet.SubnetId')
PUBLIC_SUBNET_ID_2=$(aws ec2 create-subnet --vpc-id "$VPC_ID" --cidr-block "$PUBLIC_SUBNET_CIDR_2" --availability-zone "$AZ_2" --output text --query 'Subnet.SubnetId')
PUBLIC_SUBNET_ID_3=$(aws ec2 create-subnet --vpc-id "$VPC_ID" --cidr-block "$PUBLIC_SUBNET_CIDR_3" --availability-zone "$AZ_3" --output text --query 'Subnet.SubnetId')
echo "Created Public Subnets: $PUBLIC_SUBNET_ID_1, $PUBLIC_SUBNET_ID_2, $PUBLIC_SUBNET_ID_3"

# Create Private Subnets
PRIVATE_SUBNET_ID_1=$(aws ec2 create-subnet --vpc-id "$VPC_ID" --cidr-block "$PRIVATE_SUBNET_CIDR_1" --availability-zone "$AZ_1" --output text --query 'Subnet.SubnetId')
PRIVATE_SUBNET_ID_2=$(aws ec2 create-subnet --vpc-id "$VPC_ID" --cidr-block "$PRIVATE_SUBNET_CIDR_2" --availability-zone "$AZ_2" --output text --query 'Subnet.SubnetId')
PRIVATE_SUBNET_ID_3=$(aws ec2 create-subnet --vpc-id "$VPC_ID" --cidr-block "$PRIVATE_SUBNET_CIDR_3" --availability-zone "$AZ_3" --output text --query 'Subnet.SubnetId')
echo "Created Private Subnets: $PRIVATE_SUBNET_ID_1, $PRIVATE_SUBNET_ID_2, $PRIVATE_SUBNET_ID_3"

# Create an Internet Gateway
INTERNET_GATEWAY_ID=$(aws ec2 create-internet-gateway --output text --query 'InternetGateway.InternetGatewayId')
aws ec2 attach-internet-gateway --vpc-id "$VPC_ID" --internet-gateway-id "$INTERNET_GATEWAY_ID"
echo "Created and attached Internet Gateway: $INTERNET_GATEWAY_ID"

# Create a Route Table for public subnets
PUBLIC_ROUTE_TABLE_ID=$(aws ec2 create-route-table --vpc-id "$VPC_ID" --output text --query 'RouteTable.RouteTableId')
aws ec2 create-route --route-table-id "$PUBLIC_ROUTE_TABLE_ID" --destination-cidr-block 0.0.0.0/0 --gateway-id "$INTERNET_GATEWAY_ID"
echo "Created Route Table for public subnets: $PUBLIC_ROUTE_TABLE_ID"

# Associate Public Subnets with the Route Table
aws ec2 associate-route-table --subnet-id "$PUBLIC_SUBNET_ID_1" --route-table-id "$PUBLIC_ROUTE_TABLE_ID"
aws ec2 associate-route-table --subnet-id "$PUBLIC_SUBNET_ID_2" --route-table-id "$PUBLIC_ROUTE_TABLE_ID"
aws ec2 associate-route-table --subnet-id "$PUBLIC_SUBNET_ID_3" --route-table-id "$PUBLIC_ROUTE_TABLE_ID"

echo "Associated Public Subnets with Route Table"

# Create NAT Gateway for private subnets
NAT_GATEWAY_EIP=$(aws ec2 allocate-address --query 'AllocationId' --output text)
NAT_GATEWAY_ID=$(aws ec2 create-nat-gateway --subnet-id "$PUBLIC_SUBNET_ID_1" --allocation-id "$NAT_GATEWAY_EIP" --query 'NatGateway.NatGatewayId' --output text)
aws ec2 wait nat-gateway-available --nat-gateway-ids "$NAT_GATEWAY_ID"
echo "Created NAT Gateway: $NAT_GATEWAY_ID"

# Create a Route Table for private subnets
PRIVATE_ROUTE_TABLE_ID=$(aws ec2 create-route-table --vpc-id "$VPC_ID" --output text --query 'RouteTable.RouteTableId')
aws ec2 create-route --route-table-id "$PRIVATE_ROUTE_TABLE_ID" --destination-cidr-block 0.0.0.0/0 --nat-gateway-id "$NAT_GATEWAY_ID"
echo "Created Route Table for private subnets: $PRIVATE_ROUTE_TABLE_ID"

# Associate Private Subnets with the Route Table
aws ec2 associate-route-table --subnet-id "$PRIVATE_SUBNET_ID_1" --route-table-id "$PRIVATE_ROUTE_TABLE_ID"
aws ec2 associate-route-table --subnet-id "$PRIVATE_SUBNET_ID_2" --route-table-id "$PRIVATE_ROUTE_TABLE_ID"
aws ec2 associate-route-table --subnet-id "$PRIVATE_SUBNET_ID_3" --route-table-id "$PRIVATE_ROUTE_TABLE_ID"

echo "Associated Private Subnets with Route Table"

# Add load balance tags to the private subnets
aws ec2 create-tags --resources "$PRIVATE_SUBNET_ID_1" --tags Key=kubernetes.io/role/internal-elb,Value=1
aws ec2 create-tags --resources "$PRIVATE_SUBNET_ID_2" --tags Key=kubernetes.io/role/internal-elb,Value=1
aws ec2 create-tags --resources "$PRIVATE_SUBNET_ID_3" --tags Key=kubernetes.io/role/internal-elb,Value=1

echo "Added Internal-ELB Tags to Private Subnets"

# Add load balance tags to the public subnets
aws ec2 create-tags --resources "$PUBLIC_SUBNET_ID_1" --tags Key=kubernetes.io/role/elb,Value=1
aws ec2 create-tags --resources "$PUBLIC_SUBNET_ID_2" --tags Key=kubernetes.io/role/elb,Value=1
aws ec2 create-tags --resources "$PUBLIC_SUBNET_ID_3" --tags Key=kubernetes.io/role/elb,Value=1

echo "Added External-ELB Tags to Public Subnets"

# Add map public ip to the public subnets
aws ec2 modify-subnet-attribute --subnet-id "$PUBLIC_SUBNET_ID_1" --map-public-ip-on-launch
aws ec2 modify-subnet-attribute --subnet-id "$PUBLIC_SUBNET_ID_2" --map-public-ip-on-launch
aws ec2 modify-subnet-attribute --subnet-id "$PUBLIC_SUBNET_ID_3" --map-public-ip-on-launch

echo "Added Map Public IP to Public Subnets"

# Tagging Resources
aws ec2 create-tags --resources "$VPC_ID" --tags Key=Name,Value=example-cluster
aws ec2 create-tags --resources "$INTERNET_GATEWAY_ID" --tags Key=Name,Value=example-cluster-ig
aws ec2 create-tags --resources "$NAT_GATEWAY_ID" --tags Key=Name,Value=example-cluster-nat

# Generate the configuration files
cat << EOF > cluster-launch.yaml
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: example-cluster
  region: us-east-1
  version: '1.31'
vpc:
  id: "$VPC_ID"
  cidr: "$VPC_CIDR"
  subnets:
    public:
      $AZ_1:
        id: "$PUBLIC_SUBNET_ID_1"
        cidr: "$PUBLIC_SUBNET_CIDR_1"
      $AZ_2:
        id: "$PUBLIC_SUBNET_ID_2"
        cidr: "$PUBLIC_SUBNET_CIDR_2"
      $AZ_3:
        id: "$PUBLIC_SUBNET_ID_3"
        cidr: "$PUBLIC_SUBNET_CIDR_3"
    private:
      $AZ_1:
        id: "$PRIVATE_SUBNET_ID_1"
        cidr: "$PRIVATE_SUBNET_CIDR_1"
      $AZ_2:
        id: "$PRIVATE_SUBNET_ID_2"
        cidr: "$PRIVATE_SUBNET_CIDR_2"
      $AZ_3:
        id: "$PRIVATE_SUBNET_ID_3"
        cidr: "$PRIVATE_SUBNET_CIDR_3"
managedNodeGroups:
  - name: system-node
    instanceType: m5.large
    desiredCapacity: 1
    amiFamily: AmazonLinux2
    volumeSize: 30
    volumeIOPS: 3000
    volumeThroughput: 125
    volumeType: gp3
    privateNetworking: true
    availabilityZones: ["$AZ_1", "$AZ_2", "$AZ_3"]
    labels:
      cluster.io/instance-type: system
EOF

echo "cluster-launch.yaml generated successfully."

cat <<EOL > database-node-launch.yaml
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: example-cluster
  region: us-east-1
  version: '1.31'
managedNodeGroups:
  - name: redis-database-node
    instanceType: m5.large
    desiredCapacity: 3
    amiFamily: AmazonLinux2
    volumeSize: 30
    volumeIOPS: 3000
    volumeThroughput: 125
    volumeType: gp3
    privateNetworking: true
    availabilityZones: ["$AZ_1", "$AZ_2", "$AZ_3"]
    labels:
      cluster.io/instance-type: memory-database
    taints:
      - key: redis-database-key
        value: "true"
        effect: NoSchedule
EOL


cat <<EOL > k6-node-launch.yaml
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: example-cluster
  region: us-east-1
  version: '1.31'
managedNodeGroups:
  - name: redis-k6-node
    instanceType: m5.2xlarge
    desiredCapacity: 1
    amiFamily: AmazonLinux2
    volumeSize: 30
    volumeIOPS: 3000
    volumeThroughput: 125
    volumeType: gp3
    privateNetworking: false
    availabilityZones: ["$AZ_1", "$AZ_2", "$AZ_3"]
    labels:
      cluster.io/instance-type: redis-k6-node
    taints:
      - key: redis-k6-key
        value: "true"
        effect: NoSchedule
EOL


echo "database-node-launch.yaml generated successfully."

cat <<EOL > gen-purpose-arm64-node-launch.yaml
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: example-cluster
  region: us-east-1
  version: '1.31' 
managedNodeGroups:
  - name: gen-purpose-arm64
    instanceType: c7g.medium
    desiredCapacity: 6
    amiFamily: AmazonLinux2
    volumeSize: 10
    volumeIOPS: 3000
    volumeThroughput: 125
    volumeType: gp3
    privateNetworking: true
    availabilityZones: ["$AZ_1", "$AZ_2", "$AZ_3"]
    labels:
      cluster.io/instance-type: gp-arm64
EOL

echo "gen-purpose-arm64-node-launch.yaml generated successfully."

cat <<EOL > gen-purpose-amd64-node-launch.yaml
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: example-cluster
  region: us-east-1
  version: '1.31' 
managedNodeGroups:
  - name: gen-purpose-amd64
    instanceType: t3.small
    desiredCapacity: 3
    amiFamily: AmazonLinux2
    volumeSize: 10
    volumeIOPS: 3000
    volumeThroughput: 125
    volumeType: gp3
    privateNetworking: true
    availabilityZones: ["$AZ_1", "$AZ_2", "$AZ_3"]
    labels:
      cluster.io/instance-type: gp-burst-amd64
EOL



echo "gen-purpose-amd64-node-launch.yaml generated successfully."


