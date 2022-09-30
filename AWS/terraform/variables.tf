# -------------------------------------------------------------
# Elastic configuration
# -------------------------------------------------------------
variable "elastic_version" {
  type = string
  default = "latest"
}

variable "elastic_region" {
  type = string
  default = "aws-eu-west-2"
}

variable "elastic_deployment_name" {
  type = string
  default = "AWS"
}

variable "elastic_deployment_template_id" {
  type = string
  default = "aws-general-purpose-arm-v5"
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
# AWS configuration
# -------------------------------------------------------------

variable "aws_region" {
  type = string
  default = "eu-west-1"
}

variable "aws_access_key" {
  type = string
}

variable "aws_secret_key" {
  type = string
}
