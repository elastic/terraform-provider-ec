# -------------------------------------------------------------
# Create Compute VM + Elastic Agent
# -------------------------------------------------------------

resource "google_compute_instance" "vm_instance" {
  depends_on = [ec_deployment.elastic_gc_deployment, data.external.elastic_create_gcp_policy] ## We want to have the elastic deployment before we install the agent
  
  name = "elastic-agent"
  machine_type = "e2-standard-2"
  tags = ["terraformed"]
  
  
  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
      size = 20
    }
  }
  
  network_interface {
    network = "default"
	access_config {
      // Ephemeral public IP
    }
  }

  metadata_startup_script = <<SCRIPT
    #!/bin/bash
	
	#######
	## INIT
	######
	
	export CLOUD_AUTH=${ec_deployment.elastic_gc_deployment.elasticsearch_username}:${ec_deployment.elastic_gc_deployment.elasticsearch_password}
	export KIBANA_URL=${ec_deployment.elastic_gc_deployment.kibana[0].https_endpoint}
	integration_server_endpoint=${ec_deployment.elastic_gc_deployment.integrations_server[0].https_endpoint}
	export FLEET_URL=$${integration_server_endpoint//apm/fleet}
	export HOST_POLICY_ID=${data.external.elastic_create_gcp_policy.result.id}
	
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
	
  SCRIPT


}