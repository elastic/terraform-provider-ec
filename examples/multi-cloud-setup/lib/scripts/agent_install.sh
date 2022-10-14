#!/bin/bash
	
#######
## INIT
######

export ES_VERSION=${elastic_version}
export CLOUD_AUTH=${elasticsearch_username}:${elasticsearch_password}
export KIBANA_URL=${kibana_endpoint}
integration_server_endpoint=${integration_server_endpoint}
export FLEET_URL=$${integration_server_endpoint//apm/fleet}
export HOST_POLICY_ID=${policy_id}


echo "deb http://us.archive.ubuntu.com/ubuntu vivid main universe" | sudo tee -a /etc/apt/sources.list
sudo apt-get update
sudo apt-get --assume-yes install jq

#########
## install agent
#########

## get version
version=$(curl -XGET -u $CLOUD_AUTH "$${KIBANA_URL}/api/status" -H "kbn-xsrf: true" -H "Content-Type: application/json" | jq -r '.version.number')
echo ES_VERSION=$version >> /etc/.env
export ES_VERSION=$version

echo "Command: curl -XGET -u $CLOUD_AUTH \"$${KIBANA_URL}/api/fleet/enrollment_api_keys\""

response=$(curl -XGET -u $CLOUD_AUTH "$${KIBANA_URL}/api/fleet/enrollment_api_keys" -H "kbn-xsrf: true" -H "Content-Type: application/json") 
echo $response 

echo "using $${HOST_POLICY_ID}"

endpoint_enroll_key=$( jq -r --arg policy_id "$${HOST_POLICY_ID}" '.list[] | select(.policy_id == $policy_id) | .api_key' <<< "$${response}" )
export HOST_ENROLL_KEY=$endpoint_enroll_key
echo HOST_ENROLL_KEY=$endpoint_enroll_key >> /etc/.env

echo "Loading agent"
curl -L -O https://artifacts.elastic.co/downloads/beats/elastic-agent/elastic-agent-$ES_VERSION-linux-x86_64.tar.gz
sleep 30

echo "Unpack agent"
tar xzvf elastic-agent-$ES_VERSION-linux-x86_64.tar.gz
cd elastic-agent-$ES_VERSION-linux-x86_64

echo "Install agent"
echo "Command: ./elastic-agent install --url=$FLEET_URL --enrollment-token=$HOST_ENROLL_KEY -f"
./elastic-agent install --url=$FLEET_URL --enrollment-token=$HOST_ENROLL_KEY -f
cd ..
rm elastic-agent-$ES_VERSION-linux-x86_64 -r
rm elastic-agent-$ES_VERSION-linux-x86_64.tar.gz