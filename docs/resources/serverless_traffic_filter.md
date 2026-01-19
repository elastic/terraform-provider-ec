---
page_title: "Elastic Cloud: ec_serverless_traffic_filter Resource"
description: |-
  Provides an Elastic Cloud Serverless traffic filter resource, which allows traffic filter rules to be created, updated, and deleted for serverless projects.
---

# Resource: ec_serverless_traffic_filter

Provides an Elastic Cloud Serverless traffic filter resource, which allows traffic filter rules to be created, updated, and deleted. Traffic filters are used to limit inbound traffic to serverless projects (Elasticsearch, Observability, and Security).

## Example Usage

### IP based traffic filter (AWS)

```terraform
resource "ec_serverless_traffic_filter" "example" {
  name               = "my-traffic-filter"
  region             = "aws-us-east-1"
  type               = "ip"
  include_by_default = false
  
  rules = [
    {
      source      = "203.0.113.0/24"
      description = "Office network"
    },
    {
      source = "198.51.100.42/32"
    }
  ]
}

# Associate traffic filter with Elasticsearch project
resource "ec_elasticsearch_project" "example" {
  name       = "my-project"
  region_id  = "aws-us-east-1"
  
  traffic_filter_ids = [
    ec_serverless_traffic_filter.example.id
  ]
}
```

### IP based traffic filter (Azure)

```terraform
resource "ec_serverless_traffic_filter" "azure_example" {
  name               = "azure-office-filter"
  region             = "azure-eastus2"
  type               = "ip"
  include_by_default = false
  
  rules = [
    {
      source      = "10.0.0.0/16"
      description = "Azure VNet CIDR"
    }
  ]
}

resource "ec_elasticsearch_project" "azure_example" {
  name       = "my-azure-project"
  region_id  = "azure-eastus2"
  
  traffic_filter_ids = [
    ec_serverless_traffic_filter.azure_example.id
  ]
}
```

### IP based traffic filter (GCP)

```terraform
resource "ec_serverless_traffic_filter" "gcp_example" {
  name               = "gcp-office-filter"
  region             = "gcp-us-central1"
  type               = "ip"
  include_by_default = false
  
  rules = [
    {
      source      = "172.16.0.0/12"
      description = "GCP VPC CIDR"
    }
  ]
}

resource "ec_observability_project" "gcp_example" {
  name       = "my-gcp-project"
  region_id  = "gcp-us-central1"
  
  traffic_filter_ids = [
    ec_serverless_traffic_filter.gcp_example.id
  ]
}
```

### VPCE (AWS PrivateLink) traffic filter

```terraform
resource "ec_serverless_traffic_filter" "vpce" {
  name               = "my-vpce-filter"
  region             = "aws-us-east-1"
  type               = "vpce"
  include_by_default = false
  
  rules = [
    {
      source      = "vpce-0a1b2c3d4e5f6g7h8"
      description = "VPC endpoint from production VPC"
    }
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the traffic filter.
* `region` - (Required) Region where the traffic filter will be created (e.g., "aws-us-east-1").
* `type` - (Required) Type of traffic filter. Valid values are `ip` or `vpce`. Note: Azure and GCP endpoint types are not currently supported for serverless.
* `description` - (Optional) Description of the traffic filter.
* `include_by_default` - (Optional) Whether new projects in the region should automatically include this traffic filter. Defaults to `false`.
* `rules` - (Optional) List of traffic filter rules. Each rule supports:
  * `source` - (Required) The source to allow traffic from. For `ip` type, this should be a valid CIDR block (e.g., "192.168.1.0/24"). For `vpce` type, this should be a VPC endpoint ID (e.g., "vpce-0a1b2c3d4e5f6g7h8").
  * `description` - (Optional) Description of the rule.

## Attributes Reference

In addition to the arguments, the following attributes are exported:

* `id` - The unique identifier of the traffic filter.

## Import

Traffic filters can be imported using the `id`:

```shell
terraform import ec_serverless_traffic_filter.example tf-123456789
```
