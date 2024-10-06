#!/bin/bash

# Check if required environmental variables are set
if [ -z "$VPC_ID" ]; then
    echo "Please set the following environment variable:"
    echo "VPC_ID"
    exit 1
fi

# Function to delete all subnets
delete_subnets() {
    local subnet_ids
    subnet_ids=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" --query "Subnets[].SubnetId" --output text)
    
    for subnet_id in $subnet_ids; do
        echo "Deleting subnet: $subnet_id"
        aws ec2 delete-subnet --subnet-id "$subnet_id"
    done
}

# Function to detach and delete the Internet Gateway
delete_internet_gateway() {
    local igw_id
    igw_id=$(aws ec2 describe-internet-gateways --filters "Name=attachment.vpc-id,Values=$VPC_ID" --query "InternetGateways[].InternetGatewayId" --output text)
    
    if [ -n "$igw_id" ]; then
        echo "Detaching Internet Gateway: $igw_id from VPC: $VPC_ID"
        aws ec2 detach-internet-gateway --internet-gateway-id "$igw_id" --vpc-id "$VPC_ID"
        echo "Deleting Internet Gateway: $igw_id"
        aws ec2 delete-internet-gateway --internet-gateway-id "$igw_id"
    else
        echo "No Internet Gateway found for VPC: $VPC_ID"
    fi
}

# Function to delete the NAT Gateway
delete_nat_gateway() {
    local nat_gateway_id
    nat_gateway_id=$(aws ec2 describe-nat-gateways --filter "Name=vpc-id,Values=$VPC_ID" --query "NatGateways[].NatGatewayId" --output text)

    if [ -n "$nat_gateway_id" ]; then
        echo "Deleting NAT Gateway: $nat_gateway_id"
        aws ec2 delete-nat-gateway --nat-gateway-id "$nat_gateway_id"
    else
        echo "No NAT Gateway found for VPC: $VPC_ID"
    fi
}

# Function to delete the Route Tables
delete_route_tables() {
    local route_table_ids
    route_table_ids=$(aws ec2 describe-route-tables --filters "Name=vpc-id,Values=$VPC_ID" --query "RouteTables[].RouteTableId" --output text)

    for route_table_id in $route_table_ids; do
        if [ "$route_table_id" != "$(aws ec2 describe-route-tables --filters "Name=vpc-id,Values=$VPC_ID" --query "RouteTables[?Associations[?Main==\`true\`]].RouteTableId" --output text)" ]; then
            echo "Deleting Route Table: $route_table_id"
            aws ec2 delete-route-table --route-table-id "$route_table_id"
        fi
    done
}

# Function to delete the VPC
delete_vpc() {
    echo "Deleting VPC: $VPC_ID"
    aws ec2 delete-vpc --vpc-id "$VPC_ID"
}

# Execute the cleanup functions
delete_subnets
delete_internet_gateway
delete_nat_gateway
delete_route_tables
delete_vpc

echo "VPC cleanup completed."

