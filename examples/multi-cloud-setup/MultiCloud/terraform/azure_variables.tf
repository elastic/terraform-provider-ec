# -------------------------------------------------------------
# Elastic configuration
# -------------------------------------------------------------
variable "elastic_azure_region" {
  type = string
  default = "azure-westeurope"
}

variable "elastic_azure_deployment_name" {
  type = string
  default = "Azure Observe and Protect"
}

variable "elastic_azure_deployment_template_id" {
  type = string
  default = "azure-general-purpose"
}

# -------------------------------------------------------------
# AWS configuration
# -------------------------------------------------------------

variable "azure_region" {
  type = string
  default = "West Europe"
}

variable  "azure_subscription_id" {
 type = string   
}

variable  "azure_client_id" {
 type = string   
}

variable  "azure_client_secret" {
 type = string   
}

variable  "azure_tenant_id" {
 type = string   
}



