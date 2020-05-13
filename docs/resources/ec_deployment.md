---
page_title: "Elastic Cloud: ec_deployment"
description: |-
  Provides an Elastic Cloud deployment resource. This allows deployments to be created, updated, and deleted.
---

# Resource: ec_deployment

Provides an Elastic Cloud deployment resource. This allows deployments to be created, updated, and deleted.

## Example Usage

```hcl
resource "ec_deployment" "my_deployment" {
  name = "my example deployment"

  version = "7.6.2"
  region  = "us-east-1"

  elasticsearch {
    deployment_template_id = "aws-io-optimized"

    topology {
      instance_configuration_id = "aws.data.highio.i3"
    }
  }

  kibana {
    deployment_template_id = "aws-io-optimized"

    topology {
      instance_configuration_id = "aws.kibana.r4"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* List
<!-- TODO -->

### Timeouts


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* list

## Import

Deployments can be imported using the `id`, e.g.

```
$ terraform import ec_deployment.search 320b7b540dfc967a7a649c18e2fce4ed
```
