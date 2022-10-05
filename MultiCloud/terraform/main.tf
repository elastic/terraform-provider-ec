terraform {
  required_version = ">= 1.0.2"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "0.4.1"
    }
  }
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

  elastic_remotes = [{id = module.aws_environment.elastic_cluster_id, alias = module.aws_environment.elastic_cluster_alias}]
}

