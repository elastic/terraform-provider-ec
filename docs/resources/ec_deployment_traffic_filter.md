---
page_title: "Elastic Cloud: ec_deployment_traffic_filter"
description: |-
  Provides an Elastic Cloud traffic filter resource, which allows traffic filter rules to be created, updated, and deleted. Traffic filter rules are used to limit inbound traffic to deployment resources.
---

# Resource: ec_deployment_traffic_filter

Provides an Elastic Cloud traffic filter resource, which allows traffic filter rules to be created, updated, and deleted. Traffic filter rules are used to limit inbound traffic to deployment resources.

## Example Usage

```hcl
resource "ec_deployment" "example_minimal" {
  region                 = "us-east-1"
  version                = "7.8.1"
  deployment_template_id = "aws-io-optimized-v2"

  traffic_filter = [
    ec_deployment_traffic_filter.example.id
  ]

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highio.i3"
    }
  }

  kibana {
    topology {
      instance_configuration_id = "aws.kibana.r5d"
    }
  }

  apm {
    topology {
      instance_configuration_id = "aws.apm.r5d"
    }
  }
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

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the ruleset.
* `type` - (Required) Type of the ruleset.  It can be `"ip"` or `"vpce"`.
* `region` - (Required) Filter region, the ruleset can only be attached to deployments in the specific region.
* `rule` (Required) Rule block, which can be specified multiple times for multiple rules.
* `include_by_default` - (Optional) To automatically include the ruleset in the new deployments. Defaults to `false`.
* `description` - (Optional) Description of the ruleset.

### Rules

The `rule` block supports the following configuration options:

* `source` - (Required) Source type, `"ip"` or `"vpce"`, from which the ruleset accepts traffic.
* `description` - (Optional) Description of this individual rule.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ruleset ID.

For any `rule` an `id` is exported. For example: `ec_deployment_traffic_filter.default.rule.0.id`.

## Import

You can import traffic filters using the `id`, for example:

```
$ terraform import ec_deployment_traffic_filter.name 320b7b540dfc967a7a649c18e2fce4ed
```
