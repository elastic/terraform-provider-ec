---
page_title: "Elastic Cloud: ec_deployment_traffic_filter"
description: |-
  Provides an Elastic Cloud traffic filter resource. Allowing traffic filter rules intended to limit inbound traffic to a deployment resources to be created, updated, and deleted.
---

# Resource: ec_deployment_traffic_filter

Provides an Elastic Cloud traffic filter resource. Allowing traffic filter rules intended to limit inbound traffic to a deployment resources to be created, updated, and deleted.

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

* `name` - (Required) name of the ruleset.
* `type` - (Required) type of the ruleset (`"ip"` or `"vpce"`).
* `region` - (Required) filter region, the ruleset can only be attached to deployments in the specific region.
* `rule` (Required) rule block, which can be specified multiple times for multiple rules.
* `include_by_default` - (Optional) Should the ruleset be automatically included in the new deployments (Defaults to `false`).
* `description` - (Optional) description of the ruleset.

### Rules

The `rule` supports the following:

* `source` - (Required) source (IP or VPC endpoint) which the ruleset will accept traffic from.
* `description` - (Optional) description to attach to this individual rule.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ruleset ID.

For any `rule` an `id` is exported.

e.g. `ec_deployment_traffic_filter.default.rule.0.id`

## Import

Traffic filters can be imported using the `id`, e.g.

```
$ terraform import ec_deployment_traffic_filter.name 320b7b540dfc967a7a649c18e2fce4ed
```
