# -------------------------------------------------------------
#  Deploy Elastic Cloud
# -------------------------------------------------------------
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = var.elastic_gc_region
}

resource "ec_deployment" "elastic_gc_deployment" {
  name                    = var.elastic_gc_deployment_name
  region                  = var.elastic_gc_region
  version                 = var.elastic_version == "latest" ? data.ec_stack.latest.version : var.elastic_version
  deployment_template_id  = var.elastic_gc_deployment_template_id
  elasticsearch {
	autoscale = "true"
  }
  kibana {}
  integrations_server {}
}

output "elastic_endpoint" {
  value = ec_deployment.elastic_gc_deployment.elasticsearch[0].https_endpoint
}

output "elastic_password" {
  value = ec_deployment.elastic_gc_deployment.elasticsearch_password
  sensitive=true
}

output "elastic_cloud_id" {
  value = ec_deployment.elastic_gc_deployment.elasticsearch[0].cloud_id
}

output "elastic_username" {
  value = ec_deployment.elastic_gc_deployment.elasticsearch_username
}

# -------------------------------------------------------------
#  Load Policy
# -------------------------------------------------------------

data "external" "elastic_create_gcp_policy" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_gc_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_gc_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_gc_deployment.elasticsearch_password
    elastic_json_body = templatefile("../json_templates/default-policy.json", {"policy_name": "GC_${var.google_cloud_project}"})
  }
  program = ["sh", "../scripts/kb_create_agent_policy.sh" ]
  depends_on = [ec_deployment.elastic_gc_deployment]
}

output "elastic_create_gcp_policy" {
  value = data.external.elastic_create_gcp_policy.result
  depends_on = [data.external.elastic_create_gcp_policy]
}

data "external" "elastic_add_gcp_integration" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_gc_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_gc_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_gc_deployment.elasticsearch_password
    elastic_json_body = templatefile("../json_templates/gcp_integration.json", 
    {
    "policy_id": data.external.elastic_create_gcp_policy.result.id,
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
  program = ["sh", "../scripts/kb_add_integration_to_policy.sh" ]
  depends_on = [data.external.elastic_create_gcp_policy]
}

output "elastic_add_gcp_integration" {
  value = data.external.elastic_add_gcp_integration.result
  depends_on = [data.external.elastic_add_gcp_integration]
}

# -------------------------------------------------------------
#  Load Rules
# -------------------------------------------------------------

data "external" "elastic_load_rules" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_gc_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_gc_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_gc_deployment.elasticsearch_password
  }
  program = ["sh", "../scripts/kb_load_detection_rules.sh" ]
  depends_on = [ec_deployment.elastic_gc_deployment]
}

data "external" "elastic_enable_gcp_rules" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_gc_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_gc_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_gc_deployment.elasticsearch_password
    elastic_json_body = templatefile("../json_templates/es_gcp_rule_activation.json",{})
  }
  program = ["sh", "../scripts/kb_enable_detection_rules.sh" ]
  depends_on = [data.external.elastic_load_rules]
}

output "elastic_enable_gcp_rules" {
  value = data.external.elastic_enable_gcp_rules.result
  depends_on = [data.external.elastic_enable_gcp_rules]
}

# -------------------------------------------------------------
#  Create and Start transforms
# -------------------------------------------------------------

data "external" "elastic_gcp_create_transform_gcs" {
  query = {
    elastic_endpoint  = ec_deployment.elastic_gc_deployment.elasticsearch[0].https_endpoint
    elastic_username  = ec_deployment.elastic_gc_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_gc_deployment.elasticsearch_password
    transform_name    = "gcs-repo-transform"
    elastic_json_body = templatefile("../json_templates/es_gcp_repo_transform.json",{})
  }
  program = ["sh", "../scripts/es_create_transform.sh" ]
  depends_on = [ec_deployment.elastic_gc_deployment]
}

