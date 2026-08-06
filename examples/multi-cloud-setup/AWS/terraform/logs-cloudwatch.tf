# -------------------------------------------------------------
# Get all Log Groups
# -------------------------------------------------------------

data "aws_cloudwatch_log_groups" "all" {}

# -------------------------------------------------------------
# Data Collection
#  -- For Cloudwatch we use Elastic Agent to collect data from each log group
# -------------------------------------------------------------

# data "external" "elastic_add_cw_integrations" {
#   for_each=data.aws_cloudwatch_log_groups.all.arns
#   query = {
#     kibana_endpoint  = ec_deployment.elastic_deployment.kibana[0].https_endpoint
#     elastic_username  = ec_deployment.elastic_deployment.elasticsearch_username
#     elastic_password  = ec_deployment.elastic_deployment.elasticsearch_password
#     elastic_json_body = templatefile("${path.module}/../json_templates/aws_cw_integration.json", 
#     {
#     "name_suffix": each.key,
#     //"log_group_name": each.value.log_group_names,
#     "log_group_arn": each.value,
#     "policy_id": data.external.elastic_create_policy.result.id,
#     "access_key": var.aws_access_key,
#     "access_secret": var.aws_secret_key,
#     }
#     )
#   }
#   program = ["sh", "${path.module}/../../lib/elastic_api/kb_add_integration_to_policy.sh" ]
#   depends_on = [data.external.elastic_create_policy, data.aws_cloudwatch_log_groups.all]
# }

# -------------------------------------------------------------
# Data Collection
#  -- using the serverless forwarder
# -------------------------------------------------------------

