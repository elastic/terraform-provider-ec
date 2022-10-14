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

output "elastic_cluster_id" {
  value = ec_deployment.elastic_deployment.id
}

output "elastic_cluster_alias" {
  value = ec_deployment.elastic_deployment.name
}

output "elastic_endpoint" {
  value = ec_deployment.elastic_deployment.elasticsearch[0].https_endpoint
}

output "elastic_password" {
  value = ec_deployment.elastic_deployment.elasticsearch_password
  sensitive=true
}

output "elastic_cloud_id" {
  value = ec_deployment.elastic_deployment.elasticsearch[0].cloud_id
}

output "elastic_username" {
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
    elastic_json_body = templatefile("${path.module}/../json_templates/default-policy.json", {"policy_name": "Azure"})
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/kb_create_agent_policy.sh" ]
  depends_on = [ec_deployment.elastic_deployment]
}

data "external" "elastic_add_metrics_integration" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    elastic_json_body = templatefile("${path.module}/../json_templates/azure-metrics-integration.json", 
    {
    "policy_id": data.external.elastic_create_policy.result.id,
    "client_id": var.azure_client_id,
    "client_secret": var.azure_client_secret,
    "tenant_id": var.azure_tenant_id,
    "subscription_id": var.azure_subscription_id
    }
    )
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/kb_add_integration_to_policy.sh" ]
  depends_on = [data.external.elastic_create_policy]
}

data "external" "elastic_add_logs_integration" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    elastic_json_body = templatefile("${path.module}/../json_templates/azure-logs-integration.json", 
    {
    "policy_id": data.external.elastic_create_policy.result.id,
    "eventhub": azurerm_eventhub.elastic.name,
    "connection_string": azurerm_eventhub_authorization_rule.elastic.primary_connection_string ,
    "storage_account": azurerm_storage_account.elastic.id ,
    "storage_account_key": azurerm_storage_account.elastic.primary_access_key
    }
    )
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/kb_add_integration_to_policy.sh" ]
  depends_on = [
    data.external.elastic_create_policy,
    azurerm_eventhub.elastic,
    azurerm_eventhub_authorization_rule.elastic,
    azurerm_storage_account.elastic
  ]
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

# data "external" "elastic_enable_rules" {
#   query = {
#     kibana_endpoint  = ec_deployment.elastic_deployment.kibana[0].https_endpoint
#     elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
#     elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
#     elastic_json_body = templatefile("${path.module}/../json_templates/es_rule_activation.json",{})
#   }
#   program = ["sh", "${path.module}/../../lib/elastic_api/kb_enable_detection_rules.sh" ]
#   depends_on = [data.external.elastic_load_rules]
# }


# -------------------------------------------------------------
#  Create and Start transforms
# -------------------------------------------------------------


# -------------------------------------------------------------
#  Load Dashboards
# -------------------------------------------------------------

