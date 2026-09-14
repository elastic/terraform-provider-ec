# -------------------------------------------------------------
# Elastic configuration
# -------------------------------------------------------------
variable "elastic_version" {
  type = string
  default = "latest"
}

variable "elastic_region" {
  type = string
  default = "azure-westeurope"
}

variable "elastic_deployment_name" {
  type = string
  default = "Azure Observe and Protect"
}

variable "elastic_deployment_template_id" {
  type = string
  default = "azure-general-purpose"
}

variable "elastic_remotes" {
    type = list(
            object({
                id    = string
                alias = string
        })
    )
    default = []
}

variable "elastic_agent_vm_name" {
  type = string
  default = "elastic-agent"
}

# -------------------------------------------------------------
# Azure configuration
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