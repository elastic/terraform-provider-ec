# -------------------------------------------------------------
# Elastic configuration
# -------------------------------------------------------------
variable "elastic_version" {
  type = string
  default = "latest"
}

variable "elastic_region" {
  type = string
  default = "azure-europe-west3"
}

variable "elastic_deployment_name" {
  type = string
  default = "Azure Observe and Protect"
}

variable "elastic_deployment_template_id" {
  type = string
  default = "azure-io-optimized-v2"
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

# -------------------------------------------------------------
# Azure configuration
# -------------------------------------------------------------

