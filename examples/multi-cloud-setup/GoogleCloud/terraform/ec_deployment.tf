# -------------------------------------------------------------
#  Deploy Elastic Cloud
# -------------------------------------------------------------
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = var.elastic_region
}

resource "ec_deployment" "elastic_deployment" {
  name                    = var.elastic_deployment_name
  region                  = var.elastic_region
  version                 = var.elastic_version == "latest" ? data.ec_stack.latest.version : var.elastic_version
  deployment_template_id  = var.elastic_deployment_template_id
  elasticsearch {
	  autoscale = "true"

    dynamic "remote_cluster" {
      for_each = var.elastic_remotes
      content {
        deployment_id = remote_cluster.value["id"]
        alias         = remote_cluster.value["alias"]
      }
    }
  }
  kibana {}
  integrations_server {}
}

output "elastic_cluster_id_google" {
  value = ec_deployment.elastic_deployment.id
}

output "elastic_cluster_alia_google" {
  value = ec_deployment.elastic_deployment.name
}

output "elastic_endpoint_google" {
  value = ec_deployment.elastic_deployment.kibana[0].https_endpoint
}

output "elastic_cloud_id_google" {
  value = ec_deployment.elastic_deployment.elasticsearch[0].cloud_id
}

output "elastic_username_google" {
  value = ec_deployment.elastic_deployment.elasticsearch_username
}

# -------------------------------------------------------------
#  Load Policy
# -------------------------------------------------------------

data "external" "elastic_create_policy" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    elastic_json_body = templatefile("${path.module}/../json_templates/default-policy.json", {"policy_name": "GC_${var.google_cloud_project}"})
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/kb_create_agent_policy.sh" ]
  depends_on = [ec_deployment.elastic_deployment]
}

output "elastic_create_policy" {
  value = data.external.elastic_create_policy.result
  depends_on = [data.external.elastic_create_policy]
}

data "external" "elastic_add_integration" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    elastic_json_body = templatefile("${path.module}/../json_templates/gcp_integration.json", 
    {
    "policy_id": data.external.elastic_create_policy.result.id,
    "gcp_project": var.google_cloud_project,
    "gcp_credentials_json": jsonencode(file(var.google_cloud_service_account_path)),
    "audit_log_topic": var.google_pubsub_audit_topic,
    "firewall_log_topic": var.google_pubsub_firewall_topic,
    "vpcflow_log_topic": var.google_pubsub_vpcflow_topic,
    "dns_log_topic": var.google_pubsub_dns_topic,
    "lb_log_topic": var.google_pubsub_lb_topic     
    }
    )
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/kb_add_integration_to_policy.sh" ]
  depends_on = [data.external.elastic_create_policy]
}

output "elastic_add_integration" {
  value = data.external.elastic_add_integration.result
  depends_on = [data.external.elastic_add_integration]
}

# -------------------------------------------------------------
#  Load Rules
# -------------------------------------------------------------

data "external" "elastic_load_rules" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/kb_load_detection_rules.sh" ]
  depends_on = [ec_deployment.elastic_deployment]
}

data "external" "elastic_enable_rules" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    elastic_json_body = templatefile("${path.module}/../json_templates/es_rule_activation.json",{})
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/kb_enable_detection_rules.sh" ]
  depends_on = [data.external.elastic_load_rules]
}

output "elastic_enable_rules" {
  value = data.external.elastic_enable_rules.result
  depends_on = [data.external.elastic_enable_rules]
}

# -------------------------------------------------------------
#  Create and Start transforms
# -------------------------------------------------------------

data "external" "elastic_create_transforms" {
  query = {
    elastic_endpoint  = ec_deployment.elastic_deployment.elasticsearch[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    transform_name    = "gcs-repo-transform"
    elastic_json_body = templatefile("${path.module}/../json_templates/es_repo_transform.json",{})
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/es_create_transform.sh" ]
  depends_on = [ec_deployment.elastic_deployment]
}

data "external" "elastic_start_transforms" {
  query = {
    elastic_endpoint  = ec_deployment.elastic_deployment.elasticsearch[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    transform_name    = "gcs-repo-transform"
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/es_start_transform.sh" ]
  depends_on = [data.external.elastic_create_transforms]
}

output "elastic_start_transforms" {
  value = data.external.elastic_start_transforms.result
  depends_on = [data.external.elastic_start_transforms]
}

################################################################################

data "external" "elastic_create_transform_host_metrics" {
  query = {
    elastic_endpoint  = ec_deployment.elastic_deployment.elasticsearch[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    transform_name    = "host-profile-transform"
    elastic_json_body = templatefile("${path.module}/../json_templates/es_host_transform.json",{})
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/es_create_transform.sh" ]
  depends_on = [ec_deployment.elastic_deployment]
}

data "external" "elastic_start_transform_host_metrics" {
  query = {
    elastic_endpoint  = ec_deployment.elastic_deployment.elasticsearch[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    transform_name    = "host-profile-transform"
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/es_start_transform.sh" ]
  depends_on = [data.external.elastic_create_transform_host_metrics]
}

output "elastic_start_transform_host_metrics" {
  value = data.external.elastic_start_transform_host_metrics.result
  depends_on = [data.external.elastic_start_transform_host_metrics]
}

################################################################################

data "external" "elastic_create_transform_vpc_flow" {
  query = {
    elastic_endpoint  = ec_deployment.elastic_deployment.elasticsearch[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    transform_name    = "vpc_flow-transform"
    elastic_json_body = templatefile("${path.module}/../json_templates/es_vpc_flow_transform.json",{})
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/es_create_transform.sh" ]
  depends_on = [ec_deployment.elastic_deployment]
}

data "external" "elastic_start_transform_vpc_flow" {
  query = {
    elastic_endpoint  = ec_deployment.elastic_deployment.elasticsearch[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    transform_name    = "vpc_flow-transform"
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/es_start_transform.sh" ]
  depends_on = [data.external.elastic_create_transform_vpc_flow]
}

output "elastic_start_transform_vpc_flow" {
  value = data.external.elastic_start_transform_vpc_flow.result
  depends_on = [data.external.elastic_start_transform_vpc_flow]
}

# -------------------------------------------------------------
#  Load Dashboards
# -------------------------------------------------------------

data "external" "elastic_upload_saved_objects" {
  query = {
	elastic_http_method = "POST"
    kibana_endpoint  = ec_deployment.elastic_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    so_file      		= "${path.module}/../dashboards/google_cloud_dashboards.ndjson"
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/kb_upload_saved_objects.sh" ]
  depends_on = [ec_deployment.elastic_deployment]
}

output "elastic_upload_saved_objects" {
  value = data.external.elastic_upload_saved_objects.result
  depends_on = [data.external.elastic_upload_saved_objects]
}