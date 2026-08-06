# -------------------------------------------------------------
# Create Compute VM + Elastic Agent
# -------------------------------------------------------------

data "template_file" "install_agent" {
  template = file("../../lib/scripts/agent_install.sh")
  vars = {
    elastic_version = var.elastic_version
    elasticsearch_username = ec_deployment.elastic_deployment.elasticsearch_username
    elasticsearch_password = ec_deployment.elastic_deployment.elasticsearch_password
    kibana_endpoint = ec_deployment.elastic_deployment.kibana[0].https_endpoint
    integration_server_endpoint = ec_deployment.elastic_deployment.integrations_server[0].https_endpoint
    policy_id = data.external.elastic_create_policy.result.id
  }
}

resource "google_compute_instance" "vm_instance" {
  depends_on = [ec_deployment.elastic_deployment, data.external.elastic_create_policy] ## We want to have the elastic deployment before we install the agent
  
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
    network = var.google_cloud_network
	access_config {
      // Ephemeral public IP
    }
  }

  metadata_startup_script = "${data.template_file.install_agent.rendered}"
}