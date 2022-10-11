variable "deploy_aws" {
  type = bool
  default = true
}

variable "deploy_gc" {
  type = bool
  default = true
}

module "aws_environment" {
  source = "../../AWS/terraform"

  elastic_version = var.elastic_version
  elastic_region = var.elastic_aws_region
  elastic_deployment_name = var.elastic_aws_deployment_name
  elastic_deployment_template_id = var.elastic_aws_deployment_template_id
  aws_region = var.aws_region
  aws_access_key = var.aws_access_key
  aws_secret_key = var.aws_secret_key
  
  ###
  # Uncomment the following line to make the AWS cluster the All in One Cluster via CCS
  ###
  //elastic_remotes = [{id = module.gc_environment[0].elastic_cluster_id_google, alias = module.gc_environment[0].elastic_cluster_alias_google}]

  count  = (var.deploy_aws == true) ? 1 : 0
}

module "gc_environment" {
  source = "../../GoogleCloud/terraform"

  elastic_version = var.elastic_version
  elastic_region = var.elastic_gc_region
  elastic_deployment_name = var.elastic_gc_deployment_name
  elastic_deployment_template_id = var.elastic_gc_deployment_template_id

  google_cloud_project = var.google_cloud_project
  google_cloud_region = var.google_cloud_region
  google_cloud_service_account_path = var.google_cloud_service_account_path
  google_cloud_network = var.google_cloud_network
  
  ###
  # Uncomment the following line to make the Google Cloud cluster the All in One Cluster via CCS
  ###
  elastic_remotes = [{id = module.aws_environment[0].elastic_cluster_id_aws, alias = module.aws_environment[0].elastic_cluster_alias_aws}]

  count  = (var.deploy_gc == true) ? 1 : 0
}