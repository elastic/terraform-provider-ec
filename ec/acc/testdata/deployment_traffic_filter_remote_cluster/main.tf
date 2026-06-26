variable "name" {
  type = string
}

variable "region" {
  type = string
}

variable "remote_cluster_id" {
  type = string
}

variable "remote_cluster_org_id" {
  type = string
}

resource "ec_deployment_traffic_filter" "remote_cluster" {
  name   = var.name
  region = var.region
  type   = "remote_cluster"

  rule {
    remote_cluster_id     = var.remote_cluster_id
    remote_cluster_org_id = var.remote_cluster_org_id
  }
}
