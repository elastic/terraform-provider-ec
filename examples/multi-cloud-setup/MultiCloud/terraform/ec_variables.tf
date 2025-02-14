# -------------------------------------------------------------
# Elastic configuration for every cluster
# -------------------------------------------------------------
variable "elastic_version" {
  type = string
  default = "latest"
}

variable "elastic_agent_vm_name" {
  type = string
  default = "elastic-agent"
}