data "external" "elastic_gcp_start_transform_gcs" {
  query = {
    elastic_endpoint  = ec_deployment.elastic_gc_deployment.elasticsearch[0].https_endpoint
    elastic_username  = ec_deployment.elastic_gc_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_gc_deployment.elasticsearch_password
    transform_name    = "gcs-repo-transform"
  }
  program = ["sh", "../scripts/es_start_transform.sh" ]
  depends_on = [data.external.elastic_gcp_create_transform_gcs]
}

output "elastic_gcp_start_transform_gcs" {
  value = data.external.elastic_gcp_start_transform_gcs.result
  depends_on = [data.external.elastic_gcp_start_transform_gcs]
}

################################################################################

data "external" "elastic_gcp_create_transform_host_metrics" {
  query = {
    elastic_endpoint  = ec_deployment.elastic_gc_deployment.elasticsearch[0].https_endpoint
    elastic_username  = ec_deployment.elastic_gc_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_gc_deployment.elasticsearch_password
    transform_name    = "host-profile-transform"
    elastic_json_body = templatefile("../json_templates/es_gcp_host_transform.json",{})
  }
  program = ["sh", "../scripts/es_create_transform.sh" ]
  depends_on = [ec_deployment.elastic_gc_deployment]
}

data "external" "elastic_gcp_start_transform_host_metrics" {
  query = {
    elastic_endpoint  = ec_deployment.elastic_gc_deployment.elasticsearch[0].https_endpoint
    elastic_username  = ec_deployment.elastic_gc_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_gc_deployment.elasticsearch_password
    transform_name    = "host-profile-transform"
  }
  program = ["sh", "../scripts/es_start_transform.sh" ]
  depends_on = [data.external.elastic_gcp_create_transform_host_metrics]
}

output "elastic_gcp_start_transform_host_metrics" {
  value = data.external.elastic_gcp_start_transform_host_metrics.result
  depends_on = [data.external.elastic_gcp_start_transform_host_metrics]
}

################################################################################

data "external" "elastic_gcp_create_transform_vpc_flow" {
  query = {
    elastic_endpoint  = ec_deployment.elastic_gc_deployment.elasticsearch[0].https_endpoint
    elastic_username  = ec_deployment.elastic_gc_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_gc_deployment.elasticsearch_password
    transform_name    = "vpc_flow-transform"
    elastic_json_body = templatefile("../json_templates/es_gcp_vpc_flow_transform.json",{})
  }
  program = ["sh", "../scripts/es_create_transform.sh" ]
  depends_on = [ec_deployment.elastic_gc_deployment]
}

data "external" "elastic_gcp_start_transform_vpc_flow" {
  query = {
    elastic_endpoint  = ec_deployment.elastic_gc_deployment.elasticsearch[0].https_endpoint
    elastic_username  = ec_deployment.elastic_gc_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_gc_deployment.elasticsearch_password
    transform_name    = "vpc_flow-transform"
  }
  program = ["sh", "../scripts/es_start_transform.sh" ]
  depends_on = [data.external.elastic_gcp_create_transform_vpc_flow]
}

output "elastic_gcp_start_transform_vpc_flow" {
  value = data.external.elastic_gcp_start_transform_vpc_flow.result
  depends_on = [data.external.elastic_gcp_start_transform_vpc_flow]
}

# -------------------------------------------------------------
#  Load Dashboards
# -------------------------------------------------------------

data "external" "elastic_upload_gcp_saved_objects" {
  query = {
	elastic_http_method = "POST"
    kibana_endpoint  = ec_deployment.elastic_gc_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_gc_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_gc_deployment.elasticsearch_password
    so_file      		= "../dashboards/google_cloud_dashboards.ndjson"
  }
  program = ["sh", "../scripts/kb_upload_saved_objects.sh" ]
  depends_on = [ec_deployment.elastic_gc_deployment]
}

output "elastic_upload_gcp_saved_objects" {
  value = data.external.elastic_upload_gcp_saved_objects.result
  depends_on = [data.external.elastic_upload_gcp_saved_objects]
}