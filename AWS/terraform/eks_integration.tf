# -------------------------------------------------------------
# Load integration policy for Elastic Agent
# -------------------------------------------------------------

data "external" "elastic_create_k8s_policy" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    elastic_json_body = templatefile("${path.module}/../json_templates/default-policy.json", {"policy_name": "k8s"})
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/kb_create_agent_policy.sh" ]
  depends_on = [ec_deployment.elastic_deployment]
}

data "external" "elastic_add_k8s_integration" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    elastic_json_body = templatefile("${path.module}/../json_templates/k8s_integration.json", 
    {
    "policy_id": data.external.elastic_create_k8s_policy.result.id
    }
    )
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/kb_add_integration_to_policy.sh" ]
  depends_on = [data.external.elastic_create_k8s_policy]
}

data "external" "elastic_add_cspm_integration" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    elastic_json_body = templatefile("${path.module}/../json_templates/k8s_cspm_integration.json", 
    {
    "policy_id": data.external.elastic_create_k8s_policy.result.id,
    "access_key": var.aws_access_key,
    "access_secret": var.aws_secret_key,
    }
    )
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/kb_add_integration_to_policy.sh" ]
  depends_on = [data.external.elastic_add_k8s_integration]
}

data "external" "elastic_add_endpoint_integration" {
  query = {
    kibana_endpoint  = ec_deployment.elastic_deployment.kibana[0].https_endpoint
    elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
    elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
    elastic_json_body = templatefile("${path.module}/../json_templates/k8s_endpoint_integration.json", 
    {
    "policy_id": data.external.elastic_create_k8s_policy.result.id,
    }
    )
  }
  program = ["sh", "${path.module}/../../lib/elastic_api/kb_add_integration_to_policy.sh" ]
  depends_on = [data.external.elastic_add_cspm_integration]
}

output "elastic_add_endpoint_integration_template" {
  value = templatefile("${path.module}/../json_templates/k8s_endpoint_integration.json", 
    {
    "policy_id": data.external.elastic_create_k8s_policy.result.id,
    }
    )
}