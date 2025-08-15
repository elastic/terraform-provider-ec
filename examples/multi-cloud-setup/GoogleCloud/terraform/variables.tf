# -------------------------------------------------------------
# Elastic configuration
# -------------------------------------------------------------
variable "elastic_version" {
  type = string
  default = "latest"
}

variable "elastic_region" {
  type = string
  default = "gcp-europe-west3"
}

variable "elastic_deployment_name" {
  type = string
  default = "Google Cloud Observe and Protect"
}

variable "elastic_deployment_template_id" {
  type = string
  default = "gcp-io-optimized-v2"
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
# GCP configuration
# -------------------------------------------------------------

variable "google_cloud_project" {
  type = string
  default = "elastic-pme-team"
}

variable "google_cloud_region" {
  type = string
  default = "europe-west3"
}

variable "google_cloud_service_account_path" {
  type = string
}

variable "google_cloud_network" {
  type = string
  default = "default"
}

# -------------------------------------------------------------
# PubSub configuration
# -------------------------------------------------------------

//Audit Logs
variable "google_pubsub_audit_topic" {
  type = string
  default = "elastic-audit-logs"
}

variable "google_pubsub_audit_filter" {
  type = string
  default = "protoPayload.@type=\"type.googleapis.com/google.cloud.audit.AuditLog\""
}

//Firewall Logs
variable "google_pubsub_firewall_topic" {
  type = string
  default = "elastic-firewall-logs"
}

variable "google_pubsub_firewall_filter" {
  type = string
  default = "logName:\"compute.googleapis.com%2Ffirewall\""
}

//VPC Flow Logs
variable "google_pubsub_vpcflow_topic" {
  type = string
  default = "elastic-vpcflow-logs"
}

variable "google_pubsub_vpcflow_filter" {
  type = string
  default = "log_id(\"compute.googleapis.com/vpc_flows\")"
}

//DNS Logs
variable "google_pubsub_dns_topic" {
  type = string
  default = "elastic-dns-logs"
}

variable "google_pubsub_dns_filter" {
  type = string
  default = "resource.type=\"dns_query\""
}

//Loadbalancer Logs
variable "google_pubsub_lb_topic" {
  type = string
  default = "elastic-lb-logs"
}

variable "google_pubsub_lb_filter" {
  type = string
  default = "resource.type=\"http_load_balancer\""
}

# -------------------------------------------------------------
# BigQuery configuration -- Not used at the moment
# -------------------------------------------------------------

variable "google_cloud_container_specs_path"  {
  type = string
  default = "gs://dataflow-templates/latest/flex/BigQuery_to_Elasticsearch"
}

variable "google_cloud_maxNumWorkers"  {
  type = number
  default = 5
}

variable "google_cloud_maxRetryAttempts" {
  type = string
  default = 1
}

variable "google_cloud_maxRetryDuration" {
  type = string
  default = 30
}