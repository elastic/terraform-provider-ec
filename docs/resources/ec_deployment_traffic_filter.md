---
page_title: "Elastic Cloud: ec_deployment_traffic_filter"
description: |-
  Provides an Elastic Cloud traffic filter resource, which allows traffic filter rules to be created, updated, and deleted. Traffic filter rules are used to limit inbound traffic to deployment resources.
---

# Resource: ec_deployment_traffic_filter

Provides an Elastic Cloud traffic filter resource, which allows traffic filter rules to be created, updated, and deleted. Traffic filter rules are used to limit inbound traffic to deployment resources.

## Example Usage

### IP type

```hcl
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  traffic_filter = [
    ec_deployment_traffic_filter.example.id
  ]

  # Use the deployment template defaults
  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}
}

resource "ec_deployment_traffic_filter" "example" {
  name   = "my traffic filter name"
  region = "us-east-1"
  type   = "ip"

  rule {
    source = "0.0.0.0/0"
  }
}
```

### Azure Private Link type

```hcl
locals {
  region = azure-australiaeast
}

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = local.region
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = local.region
  version                = data.ec_stack.latest.version
  deployment_template_id = "azure-io-optimized-v3"

  traffic_filter = [
    ec_deployment_traffic_filter.azure.id
  ]

  # Use the deployment template defaults
  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}
}

resource "ec_deployment_traffic_filter" "azure" {
  name   = "my traffic filter name"
  region = local.region
  type   = "azure_private_endpoint"

  rule {
    azure_endpoint_name = "my-azure-pl"
    azure_endpoint_guid = "78c64959-fd88-41cc-81ac-1cfcdb1ac32e"
  }
}

```

### GCP Private Service Connect type

```hcl
locals {
  region = asia-east1
}

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = local.region
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = local.region
  version                = data.ec_stack.latest.version
  deployment_template_id = "gcp-storage-optimized"

  traffic_filter = [
    ec_deployment_traffic_filter.gcp_psc.id
  ]

  # Use the deployment template defaults
  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}
}

resource "ec_deployment_traffic_filter" "gcp_psc" {
  name   = "my traffic filter name"
  region = local.region
  type   = "gcp_private_service_connect_endpoint"

  rule {
    source = "18446744072646845332"
  }
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the ruleset.
* `type` - (Required) Type of the ruleset.  It can be `"ip"`, `"vpce"`, `"azure_private_endpoint"`, or `"gcp_private_service_connect_endpoint"`.
* `region` - (Required) Filter region, the ruleset can only be attached to deployments in the specific region.
* `rule` (Required) Rule block, which can be specified multiple times for multiple rules.
* `include_by_default` - (Optional) To automatically include the ruleset in the new deployments. Defaults to `false`.
* `description` - (Optional) Description of the ruleset.

### Rules

The `rule` block supports the following configuration options:

* `source` - (Optional) traffic filter source: IP address, CIDR mask, or VPC endpoint ID, **only required** when the type is not `"azure_private_endpoint"`.
* `description` - (Optional) Description of this individual rule.
* `azure_endpoint_name` - (Optional) Azure endpoint name. Only applicable when the ruleset type is set to `"azure_private_endpoint"`.
* `azure_endpoint_guid` - (Optional) Azure endpoint GUID. Only applicable when the ruleset type is set to `"azure_private_endpoint"`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ruleset ID.

For any `rule` an `id` is exported. For example: `ec_deployment_traffic_filter.default.rule.0.id`.

## Import

You can import traffic filters using the `id`, for example:

```
$ terraform import ec_deployment_traffic_filter.name 320b7b540dfc967a7a649c18e2fce4ed
```
